package devenv

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"runtime"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/archive"
	"github.com/recode-sh/agent/constants"
	"github.com/recode-sh/agent/internal/docker"
	"github.com/recode-sh/agent/proto"
	"github.com/recode-sh/recode/entities"
)

func Build(
	dockerClient *client.Client,
	stream proto.Agent_BuildAndStartDevEnvServer,
	userConfigRepoOwner string,
	userConfigRepoName string,
	repoOwner string,
	repoName string,
	preparedWorkspaceMetadata *PreparedWorkspaceMetadata,
) error {

	defer removeDanglingDockerImages(dockerClient)

	err := stream.Send(&proto.BuildAndStartDevEnvReply{
		LogLineHeader: fmt.Sprintf(
			"Building %s/%s/%s",
			userConfigRepoOwner,
			userConfigRepoName,
			entities.DevEnvUserConfigDockerfileFileName,
		),
	})

	if err != nil {
		return err
	}

	err = ensureUserConfigDockerfileDerivesFromBaseDevEnv(
		filepath.Join(
			preparedWorkspaceMetadata.TmpUserConfigRepoDirPath,
			entities.DevEnvUserConfigDockerfileFileName,
		),
	)

	if err != nil {
		return err
	}

	includeRecodeBuildArgs := true // User base derives from "recodesh/base-dev-env"
	dockerBuildArgs, err := resolveDockerBuildArgs(
		map[string]string{},
		includeRecodeBuildArgs,
	)

	if err != nil {
		return err
	}

	dockerBuildContext := preparedWorkspaceMetadata.TmpUserConfigRepoDirPath
	userConfigDockerfileIsFinalImage := !preparedWorkspaceMetadata.DevEnvRepoHasDockerfile

	err = buildDockerImage(
		dockerClient,
		stream,
		dockerBuildContext,
		dockerBuildArgs,
		entities.DevEnvUserConfigDockerfileFileName,
		userConfigDockerfileIsFinalImage,
	)

	if err != nil {
		return err
	}

	if userConfigDockerfileIsFinalImage {
		return nil
	}

	err = stream.Send(&proto.BuildAndStartDevEnvReply{
		LogLineHeader: fmt.Sprintf(
			"Building %s/%s/%s/%s",
			repoOwner,
			repoName,
			entities.DevEnvRepositoryConfigDirectory,
			entities.DevEnvRepositoryDockerfileFileName,
		),
	})

	if err != nil {
		return err
	}

	err = ensureRepoDockerfileDerivesFromUserBase(
		preparedWorkspaceMetadata.TmpDevEnvRepoDockerfilePath,
	)

	if err != nil {
		return err
	}

	dockerBuildContext = preparedWorkspaceMetadata.TmpDevEnvRepoConfigDirPath

	includeRecodeBuildArgs = false
	dockerBuildArgs, err = resolveDockerBuildArgs(
		map[string]string{},
		includeRecodeBuildArgs,
	)

	if err != nil {
		return err
	}

	isFinalImage := true

	return buildDockerImage(
		dockerClient,
		stream,
		dockerBuildContext,
		dockerBuildArgs,
		entities.DevEnvRepositoryDockerfileFileName,
		isFinalImage,
	)
}

func buildDockerImage(
	dockerClient *client.Client,
	stream proto.Agent_BuildAndStartDevEnvServer,
	dockerBuildContext string,
	dockerBuildArgs map[string]*string,
	dockerfilePath string,
	isFinalImage bool,
) error {

	buildContextAsTAR, err := getDockerBuildContextAsTAR(
		dockerBuildContext,
	)

	if err != nil {
		return err
	}

	imageTag := entities.DevEnvUserConfigDockerfileImageName
	if isFinalImage {
		imageTag = constants.DevEnvDockerImageName
	}

	buildImageResp, err := dockerClient.ImageBuild(
		context.TODO(),
		buildContextAsTAR,
		types.ImageBuildOptions{
			Dockerfile: dockerfilePath,
			Tags:       []string{imageTag},
			Remove:     true,
			BuildArgs:  dockerBuildArgs,
		},
	)

	if err != nil {
		return err
	}

	defer buildImageResp.Body.Close()

	return docker.HandleBuildOutput(
		buildImageResp.Body,
		func(logLine string) error {
			return stream.Send(&proto.BuildAndStartDevEnvReply{
				LogLine: logLine,
			})
		},
	)
}

func getDockerBuildContextAsTAR(buildContext string) (io.Reader, error) {
	return archive.TarWithOptions(
		buildContext,
		&archive.TarOptions{},
	)
}

func resolveDockerBuildArgs(
	dockerfileArgs map[string]string,
	includeRecodeArgs bool,
) (map[string]*string, error) {

	buildArgs := map[string]*string{}

	for dockerfileArgName, dockerfileArgVal := range dockerfileArgs {
		dockerfileArgValCopy := dockerfileArgVal
		// If you take the address of "dockerfileArgVal" directly
		// all build args will point to the same address
		// that will point to the last argument visited.
		buildArgs[dockerfileArgName] = &dockerfileArgValCopy
	}

	if includeRecodeArgs {
		recodeUser, dockerGroup, err := lookupRecodeUserAndDockerGroup()

		if err != nil {
			return nil, err
		}

		buildArgs["RECODE_USER_ID"] = &recodeUser.Uid
		buildArgs["RECODE_USER_GROUP_ID"] = &recodeUser.Gid
		buildArgs["RECODE_DOCKER_GROUP_ID"] = &dockerGroup.Gid
	}

	instanceArch := runtime.GOARCH
	instanceOS := runtime.GOOS

	buildArgs["RECODE_INSTANCE_ARCH"] = &instanceArch
	buildArgs["RECODE_INSTANCE_OS"] = &instanceOS

	return buildArgs, nil
}

func removeDanglingDockerImages(
	dockerClient *client.Client,
) error {

	_, err := dockerClient.ImagesPrune(
		context.TODO(),
		filters.Args{},
	)

	return err
}
