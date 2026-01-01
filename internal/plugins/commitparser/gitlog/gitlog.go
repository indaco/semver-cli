package gitlog

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

var (
	GetCommitsFn = getCommits
	execCommand  = exec.Command
)

func getCommits(since string, until string) ([]string, error) {
	if until == "" {
		until = "HEAD"
	}

	if since == "" {
		lastTag, err := getLastTag()
		if err != nil {
			since = "HEAD~10"
		} else {
			since = lastTag
		}
	}

	revRange := since + ".." + until
	cmd := execCommand("git", "log", "--pretty=format:%s", revRange)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	output, err := cmd.Output()
	if err != nil {
		stderrMsg := strings.TrimSpace(stderr.String())
		if stderrMsg != "" {
			return nil, fmt.Errorf("git log failed: %s: %w", stderrMsg, err)
		}
		return nil, fmt.Errorf("git log failed: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) == 1 && lines[0] == "" {
		return []string{}, nil
	}
	return lines, nil
}

func getLastTag() (string, error) {
	cmd := execCommand("git", "describe", "--tags", "--abbrev=0")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	out, err := cmd.Output()
	if err != nil {
		stderrMsg := strings.TrimSpace(stderr.String())
		if stderrMsg != "" {
			return "", fmt.Errorf("git describe failed: %s: %w", stderrMsg, err)
		}
		return "", fmt.Errorf("git describe failed: %w", err)
	}

	tag := strings.TrimSpace(string(out))
	if tag == "" {
		return "", fmt.Errorf("no tags found")
	}

	return tag, nil
}
