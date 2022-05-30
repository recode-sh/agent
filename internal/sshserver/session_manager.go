package sshserver

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/creack/pty"
	"github.com/gliderlabs/ssh"
)

type UserCommandBuilder interface {
	Build(args ...string) *exec.Cmd
	BuildShell() *exec.Cmd
	BuildShellPTY() *exec.Cmd
}

type PTYWindowSizer interface {
	SetWindowSize(f *os.File, width, height int)
}

type SessionManager struct {
	userCommandBuilder UserCommandBuilder
	ptyManager         PTYWindowSizer
}

func NewSessionManager(
	userCommandBuilder UserCommandBuilder,
	ptyManager PTYWindowSizer,
) SessionManager {

	return SessionManager{
		userCommandBuilder: userCommandBuilder,
		ptyManager:         ptyManager,
	}
}

func (s SessionManager) ManageShell(sshSession ssh.Session) error {
	_, _, hasPTY := sshSession.Pty()

	if hasPTY {
		return errors.New("expected no PTY, got PTY")
	}

	shellCmd := s.userCommandBuilder.BuildShell()

	shellCmd.Stdin = sshSession
	shellCmd.Stdout = sshSession
	shellCmd.Stderr = sshSession

	err := shellCmd.Start()

	if err != nil {
		return err
	}

	return shellCmd.Wait()
}

func (s SessionManager) ManageShellPTY(sshSession ssh.Session) error {
	ptyReq, windowChan, hasPTY := sshSession.Pty()

	if !hasPTY {
		return errors.New("expected PTY, got no PTY")
	}

	shellCmd := s.userCommandBuilder.BuildShellPTY()

	shellCmd.Env = append(
		shellCmd.Env,
		fmt.Sprintf("TERM=%s", ptyReq.Term),
	)

	shellCmdPty, err := pty.Start(shellCmd)

	if err != nil {
		return err
	}

	go func() {
		for window := range windowChan {
			s.ptyManager.SetWindowSize(
				shellCmdPty,
				window.Width,
				window.Height,
			)
		}
	}()

	go func() {
		io.Copy(shellCmdPty, sshSession) // stdin
	}()

	io.Copy(sshSession, shellCmdPty) // stdout

	return shellCmd.Wait()
}

func (s SessionManager) ManageExec(sshSession ssh.Session) error {
	passedCmd := sshSession.Command()

	if len(passedCmd) == 0 {
		return errors.New("expected command, got nothing")
	}

	cmdToExec := s.userCommandBuilder.Build(passedCmd...)

	cmdToExec.Stdin = sshSession
	cmdToExec.Stdout = sshSession
	cmdToExec.Stderr = sshSession

	err := cmdToExec.Start()

	if err != nil {
		return err
	}

	// We use ".Process.Wait()" here given that
	// ".Wait()" will wait indefinitly for "Stdin"
	// (the SSH channel) to close before returning.
	cmdState, err := cmdToExec.Process.Wait()

	if err != nil {
		return err
	}

	cmdExitCode := cmdState.ExitCode()

	if cmdExitCode != 0 {
		return fmt.Errorf(
			"the command \"%s\" has returned a non-zero (%d) exit code",
			passedCmd,
			cmdExitCode,
		)
	}

	return nil
}
