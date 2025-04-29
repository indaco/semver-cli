package testutils

import (
	"bytes"
	"os"
	"strings"
)

// CaptureStdout captures both stdout and stderr output produced during the execution of f.
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
