package devenv

import (
	"encoding/json"
	"os"
)

// VSCodeWorkspaceConfig matches .code-workspace schema.
// See: https://code.visualstudio.com/docs/editor/multi-root-workspaces#_workspace-file-schema
type VSCodeWorkspaceConfig struct {
	Folders    []VSCodeWorkspaceConfigFolder   `json:"folders"`
	Settings   map[string]interface{}          `json:"settings"`
	Extensions VSCodeWorkspaceConfigExtensions `json:"extensions"`
}

type VSCodeWorkspaceConfigFolder struct {
	Path string `json:"path"`
}

type VSCodeWorkspaceConfigExtensions struct {
	Recommendations []string `json:"recommendations"`
}

func buildInitialVSCodeWorkspaceConfig() VSCodeWorkspaceConfig {
	return VSCodeWorkspaceConfig{
		Folders: []VSCodeWorkspaceConfigFolder{},
		Settings: map[string]interface{}{
			"remote.autoForwardPorts":      true,
			"remote.restoreForwardedPorts": true,
			// Auto-detect (using "/proc") and forward opened port.
			// Way better than "output" that parse terminal output.
			// See: https://github.com/microsoft/vscode/issues/143958#issuecomment-1050959241
			"remote.autoForwardPortsSource": "process",
		},
		Extensions: VSCodeWorkspaceConfigExtensions{
			Recommendations: []string{},
		},
	}
}

func LoadVSCodeWorkspaceConfig(
	vscodeWorkspaceConfigFilePath string,
) (*VSCodeWorkspaceConfig, error) {

	vscodeWorkspaceConfigFileContent, err := os.ReadFile(vscodeWorkspaceConfigFilePath)

	if err != nil {
		return nil, err
	}

	var vscodeWorkspaceConfig *VSCodeWorkspaceConfig
	err = json.Unmarshal(
		vscodeWorkspaceConfigFileContent,
		&vscodeWorkspaceConfig,
	)

	if err != nil {
		return nil, err
	}

	return vscodeWorkspaceConfig, nil
}

func saveVSCodeWorkspaceConfigAsFile(
	vscodeWorkspaceConfigFilePath string,
	vscodeWorkspaceConfig VSCodeWorkspaceConfig,
) error {

	vscodeWorkspaceConfigAsJSON, err := json.Marshal(&vscodeWorkspaceConfig)

	if err != nil {
		return err
	}

	return os.WriteFile(
		vscodeWorkspaceConfigFilePath,
		vscodeWorkspaceConfigAsJSON,
		os.FileMode(0644),
	)
}

func mergeVSCodeExtensionsRecos(
	currentRecos []string,
	recosToAdd []string,
) []string {

	allRecos := []string{}
	hasRecoMap := map[string]bool{}

	for _, currentReco := range currentRecos {
		allRecos = append(allRecos, currentReco)
		hasRecoMap[currentReco] = true
	}

	for _, recoToAdd := range recosToAdd {
		_, alreadyHasReco := hasRecoMap[recoToAdd]

		if alreadyHasReco {
			continue
		}

		allRecos = append(allRecos, recoToAdd)
		hasRecoMap[recoToAdd] = true
	}

	return allRecos
}
