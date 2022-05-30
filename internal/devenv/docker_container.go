package devenv

import (
	"context"
	"fmt"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/strslice"
	"github.com/docker/docker/client"
	"github.com/recode-sh/agent/constants"
	"github.com/recode-sh/agent/internal/docker"
)

func EnsureDockerContainerRunning(
	dockerClient *client.Client,
) error {

	isContainerRunning, err := docker.IsContainerRunning(
		dockerClient,
		constants.DevEnvDockerContainerName,
	)

	if err != nil {
		return err
	}

	if isContainerRunning {
		return nil
	}

	dockerContainer, err := docker.LookupContainer(
		dockerClient,
		constants.DevEnvDockerContainerName,
	)

	if err != nil {
		return err
	}

	if dockerContainer != nil { // Container exists but is not running
		return dockerClient.ContainerStart(
			context.TODO(),
			dockerContainer.ID,
			types.ContainerStartOptions{},
		)
	}

	createdDockerContainer, err := dockerClient.ContainerCreate(
		context.TODO(),

		&container.Config{
			WorkingDir: constants.DevEnvWorkspaceDirPath,
			Image:      constants.DevEnvDockerImageName,
			User:       constants.DevEnvRecodeUserName,
			Entrypoint: strslice.StrSlice{
				constants.DevEnvDockerContainerEntrypointFilePath,
			},
			Cmd: constants.DevEnvDockerContainerStartCmd,
		},

		&container.HostConfig{
			AutoRemove:  false,
			Binds:       buildHostMounts(),
			NetworkMode: container.NetworkMode("host"),
			Privileged:  true,
			RestartPolicy: container.RestartPolicy{
				Name: "always",
			},
		},

		nil,

		nil,

		constants.DevEnvDockerContainerName,
	)

	if err != nil {
		return err
	}

	return dockerClient.ContainerStart(
		context.TODO(),
		createdDockerContainer.ID,
		types.ContainerStartOptions{},
	)
}

func buildHostMounts() []string {
	return []string{
		// Working dir

		fmt.Sprintf(
			"%s:%s",
			constants.DevEnvWorkspaceDirPath,
			constants.DevEnvWorkspaceDirPath,
		),

		fmt.Sprintf(
			"%s:%s",
			constants.DevEnvWorkspaceConfigDirPath,
			constants.DevEnvWorkspaceConfigDirPath,
		),

		/* Config files are mounted to /etc/
		   to let users overwrite them, if needed,
		   using config files in home dir. */

		// Git config

		fmt.Sprintf(
			"/home/%s/.gitconfig:/etc/gitconfig",
			constants.DevEnvRecodeUserName,
		),

		// SSH config

		fmt.Sprintf(
			"/home/%s/.ssh/config:/etc/ssh/ssh_config",
			constants.DevEnvRecodeUserName,
		),

		fmt.Sprintf(
			"/home/%s/.ssh/known_hosts:/etc/ssh/ssh_known_hosts",
			constants.DevEnvRecodeUserName,
		),

		// SSH GitHub keys

		fmt.Sprintf(
			"/home/%s/.ssh/recode_github:/home/%s/.ssh/recode_github",
			constants.DevEnvRecodeUserName,
			constants.DevEnvRecodeUserName,
		),

		fmt.Sprintf(
			"/home/%s/.ssh/recode_github.pub:/home/%s/.ssh/recode_github.pub",
			constants.DevEnvRecodeUserName,
			constants.DevEnvRecodeUserName,
		),

		// GnuPG GitHub keys

		fmt.Sprintf(
			"/home/%s/.gnupg/recode_github_gpg_public.pgp:/home/%s/.gnupg/recode_github_gpg_public.pgp",
			constants.DevEnvRecodeUserName,
			constants.DevEnvRecodeUserName,
		),

		fmt.Sprintf(
			"/home/%s/.gnupg/recode_github_gpg_private.pgp:/home/%s/.gnupg/recode_github_gpg_private.pgp",
			constants.DevEnvRecodeUserName,
			constants.DevEnvRecodeUserName,
		),

		// VSCode server

		// fmt.Sprintf(
		// 	"/home/%s/.vscode-server/:/home/%s/.vscode-server/",
		// 	constants.DevEnvRecodeUserName,
		// 	constants.DevEnvRecodeUserName,
		// ),

		// Docker daemon socket

		"/var/run/docker.sock:/var/run/docker.sock",
	}
}

func EnsureDockerContainerRemoved(dockerClient *client.Client) error {
	dockerContainer, err := docker.LookupContainer(
		dockerClient,
		constants.DevEnvDockerContainerName,
	)

	if err != nil {
		return err
	}

	if dockerContainer == nil {
		return nil
	}

	return dockerClient.ContainerRemove(
		context.TODO(),
		dockerContainer.ID,
		types.ContainerRemoveOptions{
			Force: true,
		},
	)
}
