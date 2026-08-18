package appdir

import (
	"os"
	"path/filepath"
)

func Home() (string, error) {
	if v := os.Getenv("SEN_HOME"); v != "" {
		return filepath.Clean(v), nil
	}

	if xdg := os.Getenv("XDG_DATA_HOME"); xdg != "" {
		return filepath.Join(xdg, "sen"), nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".local", "share", "sen"), nil
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
