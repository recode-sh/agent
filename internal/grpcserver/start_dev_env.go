package grpcserver

import (
	"os"

	"github.com/recode-sh/agent/constants"
	"github.com/recode-sh/agent/internal/devenv"
	"github.com/recode-sh/agent/internal/docker"
	"github.com/recode-sh/agent/proto"
)

func (s *agentServer) BuildAndStartDevEnv(
	req *proto.BuildAndStartDevEnvRequest,
	stream proto.Agent_BuildAndStartDevEnvServer,
) error {

	dockerClient, err := docker.NewDefaultClient()

	if err != nil {
		return err
	}

	// The method "BuildAndStartDevEnv" may be run multiple times
	// so we need to ensure idempotency
	err = devenv.EnsureDockerContainerRemoved(dockerClient)

	if err != nil {
		return err
	}

	workspaceConfig, err := devenv.LoadWorkspaceConfig(
		constants.DevEnvWorkspaceConfigFilePath,
	)

	if err != nil {
		return err
	}

	preparedWorkspaceMetadata, err := devenv.PrepareWorkspace(
		req.UserConfigRepoOwner,
		req.UserConfigRepoName,
		req.DevEnvRepoOwner,
		req.DevEnvRepoName,
		workspaceConfig,
	)

	if err != nil {
		return err
	}

	defer os.RemoveAll(preparedWorkspaceMetadata.TmpUserConfigRepoDirPath)
	defer os.RemoveAll(preparedWorkspaceMetadata.TmpDevEnvRepoDirPath)

	err = devenv.Build(
		dockerClient,
		stream,
		req.UserConfigRepoOwner,
		req.UserConfigRepoName,
		req.DevEnvRepoOwner,
		req.DevEnvRepoName,
		preparedWorkspaceMetadata,
	)

	if err != nil {
		return err
	}

	err = devenv.EnsureDockerContainerRunning(dockerClient)

	if err != nil {
		return err
	}

	return devenv.RunWorkspaceHooks(
		dockerClient,
		stream,
		workspaceConfig,
	)
}
