package config

import (
	"fmt"
	"os/exec"
	"strings"
)

func DetectRepo() (string, error) {
	cmd := exec.Command("git", "remote", "get-url", "origin")
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get git remote: %w", err)
	}

	url := strings.TrimSpace(string(out))
	url = strings.TrimPrefix(url, "git@github.com:")
	url = strings.TrimSuffix(url, ".git")
	url = strings.TrimPrefix(url, "https://github.com/")

	parts := strings.Split(url, "/")
	if len(parts) != 2 {
		return "", fmt.Errorf("could not parse repository from remote: %s", url)
	}

	return strings.Join(parts, "/"), nil
}
