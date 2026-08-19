package appdir

import (
	"os"
	"path/filepath"
)

const DefaultDir = ".sen"

func Home() (string, error) {
	if v := os.Getenv("SEN_HOME"); v != "" {
		return filepath.Clean(v), nil
	}
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return filepath.Join(cwd, DefaultDir), nil
}

func Marker(root string) string {
	return filepath.Join(root, "workspace.toml")
}

func Initialized(root string) bool {
	if _, err := os.Stat(Marker(root)); err == nil {
		return true
	}
	_, err := os.Stat(filepath.Join(root, "workspace.yaml"))
	return err == nil
}
