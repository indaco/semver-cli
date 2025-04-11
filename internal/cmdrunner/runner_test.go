package cmdrunner

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestRunCommand_Success(t *testing.T) {
	tempDir := os.TempDir()
	err := RunCommand(tempDir, "echo", "Hello, Tempo!")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}

func TestRunCommandWithTimeout_Success(t *testing.T) {
	tempDir := os.TempDir()
	err := RunCommandWithTimeout(tempDir, 10*time.Second, "echo", "Hello, Tempo!")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}

func TestRunCommand_InvalidCommand(t *testing.T) {
	tempDir := os.TempDir()
	err := RunCommand(tempDir, "invalid_command_xyz")
	if err == nil {
		t.Fatal("Expected error for invalid command, got nil")
	}
}

func TestRunCommandWithTimeout_Timeout(t *testing.T) {
	tempDir := os.TempDir()

	// Run a sleep command for longer than the timeout (e.g., 10s sleep, but 2s timeout)
	err := RunCommandWithTimeout(tempDir, 2*time.Second, "sleep", "10") // Linux/Mac
	if err == nil {
		t.Fatal("Expected timeout error, but got nil")
	}
}

func TestRunCommand_Cancel(t *testing.T) {
	tempDir := os.TempDir()

	ctx, cancel := context.WithCancel(context.Background())

	// Run a long-running process (sleep for 10s)
	cmd := exec.CommandContext(ctx, "sleep", "10")
	cmd.Dir = tempDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Start the command
	if err := cmd.Start(); err != nil {
		t.Fatalf("Failed to start command: %v", err)
	}

	// Cancel execution after 1 second
	time.AfterFunc(1*time.Second, cancel)

	// Wait for the command to complete
	err := cmd.Wait()
	if err == nil {
		t.Fatal("Expected error due to cancellation, but got nil")
	} else if ctx.Err() != context.Canceled {
		t.Fatalf("Expected cancellation error, but got: %v", err)
	}
}

func TestRunCommand_InsideDir(t *testing.T) {
	tempDir := os.TempDir()
	testDir := filepath.Join(tempDir, "test-cmd-runner")
	_ = os.MkdirAll(testDir, os.ModePerm)
	defer os.RemoveAll(testDir)

	err := RunCommand(testDir, "touch", "testfile.txt")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify file was created
	testFilePath := filepath.Join(testDir, "testfile.txt")
	if _, err := os.Stat(testFilePath); os.IsNotExist(err) {
		t.Fatal("Expected file to be created but it does not exist")
	}
}

func TestRunCommand_PermissionDenied(t *testing.T) {
	// Try running inside `/root` (which requires sudo)
	err := RunCommand("/root", "echo", "Hello")
	if err == nil {
		t.Fatal("Expected permission error, but got nil")
	}
}

func TestRunCommand_NonExistentDirectory(t *testing.T) {
	invalidDir := filepath.Join(os.TempDir(), "does-not-exist-123456")

	err := RunCommand(invalidDir, "echo", "Hello")
	if err == nil {
		t.Fatal("Expected error for non-existent directory, but got nil")
	}
}

func TestRunCommandOutput_Success(t *testing.T) {
	tempDir := os.TempDir()

	output, err := RunCommandOutput(tempDir, "echo", "Hello, Tempo!")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Trim output for compatibility across OS (newline differences)
	output = strings.TrimSpace(output)

	if output != "Hello, Tempo!" {
		t.Fatalf("Expected output: %q, got: %q", "Hello, Tempo!", output)
	}
}

func TestRunCommandOutput_InvalidCommand(t *testing.T) {
	tempDir := os.TempDir()

	_, err := RunCommandOutput(tempDir, "invalid_command_xyz")
	if err == nil {
		t.Fatal("Expected error for invalid command, got nil")
	}
}

func TestRunCommandOutput_Timeout(t *testing.T) {
	tempDir := os.TempDir()

	// Expect a timeout when running a long-running command with default timeout
	_, err := RunCommandOutput(tempDir, "sleep", "10") // Linux/Mac
	if err == nil {
		t.Fatal("Expected timeout error, but got nil")
	}
}

func TestRunCommandOutput_EmptyOutput(t *testing.T) {
	tempDir := os.TempDir()

	output, err := RunCommandOutput(tempDir, "true") // `true` command has no output
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if output != "" {
		t.Fatalf("Expected empty output, got: %q", output)
	}
}

func TestRunCommandOutput_ErrorOutput(t *testing.T) {
	tempDir := os.TempDir()

	_, err := RunCommandOutput(tempDir, "ls", "/does/not/exist")
	if err == nil {
		t.Fatal("Expected error for invalid path, got nil")
	}
}
