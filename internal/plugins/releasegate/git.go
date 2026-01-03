package releasegate

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// Function variables for testability.
var (
	isWorktreeCleanFn  = isWorktreeClean
	getCurrentBranchFn = getCurrentBranch
	getRecentCommitsFn = getRecentCommits
	execCommand        = exec.Command
)

// isWorktreeClean checks if the git working tree has uncommitted changes.
// Returns true if the working tree is clean (no uncommitted changes).
func isWorktreeClean() (bool, error) {
	cmd := execCommand("git", "status", "--porcelain")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		stderrMsg := strings.TrimSpace(stderr.String())
		if stderrMsg != "" {
			return false, fmt.Errorf("failed to check git status: %s: %w", stderrMsg, err)
		}
		return false, fmt.Errorf("failed to check git status: %w", err)
	}

	// Empty output means clean working tree
	output := strings.TrimSpace(stdout.String())
	return output == "", nil
}

// getCurrentBranch retrieves the current git branch name.
func getCurrentBranch() (string, error) {
	cmd := execCommand("git", "rev-parse", "--abbrev-ref", "HEAD")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		stderrMsg := strings.TrimSpace(stderr.String())
		if stderrMsg != "" {
			return "", fmt.Errorf("failed to get current branch: %s: %w", stderrMsg, err)
		}
		return "", fmt.Errorf("failed to get current branch: %w", err)
	}

	branch := strings.TrimSpace(stdout.String())
	if branch == "" {
		return "", fmt.Errorf("failed to determine current branch")
	}

	return branch, nil
}

// getRecentCommits retrieves the last N commit messages.
func getRecentCommits(count int) ([]string, error) {
	if count <= 0 {
		count = 10
	}

	cmd := execCommand("git", "log", fmt.Sprintf("-n%d", count), "--oneline", "--no-decorate")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		stderrMsg := strings.TrimSpace(stderr.String())
		if stderrMsg != "" {
			return nil, fmt.Errorf("failed to get commit history: %s: %w", stderrMsg, err)
		}
		return nil, fmt.Errorf("failed to get commit history: %w", err)
	}

	output := strings.TrimSpace(stdout.String())
	if output == "" {
		return []string{}, nil
	}

	commits := strings.Split(output, "\n")
	return commits, nil
}
