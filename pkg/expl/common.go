package expl

import (
	"os"
	"path"
	"path/filepath"
)

const Version string = "1.0.5"

func getPath(filePath string) (string, error) {

	if path.IsAbs(filePath) {
		return filePath, nil
	}

	p, err := os.Getwd()

	return path.Clean(filepath.Join(p, filePath)), err
}
