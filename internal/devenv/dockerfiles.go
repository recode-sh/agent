package devenv

import (
	"fmt"
	"strings"

	"github.com/recode-sh/agent/internal/docker"
	"github.com/recode-sh/recode/entities"
)

func ensureUserConfigDockerfileDerivesFromBaseDevEnv(
	userConfigDockerfileFilePath string,
) error {

	userConfigDockerfileRootImage, err := docker.LookupDockerfileBaseImage(
		userConfigDockerfileFilePath,
	)

	if err != nil {
		return err
	}

	if !strings.HasPrefix(
		userConfigDockerfileRootImage,
		entities.DevEnvUserConfigDockerfileRootImage,
	) {

		return fmt.Errorf(
			"\"%s\" must derive from \"%s\"",
			entities.DevEnvUserConfigDockerfileFileName,
			entities.DevEnvUserConfigDockerfileRootImage,
		)
	}

	return nil
}

func ensureRepoDockerfileDerivesFromUserBase(
	workspaceDockerfileFilePath string,
) error {

	workspaceDockerfileRootImage, err := docker.LookupDockerfileBaseImage(
		workspaceDockerfileFilePath,
	)

	if err != nil {
		return err
	}

	if !strings.HasPrefix(
		workspaceDockerfileRootImage,
		entities.DevEnvUserConfigDockerfileImageName,
	) {

		return fmt.Errorf(
			"\"%s\" must derive from \"%s\"",
			entities.DevEnvRepositoryDockerfileFileName,
			entities.DevEnvUserConfigDockerfileImageName,
		)
	}

	return nil
}

func lookupVSCodeExtensionsInDockerfileLabels(
	dockerfileFilePath string,
) ([]string, error) {

	vscodeExtLabelValue, err := docker.LookupDockerfileLabelValue(
		dockerfileFilePath,
		entities.DevEnvDockerfilesVSCodeExtLabelKey,
	)

	if err != nil {
		return nil, err
	}

	if len(vscodeExtLabelValue) == 0 {
		return []string{}, nil
	}

	vscodeExtSepRegExp := entities.DevEnvDockerfilesVSCodeExtLabelSepRegExp

	// -1 to return all matches
	vscodeExtensions := vscodeExtSepRegExp.Split(
		vscodeExtLabelValue,
		-1,
	)

	return vscodeExtensions, nil
}

func lookupRepositoriesInDockerfileLabels(
	dockerfileFilePath string,
) ([]string, error) {

	reposLabelValue, err := docker.LookupDockerfileLabelValue(
		dockerfileFilePath,
		entities.DevEnvDockerfilesReposLabelKey,
	)

	if err != nil {
		return nil, err
	}

	if len(reposLabelValue) == 0 {
		return []string{}, nil
	}

	reposSepRegExp := entities.DevEnvDockerfilesReposLabelSepRegExp

	// -1 to return all matches
	repos := reposSepRegExp.Split(
		reposLabelValue,
		-1,
	)

	return repos, nil
}
