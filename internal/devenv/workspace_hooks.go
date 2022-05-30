package devenv

import (
	"context"
	"fmt"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/recode-sh/agent/constants"
	"github.com/recode-sh/agent/proto"
	"github.com/recode-sh/recode/entities"
)

type WorkspaceConfigRepositoryHook struct {
	ScriptFilePath       string `json:"script_file_path"`
	ScriptWorkingDirPath string `json:"script_working_dir_path"`
}

func RunWorkspaceHooks(
	dockerClient *client.Client,
	stream proto.Agent_BuildAndStartDevEnvServer,
	workspaceConfig *WorkspaceConfig,
) error {

	for _, repo := range workspaceConfig.Repositories {

		if len(repo.Hooks) == 0 {
			continue
		}

		initHook := repo.Hooks[0]

		err := stream.Send(&proto.BuildAndStartDevEnvReply{
			LogLineHeader: fmt.Sprintf(
				"Running %s/%s/%s/%s/%s",
				repo.Owner,
				repo.Name,
				entities.DevEnvRepositoryConfigDirectory,
				entities.DevEnvRepositoryConfigHooksDirectory,
				entities.DevEnvRepositoryInitHookFileName,
			),
		})

		if err != nil {
			return err
		}

		exec, err := dockerClient.ContainerExecCreate(
			context.TODO(),
			constants.DevEnvDockerContainerName,
			types.ExecConfig{
				AttachStdin:  false,
				AttachStdout: true,
				AttachStderr: true,
				Detach:       false,
				Tty:          false,
				Cmd: []string{
					initHook.ScriptFilePath,
				},
				WorkingDir: initHook.ScriptWorkingDirPath,
				User:       constants.DevEnvRecodeUserName,
				Privileged: true,
			},
		)

		if err != nil {
			return err
		}

		containerStream, err := dockerClient.ContainerExecAttach(
			context.TODO(),
			exec.ID,
			types.ExecStartCheck{},
		)

		if err != nil {
			return err
		}

		defer containerStream.Close()

		grpcServerStreamWriter := NewGRPCBuildAndStartDevEnvStreamWriter(stream)

		_, err = stdcopy.StdCopy(
			grpcServerStreamWriter,
			grpcServerStreamWriter,
			containerStream.Reader,
		)

		if err != nil {
			return err
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
				"error while running \"init hook\" for \"%s/%s\". Exit status code %d",
				repo.Owner,
				repo.Name,
				containerInspect.ExitCode,
			)
		}
	}

	return nil
}

func installHookInWorkspaceConfigDir(hookFilePath string) (string, error) {
	hookFileContent, err := os.ReadFile(hookFilePath)

	if err != nil {
		return "", err
	}

	// Ensure that the hooks directory exists given that it
	// is not created during instance init
	err = os.MkdirAll(
		constants.DevEnvWorkspaceConfigHooksDirPath,
		os.FileMode(0755),
	)

	if err != nil {
		return "", err
	}

	hookTmpFile, err := os.CreateTemp(
		constants.DevEnvWorkspaceConfigHooksDirPath,
		"recode_workspace_hook_*",
	)

	if err != nil {
		return "", err
	}

	_, err = hookTmpFile.Write(hookFileContent)

	if err != nil {
		return "", err
	}

	err = hookTmpFile.Close()

	if err != nil {
		return "", err
	}

	err = os.Chmod(
		hookTmpFile.Name(),
		os.FileMode(0744),
	)

	if err != nil {
		return "", err
	}

	return hookTmpFile.Name(), nil
}
