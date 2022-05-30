package sshserver

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/gliderlabs/ssh"
	"github.com/recode-sh/agent/constants"
	"github.com/recode-sh/agent/internal/devenv"
	"github.com/recode-sh/agent/internal/docker"
)

func (s SessionManager) ManageShellInDevEnv(sshSession ssh.Session) error {
	_, _, hasPTY := sshSession.Pty()

	if hasPTY {
		return errors.New("expected no PTY, got PTY")
	}

	dockerClient, err := docker.NewDefaultClient()

	if err != nil {
		return err
	}

	vscodeWorkspaceConfig, err := devenv.LoadVSCodeWorkspaceConfig(
		constants.DevEnvVSCodeWorkspaceConfigFilePath,
	)

	if err != nil {
		return err
	}

	// err = devenv.EnsureDockerContainerRunning(dockerClient)

	// if err != nil {
	// 	return err
	// }

	exec, err := dockerClient.ContainerExecCreate(
		context.TODO(),
		constants.DevEnvDockerContainerName,
		types.ExecConfig{
			AttachStdin:  true,
			AttachStdout: true,
			AttachStderr: true,
			Detach:       false,
			Tty:          false,
			Cmd:          []string{"/bin/bash"},
			Env:          []string{},
			WorkingDir:   constants.DevEnvWorkspaceDirPath,
			User:         constants.DevEnvRecodeUserName,
			Privileged:   true,
		},
	)

	if err != nil {
		return err
	}

	stream, err := dockerClient.ContainerExecAttach(
		context.TODO(),
		exec.ID,
		types.ExecStartCheck{},
	)

	if err != nil {
		return err
	}

	defer stream.Close()

	stdinChan := make(chan error, 1)

	go func() {
		stdin := bufio.NewReader(sshSession)

		vscodeExtensionsToInstall := vscodeWorkspaceConfig.Extensions.Recommendations
		codeServerStartFlagRegExp := regexp.MustCompile(`--start-server`)

		for {
			shellLineSent, err := stdin.ReadString('\n')

			if err != nil && errors.Is(err, io.EOF) {
				stdinChan <- nil
				break
			}

			if err != nil {
				stdinChan <- err
				break
			}

			/* Holy shit, forgive me for this dirty hack
			   but I haven't found a way of forcing VSCode
				 to start with specific extensions installed.

				 The following code updates the "code-server start" command
				 sent by VSCode during the SSH session.
			*/
			if len(vscodeExtensionsToInstall) > 0 {

				if codeServerStartFlagRegExp.MatchString(shellLineSent) {

					shellLineSent = codeServerStartFlagRegExp.ReplaceAllString(
						shellLineSent,
						"--start-server --install-extension "+
							strings.Join(vscodeExtensionsToInstall, " --install-extension "),
					)
				}
			}

			_, err = stream.Conn.Write([]byte(shellLineSent))

			if err != nil {
				stdinChan <- err
				break
			}
		}

		// _, err := io.Copy(stream.Conn, sshSession)

		// stdinChan <- err
	}()

	stdoutChan := make(chan error, 1)

	go func() {
		_, err := stdcopy.StdCopy(
			sshSession,
			sshSession.Stderr(),
			stream.Reader,
		)

		stdoutChan <- err
	}()

	select {
	case stdoutErr := <-stdoutChan:
		return stdoutErr
	case stdinErr := <-stdinChan:
		return stdinErr
	}
}

func (s SessionManager) ManageShellPTYInDevEnv(sshSession ssh.Session) error {
	ptyReq, windowChan, hasPTY := sshSession.Pty()

	if !hasPTY {
		return errors.New("expected PTY, got no PTY")
	}

	dockerClient, err := docker.NewDefaultClient()

	if err != nil {
		return err
	}

	// err = devenv.EnsureDockerContainerRunning(dockerClient)

	// if err != nil {
	// 	return err
	// }

	exec, err := dockerClient.ContainerExecCreate(
		context.TODO(),
		constants.DevEnvDockerContainerName,
		types.ExecConfig{
			AttachStdin:  true,
			AttachStdout: true,
			AttachStderr: true,
			Detach:       false,
			Tty:          true,
			Cmd: []string{
				"/bin/bash",
				"-c",
				fmt.Sprintf(
					// Display Ubuntu motd and run default shell for user
					"for i in /etc/update-motd.d/*; do $i; done && $(getent passwd %s | cut -d ':' -f 7)",
					constants.DevEnvRecodeUserName,
				),
			},
			Env: []string{
				fmt.Sprintf("TERM=%s", ptyReq.Term),
			},
			WorkingDir: constants.DevEnvWorkspaceDirPath,
			User:       constants.DevEnvRecodeUserName,
			Privileged: true,
		},
	)

	if err != nil {
		return err
	}

	stream, err := dockerClient.ContainerExecAttach(
		context.TODO(),
		exec.ID,
		types.ExecStartCheck{
			Detach: false,
			Tty:    true,
		},
	)

	if err != nil {
		return err
	}

	defer stream.Close()

	resizeChan := make(chan error, 1)

	go func() {
		for window := range windowChan {
			err := dockerClient.ContainerExecResize(
				context.TODO(),
				exec.ID,
				types.ResizeOptions{
					Height: uint(window.Height),
					Width:  uint(window.Width),
				},
			)

			if err != nil {
				resizeChan <- err
				break
			}
		}
	}()

	stdinChan := make(chan error, 1)

	go func() {
		_, err := io.Copy(stream.Conn, sshSession)

		stdinChan <- err
	}()

	stdoutChan := make(chan error, 1)

	go func() {
		_, err := io.Copy(
			sshSession,
			stream.Reader,
		)

		stdoutChan <- err
	}()

	select {
	case resizeErr := <-resizeChan:
		return resizeErr
	case stdoutErr := <-stdoutChan:
		return stdoutErr
	case stdinErr := <-stdinChan:
		return stdinErr
	}
}

func (s SessionManager) ManageExecInDevEnv(sshSession ssh.Session) error {
	passedCmd := sshSession.Command()

	if len(passedCmd) == 0 {
		return errors.New("expected command, got nothing")
	}

	dockerClient, err := docker.NewDefaultClient()

	if err != nil {
		return err
	}

	// err = devenv.EnsureDockerContainerRunning(dockerClient)

	// if err != nil {
	// 	return err
	// }

	exec, err := dockerClient.ContainerExecCreate(
		context.TODO(),
		constants.DevEnvDockerContainerName,
		types.ExecConfig{
			AttachStdin:  true,
			AttachStdout: true,
			AttachStderr: true,
			Detach:       false,
			Tty:          false,
			Cmd:          passedCmd,
			Env:          []string{},
			WorkingDir:   constants.DevEnvWorkspaceDirPath,
			User:         constants.DevEnvRecodeUserName,
			Privileged:   true,
		},
	)

	if err != nil {
		return err
	}

	stream, err := dockerClient.ContainerExecAttach(
		context.TODO(),
		exec.ID,
		types.ExecStartCheck{},
	)

	if err != nil {
		return err
	}

	defer stream.Close()

	stdinChan := make(chan error, 1)

	go func() {
		_, err := io.Copy(stream.Conn, sshSession)

		stdinChan <- err
	}()

	stdoutChan := make(chan error, 1)

	go func() {
		_, err := stdcopy.StdCopy(
			sshSession,
			sshSession.Stderr(),
			stream.Reader,
		)

		stdoutChan <- err
	}()

	select {
	case stdoutErr := <-stdoutChan:
		if stdoutErr != nil {
			return stdoutErr
		}
	case stdinErr := <-stdinChan:
		if stdinErr != nil {
			return stdinErr
		}
	}

	containerInspect, err := dockerClient.ContainerExecInspect(
		context.TODO(),
		exec.ID,
	)

	if err != nil {
		return err
	}

	if containerInspect.ExitCode != 0 {
		return fmt.Errorf(
			"the command \"%s\" has returned a non-zero (%d) exit code",
			passedCmd,
			containerInspect.ExitCode,
		)
	}

	return nil
}
