package git

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/indaco/semver-cli/internal/cmdrunner"
)

// Function variables to allow mocking
var (
	CloneOrUpdate = DefaultCloneOrUpdate
	UpdateRepo    = DefaultUpdateRepo
	CloneRepoFunc = CloneRepo
)

func DefaultCloneOrUpdate(repoURL, repoPath string) error {
	if IsValidGitRepo(repoPath) {
		return UpdateRepo(repoPath)
	}
	return CloneRepoFunc(repoURL, repoPath)
}

func DefaultUpdateRepo(repoPath string) error {
	return cmdrunner.RunCommandContext(context.Background(), repoPath, "git", "pull")
}

func CloneRepo(repoURL, repoPath string) error {
	return cmdrunner.RunCommandContext(context.Background(), ".", "git", "clone", repoURL, repoPath)
}

func ForceReclone(repoURL, repoPath string) error {
	if err := os.RemoveAll(repoPath); err != nil {
		return fmt.Errorf("failed to remove existing repository: %w", err)
	}
	return CloneRepo(repoURL, repoPath)
}

func IsValidGitRepo(repoPath string) bool {
	_, err := os.Stat(filepath.Join(repoPath, ".git"))
	return err == nil
}
