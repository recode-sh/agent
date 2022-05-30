package constants

const (
	DevEnvRecodeUserName                      = "recode"
	DevEnvRecodeUserAuthorizedSSHKeysFilePath = "/home/recode/.ssh/authorized_keys"

	DevEnvDockerGroupName                   = "docker"
	DevEnvDockerImageName                   = "recode-dev-env-image"
	DevEnvDockerContainerName               = "recode-dev-env-container"
	DevEnvDockerContainerEntrypointFilePath = "/recode_entrypoint.sh"

	DevEnvWorkspaceDirPath = "/home/recode/workspace"

	DevEnvWorkspaceConfigDirPath        = "/home/recode/.workspace-config"
	DevEnvWorkspaceConfigHooksDirPath   = DevEnvWorkspaceConfigDirPath + "/hooks"
	DevEnvWorkspaceConfigFilePath       = DevEnvWorkspaceConfigDirPath + "/recode.workspace"
	DevEnvVSCodeWorkspaceConfigFilePath = DevEnvWorkspaceConfigDirPath + "/recode.code-workspace"

	DevEnvGitHubPublicSSHKeyFilePath = "/home/recode/.ssh/recode_github.pub"
	DevEnvGitHubPublicGPGKeyFilePath = "/home/recode/.gnupg/recode_github_gpg_public.pgp"
)

var (
	DevEnvDockerContainerStartCmd = []string{
		"sleep",
		"infinity",
	}
)
