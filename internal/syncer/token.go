package syncer

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func GitHubToken() (string, error) {
	if t := strings.TrimSpace(os.Getenv("GITHUB_TOKEN")); t != "" {
		return t, nil
	}
	cmd := exec.Command("gh", "auth", "token")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("github token: set GITHUB_TOKEN or run gh auth login: %w %s", err, strings.TrimSpace(stderr.String()))
	}
	t := strings.TrimSpace(string(out))
	if t == "" {
		return "", fmt.Errorf("github token: empty")
	}
	return t, nil
}
