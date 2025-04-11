package cmdrunner

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"
)

// RunCommandWithTimeout executes a command with a specified timeout.
func RunCommandWithTimeout(dir string, timeout time.Duration, command string, args ...string) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, command, args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("command timed out after %v: %w", timeout, err)
		}
		return fmt.Errorf("command failed: %w", err)
	}

	return nil
}

// RunCommand executes a command with a default timeout of 30 seconds.
func RunCommand(dir string, command string, args ...string) error {
	return RunCommandWithTimeout(dir, 30*time.Second, command, args...)
}

// RunCommandOutput executes a command and returns its output while enforcing a timeout.
func RunCommandOutput(dir string, command string, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second) // Adjust timeout
	defer cancel()

	cmd := exec.CommandContext(ctx, command, args...)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()

	if ctx.Err() == context.DeadlineExceeded {
		return "", fmt.Errorf("command timed out: %w", err)
	}

	if err != nil {
		return "", fmt.Errorf("command failed: %w", err)
	}

	return string(output), nil
}
