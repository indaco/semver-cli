package versionvalidator

import (
	"os/exec"
	"strings"
)

// getBranchFromGit retrieves the current git branch name.
func getBranchFromGit() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(output)), nil
}
