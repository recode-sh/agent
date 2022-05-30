package devenv

import (
	"os"
	"path/filepath"

	"github.com/recode-sh/agent/constants"
	"github.com/recode-sh/agent/internal/system"
	"github.com/recode-sh/recode/entities"
	"github.com/recode-sh/recode/github"
)

type PreparedWorkspaceMetadata struct {
	TmpUserConfigRepoDirPath    string
	TmpDevEnvRepoDirPath        string
	TmpDevEnvRepoConfigDirPath  string
	TmpDevEnvRepoDockerfilePath string
	DevEnvRepoHasDockerfile     bool
}

func PrepareWorkspace(
	userConfigRepoOwner string,
	userConfigRepoName string,
	devEnvRepoOwner string,
	devEnvRepoName string,
	workspaceConfig *WorkspaceConfig,
) (*PreparedWorkspaceMetadata, error) {

	preparedWorkspaceMetadata := &PreparedWorkspaceMetadata{}
	vscodeWorkspaceConfig := buildInitialVSCodeWorkspaceConfig()

	err := prepareUserConfigRepo(
		userConfigRepoOwner,
		userConfigRepoName,
		preparedWorkspaceMetadata,
		&vscodeWorkspaceConfig,
	)

	if err != nil {
		return nil, err
	}

	reposToCloneInWorkspace, err := prepareDevEnvRepo(
		devEnvRepoOwner,
		devEnvRepoName,
		preparedWorkspaceMetadata,
		&vscodeWorkspaceConfig,
	)

	if err != nil {
		return nil, err
	}

	filesManager := system.NewFileManager()

	// The method "PrepareWorkspace" could
	// be called multiple times in case of error
	// so we need to make sure that our code is idempotent
	err = filesManager.RemoveDirContent(
		constants.DevEnvWorkspaceDirPath,
	)

	if err != nil {
		return nil, err
	}

	// Same than previous comment
	err = filesManager.RemoveDirContent(
		constants.DevEnvWorkspaceConfigDirPath,
	)

	if err != nil {
		return nil, err
	}

	for _, repoToCloneInWorkspace := range reposToCloneInWorkspace {
		err = addRepoToWorkspace(
			devEnvRepoOwner,
			devEnvRepoName,
			repoToCloneInWorkspace,
			workspaceConfig,
			&vscodeWorkspaceConfig,
			preparedWorkspaceMetadata,
		)

		if err != nil {
			return nil, err
		}
	}

	err = saveVSCodeWorkspaceConfigAsFile(
		constants.DevEnvVSCodeWorkspaceConfigFilePath,
		vscodeWorkspaceConfig,
	)

	if err != nil {
		return nil, err
	}

	err = SaveWorkspaceConfigAsFile(
		constants.DevEnvWorkspaceConfigFilePath,
		workspaceConfig,
	)

	if err != nil {
		return nil, err
	}

	return preparedWorkspaceMetadata, nil
}

func prepareUserConfigRepo(
	userConfigRepoOwner string,
	userConfigRepoName string,
	preparedWorkspaceMetadata *PreparedWorkspaceMetadata,
	vscodeWorkspaceConfig *VSCodeWorkspaceConfig,
) error {

	tmpUserConfigRepoDirPath, err := createTmpDirForCloningUserConfigRepo()

	if err != nil {
		return err
	}

	preparedWorkspaceMetadata.TmpUserConfigRepoDirPath = tmpUserConfigRepoDirPath

	err = cloneGitHubRepo(
		userConfigRepoOwner,
		userConfigRepoName,
		tmpUserConfigRepoDirPath,
	)

	if err != nil {
		return err
	}

	userConfigVSCodeExtensions, err := lookupVSCodeExtensionsInDockerfileLabels(
		filepath.Join(
			tmpUserConfigRepoDirPath,
			entities.DevEnvUserConfigDockerfileFileName,
		),
	)

	if err != nil {
		return err
	}

	vscodeWorkspaceConfig.Extensions.Recommendations = mergeVSCodeExtensionsRecos(
		vscodeWorkspaceConfig.Extensions.Recommendations,
		userConfigVSCodeExtensions,
	)

	return nil
}

func prepareDevEnvRepo(
	devEnvRepoOwner string,
	devEnvRepoName string,
	preparedWorkspaceMetadata *PreparedWorkspaceMetadata,
	vscodeWorkspaceConfig *VSCodeWorkspaceConfig,
) ([]string, error) {

	tmpDevEnvRepoDirPath, err := createTmpDirForCloningDevEnvRepo()

	if err != nil {
		return nil, err
	}

	preparedWorkspaceMetadata.TmpDevEnvRepoDirPath = tmpDevEnvRepoDirPath

	err = cloneGitHubRepo(
		devEnvRepoOwner,
		devEnvRepoName,
		tmpDevEnvRepoDirPath,
	)

	if err != nil {
		return nil, err
	}

	devEnvRepoConfigDirPath := filepath.Join(
		tmpDevEnvRepoDirPath,
		entities.DevEnvRepositoryConfigDirectory,
	)

	preparedWorkspaceMetadata.TmpDevEnvRepoConfigDirPath = devEnvRepoConfigDirPath

	devEnvRepoDockerfilePath := filepath.Join(
		devEnvRepoConfigDirPath,
		entities.DevEnvRepositoryDockerfileFileName,
	)

	filesManager := system.NewFileManager()

	devEnvRepoHasDockerfile, err := filesManager.DoesFileExist(
		devEnvRepoDockerfilePath,
	)

	if err != nil {
		return nil, err
	}

	reposToCloneInWorkspace := []string{
		devEnvRepoOwner + "/" + devEnvRepoName,
	}

	if devEnvRepoHasDockerfile {
		preparedWorkspaceMetadata.TmpDevEnvRepoDockerfilePath = devEnvRepoDockerfilePath
		preparedWorkspaceMetadata.DevEnvRepoHasDockerfile = true

		devEnvRepoVSCodeExtensions, err := lookupVSCodeExtensionsInDockerfileLabels(
			devEnvRepoDockerfilePath,
		)

		if err != nil {
			return nil, err
		}

		vscodeWorkspaceConfig.Extensions.Recommendations = mergeVSCodeExtensionsRecos(
			vscodeWorkspaceConfig.Extensions.Recommendations,
			devEnvRepoVSCodeExtensions,
		)

		devEnvRepos, err := lookupRepositoriesInDockerfileLabels(
			devEnvRepoDockerfilePath,
		)

		if err != nil {
			return nil, err
		}

		if len(devEnvRepos) > 0 {
			reposToCloneInWorkspace = devEnvRepos
		}
	}

	return reposToCloneInWorkspace, nil
}

func addRepoToWorkspace(
	devEnvRepoOwner string,
	devEnvRepoName string,
	repoName string,
	workspaceConfig *WorkspaceConfig,
	vscodeWorkspaceConfig *VSCodeWorkspaceConfig,
	preparedWorkspaceMetadata *PreparedWorkspaceMetadata,
) error {

	parsedRepo, err := github.ParseRepositoryName(
		repoName,
		devEnvRepoOwner,
	)

	if err != nil {
		return err
	}

	// <!> If multiple repos it will clash if same name
	repoDirPathInWorkspace := filepath.Join(
		constants.DevEnvWorkspaceDirPath,
		parsedRepo.Name,
	)

	err = cloneGitHubRepo(
		parsedRepo.Owner,
		parsedRepo.Name,
		repoDirPathInWorkspace,
	)

	if err != nil {
		return err
	}

	repoConfigDirPath := filepath.Join(
		repoDirPathInWorkspace,
		entities.DevEnvRepositoryConfigDirectory,
	)

	workspaceConfigRepository := WorkspaceConfigRepository{
		Owner:         parsedRepo.Owner,
		Name:          parsedRepo.Name,
		RootDirPath:   repoDirPathInWorkspace,
		ConfigDirPath: repoConfigDirPath,
		Hooks:         []WorkspaceConfigRepositoryHook{},
		IsDevEnvRepo: parsedRepo.Name == devEnvRepoName &&
			parsedRepo.Owner == devEnvRepoOwner,
	}

	repoDockerfilePath := filepath.Join(
		repoConfigDirPath,
		entities.DevEnvRepositoryDockerfileFileName,
	)

	filesManager := system.NewFileManager()

	repoHasDockerfile, err := filesManager.DoesFileExist(
		repoDockerfilePath,
	)

	if err != nil {
		return err
	}

	if repoHasDockerfile {

		repoVSCodeExtensions, err := lookupVSCodeExtensionsInDockerfileLabels(
			repoDockerfilePath,
		)

		if err != nil {
			return err
		}

		vscodeWorkspaceConfig.Extensions.Recommendations = mergeVSCodeExtensionsRecos(
			vscodeWorkspaceConfig.Extensions.Recommendations,
			repoVSCodeExtensions,
		)
	}

	// In the case of development environments
	// with only one repository, hooks will only be
	// run if a "dev_env.Dockerfile" file is set.
	//
	// If we don't do this, the repositories that are parts
	// of multi-repositories development environment
	// that are run as single repository will try to run their hooks
	// without any dependencies installed.
	//
	// TODO: Find a more sensible way.
	// Maybe by displaying a warning to users that
	// some repositories need to be run as part of a workspace?
	if preparedWorkspaceMetadata.DevEnvRepoHasDockerfile {
		initHookFilePath := filepath.Join(
			repoConfigDirPath,
			entities.DevEnvRepositoryConfigHooksDirectory,
			entities.DevEnvRepositoryInitHookFileName,
		)

		initHookExists, err := filesManager.DoesFileExist(
			initHookFilePath,
		)

		if err != nil {
			return err
		}

		if initHookExists {
			hookFilePath, err := installHookInWorkspaceConfigDir(
				initHookFilePath,
			)

			if err != nil {
				return err
			}

			workspaceConfigRepository.Hooks = append(
				workspaceConfigRepository.Hooks,
				WorkspaceConfigRepositoryHook{
					ScriptFilePath:       hookFilePath,
					ScriptWorkingDirPath: workspaceConfigRepository.RootDirPath,
				},
			)
		}
	}

	workspaceConfig.Repositories = append(
		workspaceConfig.Repositories,
		workspaceConfigRepository,
	)

	vscodeWorkspaceConfig.Folders = append(
		vscodeWorkspaceConfig.Folders,
		VSCodeWorkspaceConfigFolder{
			Path: repoDirPathInWorkspace,
		},
	)

	return nil
}

func createTmpDirForCloningUserConfigRepo() (string, error) {
	return os.MkdirTemp("", "recode-dev-env-user-config-*")
}

func createTmpDirForCloningDevEnvRepo() (string, error) {
	return os.MkdirTemp("", "recode-dev-env-repo-*")
}
