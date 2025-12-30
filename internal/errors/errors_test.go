package errors

import (
	"errors"
	"fmt"
	"testing"
)

func TestVersionFileNotFoundError(t *testing.T) {
	err := &VersionFileNotFoundError{Path: "/path/to/.version"}

	if err.Error() != "version file not found at /path/to/.version" {
		t.Errorf("unexpected error message: %s", err.Error())
	}

	// Test errors.As
	var vfErr *VersionFileNotFoundError
	if !errors.As(err, &vfErr) {
		t.Error("expected errors.As to match VersionFileNotFoundError")
	}
}

func TestInvalidVersionError(t *testing.T) {
	tests := []struct {
		version  string
		reason   string
		expected string
	}{
		{"abc", "not a number", `invalid version format "abc": not a number`},
		{"1.2.x", "", "invalid version format: 1.2.x"},
	}

	for _, tt := range tests {
		err := &InvalidVersionError{Version: tt.version, Reason: tt.reason}
		if err.Error() != tt.expected {
			t.Errorf("expected %q, got %q", tt.expected, err.Error())
		}
	}
}

func TestInvalidBumpTypeError(t *testing.T) {
	err := &InvalidBumpTypeError{BumpType: "huge"}
	expected := "invalid bump type: huge (expected: patch, minor, or major)"

	if err.Error() != expected {
		t.Errorf("expected %q, got %q", expected, err.Error())
	}
}

func TestConfigError(t *testing.T) {
	inner := errors.New("file not found")
	err := &ConfigError{Operation: "load", Err: inner}

	if err.Error() != "config load failed: file not found" {
		t.Errorf("unexpected error message: %s", err.Error())
	}

	if !errors.Is(err, inner) {
		t.Error("expected errors.Is to match inner error")
	}

	if err.Unwrap() != inner {
		t.Error("expected Unwrap to return inner error")
	}
}

func TestCommandError(t *testing.T) {
	inner := fmt.Errorf("exit status 1")

	tests := []struct {
		command  string
		timeout  bool
		expected string
	}{
		{"git", false, `command "git" failed: exit status 1`},
		{"make", true, `command "make" timed out: exit status 1`},
	}

	for _, tt := range tests {
		err := &CommandError{Command: tt.command, Err: inner, Timeout: tt.timeout}
		if err.Error() != tt.expected {
			t.Errorf("expected %q, got %q", tt.expected, err.Error())
		}

		if !errors.Is(err, inner) {
			t.Error("expected errors.Is to match inner error")
		}
	}
}

func TestPathValidationError(t *testing.T) {
	err := &PathValidationError{Path: "../secret", Reason: "path traversal detected"}
	expected := `invalid path "../secret": path traversal detected`

	if err.Error() != expected {
		t.Errorf("expected %q, got %q", expected, err.Error())
	}
}

func TestHookError(t *testing.T) {
	inner := errors.New("make: *** No rule to make target")
	err := &HookError{HookName: "pre-build", Err: inner}

	if err.Error() != `hook "pre-build" failed: make: *** No rule to make target` {
		t.Errorf("unexpected error message: %s", err.Error())
	}

	if !errors.Is(err, inner) {
		t.Error("expected errors.Is to match inner error")
	}
}
