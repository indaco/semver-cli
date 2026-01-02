package tagmanager

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// Function variables for testability.
var (
	createAnnotatedTagFn   = createAnnotatedTag
	createLightweightTagFn = createLightweightTag
	tagExistsFn            = tagExists
	getLatestTagFn         = getLatestTag
	pushTagFn              = pushTag
	execCommand            = exec.Command
)

// createAnnotatedTag creates an annotated git tag with the given name and message.
func createAnnotatedTag(name, message string) error {
	cmd := execCommand("git", "tag", "-a", name, "-m", message)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		stderrMsg := strings.TrimSpace(stderr.String())
		if stderrMsg != "" {
			return fmt.Errorf("%s: %w", stderrMsg, err)
		}
		return err
	}
	return nil
}

// createLightweightTag creates a lightweight git tag with the given name.
func createLightweightTag(name string) error {
	cmd := execCommand("git", "tag", name)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		stderrMsg := strings.TrimSpace(stderr.String())
		if stderrMsg != "" {
			return fmt.Errorf("%s: %w", stderrMsg, err)
		}
		return err
	}
	return nil
}

// tagExists checks if a git tag with the given name exists.
func tagExists(name string) (bool, error) {
	cmd := execCommand("git", "tag", "-l", name)
	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	if err := cmd.Run(); err != nil {
		return false, fmt.Errorf("failed to list tags: %w", err)
	}

	// If the tag exists, git tag -l will output the tag name
	output := strings.TrimSpace(stdout.String())
	return output == name, nil
}

// getLatestTag returns the most recent semver tag from git.
func getLatestTag() (string, error) {
	cmd := execCommand("git", "describe", "--tags", "--abbrev=0")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		stderrMsg := strings.TrimSpace(stderr.String())
		if stderrMsg != "" {
			return "", fmt.Errorf("%s: %w", stderrMsg, err)
		}
		return "", fmt.Errorf("no tags found: %w", err)
	}

	tag := strings.TrimSpace(stdout.String())
	if tag == "" {
		return "", fmt.Errorf("no tags found")
	}

	return tag, nil
}

// pushTag pushes a specific tag to the remote.
func pushTag(name string) error {
	cmd := execCommand("git", "push", "origin", name)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		stderrMsg := strings.TrimSpace(stderr.String())
		if stderrMsg != "" {
			return fmt.Errorf("%s: %w", stderrMsg, err)
		}
		return err
	}
	return nil
}

// ListTags returns all git tags matching a pattern.
func ListTags(pattern string) ([]string, error) {
	args := []string{"tag", "-l"}
	if pattern != "" {
		args = append(args, pattern)
	}

	cmd := execCommand("git", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		stderrMsg := strings.TrimSpace(stderr.String())
		if stderrMsg != "" {
			return nil, fmt.Errorf("%s: %w", stderrMsg, err)
		}
		return nil, err
	}

	output := strings.TrimSpace(stdout.String())
	if output == "" {
		return []string{}, nil
	}

	return strings.Split(output, "\n"), nil
}

// DeleteTag deletes a local git tag.
func DeleteTag(name string) error {
	cmd := execCommand("git", "tag", "-d", name)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		stderrMsg := strings.TrimSpace(stderr.String())
		if stderrMsg != "" {
			return fmt.Errorf("%s: %w", stderrMsg, err)
		}
		return err
	}
	return nil
}
