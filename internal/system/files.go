package system

import (
	"os"
	"path/filepath"
)

type FileManager struct{}

func NewFileManager() FileManager {
	return FileManager{}
}

func (FileManager) ReadFile(filePath string) ([]byte, error) {
	return os.ReadFile(filePath)
}

func (FileManager) DoesFileExist(filePath string) (bool, error) {
	_, err := os.Stat(filePath)

	if os.IsNotExist(err) {
		return false, nil
	}

	if err != nil {
		return false, err
	}

	return true, nil
}

func (FileManager) RemoveDirContent(dirPath string) error {
	fileAndDirPaths, err := filepath.Glob(
		filepath.Join(dirPath, "*"),
	)

	if err != nil {
		return err
	}

	for _, fileOrDirPath := range fileAndDirPaths {
		err = os.RemoveAll(fileOrDirPath)

		if err != nil {
			return err
		}
	}

	return nil
}
