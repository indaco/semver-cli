package testutils

import (
	"bytes"
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/urfave/cli/v3"
)

func ReadFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func WriteFile(t *testing.T, path, content string, perm fs.FileMode) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), perm); err != nil {
		t.Fatalf("failed to write file %q: %v", path, err)
	}
}

func ReadTempVersionFile(t *testing.T, dir string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(dir, ".version"))
	if err != nil {
		t.Fatalf("failed to read .version file: %v", err)
	}
	return strings.TrimSpace(string(data))
}

func WriteTempVersionFile(t *testing.T, dir, version string) string {
	t.Helper()
	path := filepath.Join(dir, ".version")
	WriteFile(t, path, version, 0644)

	return path
}

func WriteTempConfig(t *testing.T, content string) string {
	t.Helper()
	tmpDir := t.TempDir()
	tmpPath := filepath.Join(tmpDir, ".semver.yaml")

	WriteFile(t, tmpPath, content, 0644)
	return tmpPath
}

func CaptureStdout(f func()) (string, error) {
	// Save original stdout, stderr, and color output
	origStdout, origStderr := os.Stdout, os.Stderr

	// Create pipes to capture stdout and stderr
	rOut, wOut, err := os.Pipe()
	if err != nil {
		return "", err
	}
	rErr, wErr, err := os.Pipe()
	if err != nil {
		return "", err
	}

	// Redirect output
	os.Stdout, os.Stderr = wOut, wErr

	// Capture output concurrently
	outputChan := make(chan string)
	go func() {
		var bufOut, bufErr bytes.Buffer
		_, _ = bufOut.ReadFrom(rOut)
		_, _ = bufErr.ReadFrom(rErr)
		outputChan <- bufOut.String() + bufErr.String()
	}()

	// Execute the function
	f()

	// Close pipes and restore output
	wOut.Close()
	wErr.Close()
	os.Stdout, os.Stderr = origStdout, origStderr

	// Retrieve captured output
	output := <-outputChan
	return strings.TrimSpace(output), nil
}

func IsWindows() bool {
	return strings.Contains(strings.ToLower(os.Getenv("OS")), "windows")
}

func RunCLITest(t *testing.T, appCli *cli.Command, args []string, workdir string) {
	t.Helper()

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}

	if err := os.Chdir(workdir); err != nil {
		t.Fatalf("failed to change to workdir: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(origDir); err != nil {
			t.Fatalf("failed to restore working directory: %v", err)
		}
	})

	err = appCli.Run(context.Background(), args)
	if err != nil {
		t.Fatalf("app.Run failed: %v", err)
	}
}
