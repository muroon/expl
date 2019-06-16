package service

import (
	"os"
	"path"
	"path/filepath"
)

func getPath(filePath string) (string, error) {

	if path.IsAbs(filePath) {
		return filePath, nil
	}

	p, err := os.Getwd()

	return path.Clean(filepath.Join(p, filePath)), err
}
