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

func TestRunCommandContext_Success(t *testing.T) {
	tempDir := os.TempDir()
	err := RunCommandContext(context.Background(), tempDir, "echo", "Hello, Tempo!")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}

func TestRunCommandContext_WithTimeout(t *testing.T) {
	tempDir := os.TempDir()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := RunCommandContext(ctx, tempDir, "echo", "Hello, Tempo!")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}

func TestRunCommandContext_InvalidCommand(t *testing.T) {
	tempDir := os.TempDir()
	err := RunCommandContext(context.Background(), tempDir, "invalid_command_xyz")
	if err == nil {
		t.Fatal("Expected error for invalid command, got nil")
	}
}

func TestRunCommandContext_Timeout(t *testing.T) {
	tempDir := os.TempDir()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Run a sleep command for longer than the timeout (e.g., 10s sleep, but 2s timeout)
	err := RunCommandContext(ctx, tempDir, "sleep", "10") // Linux/Mac
	if err == nil {
		t.Fatal("Expected timeout error, but got nil")
	}
}

func TestRunCommandContext_Cancel(t *testing.T) {
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

func TestRunCommandContext_InsideDir(t *testing.T) {
	tempDir := os.TempDir()
	testDir := filepath.Join(tempDir, "test-cmd-runner")
	_ = os.MkdirAll(testDir, os.ModePerm)
	defer os.RemoveAll(testDir)

	err := RunCommandContext(context.Background(), testDir, "touch", "testfile.txt")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify file was created
	testFilePath := filepath.Join(testDir, "testfile.txt")
	if _, err := os.Stat(testFilePath); os.IsNotExist(err) {
		t.Fatal("Expected file to be created but it does not exist")
	}
}

func TestRunCommandContext_PermissionDenied(t *testing.T) {
	// Try running inside `/root` (which requires sudo)
	err := RunCommandContext(context.Background(), "/root", "echo", "Hello")
	if err == nil {
		t.Fatal("Expected permission error, but got nil")
	}
}

func TestRunCommandContext_NonExistentDirectory(t *testing.T) {
	invalidDir := filepath.Join(os.TempDir(), "does-not-exist-123456")

	err := RunCommandContext(context.Background(), invalidDir, "echo", "Hello")
	if err == nil {
		t.Fatal("Expected error for non-existent directory, but got nil")
	}
}

func TestRunCommandOutputContext_Success(t *testing.T) {
	tempDir := os.TempDir()
	ctx, cancel := context.WithTimeout(context.Background(), DefaultOutputTimeout)
	defer cancel()

	output, err := RunCommandOutputContext(ctx, tempDir, "echo", "Hello, Tempo!")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Trim output for compatibility across OS (newline differences)
	output = strings.TrimSpace(output)

	if output != "Hello, Tempo!" {
		t.Fatalf("Expected output: %q, got: %q", "Hello, Tempo!", output)
	}
}

func TestRunCommandOutputContext_InvalidCommand(t *testing.T) {
	tempDir := os.TempDir()
	ctx, cancel := context.WithTimeout(context.Background(), DefaultOutputTimeout)
	defer cancel()

	_, err := RunCommandOutputContext(ctx, tempDir, "invalid_command_xyz")
	if err == nil {
		t.Fatal("Expected error for invalid command, got nil")
	}
}

func TestRunCommandOutputContext_Timeout(t *testing.T) {
	tempDir := os.TempDir()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Expect a timeout when running a long-running command
	_, err := RunCommandOutputContext(ctx, tempDir, "sleep", "10") // Linux/Mac
	if err == nil {
		t.Fatal("Expected timeout error, but got nil")
	}
}

func TestRunCommandOutputContext_EmptyOutput(t *testing.T) {
	tempDir := os.TempDir()
	ctx, cancel := context.WithTimeout(context.Background(), DefaultOutputTimeout)
	defer cancel()

	output, err := RunCommandOutputContext(ctx, tempDir, "true") // `true` command has no output
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if output != "" {
		t.Fatalf("Expected empty output, got: %q", output)
	}
}

func TestRunCommandOutputContext_ErrorOutput(t *testing.T) {
	tempDir := os.TempDir()
	ctx, cancel := context.WithTimeout(context.Background(), DefaultOutputTimeout)
	defer cancel()

	_, err := RunCommandOutputContext(ctx, tempDir, "ls", "/does/not/exist")
	if err == nil {
		t.Fatal("Expected error for invalid path, got nil")
	}
}
