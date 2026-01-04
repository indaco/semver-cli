// Package apperrors defines custom error types for the verso application.
// These typed errors enable proper error handling with errors.Is and errors.As
// without coupling internal packages to the CLI framework.
package apperrors

import "fmt"

// VersionFileNotFoundError indicates that the version file does not exist.
type VersionFileNotFoundError struct {
	Path string
}

func (e *VersionFileNotFoundError) Error() string {
	return fmt.Sprintf("version file not found at %s", e.Path)
}

// InvalidVersionError indicates that a version string does not conform to semver.
type InvalidVersionError struct {
	Version string
	Reason  string
}

func (e *InvalidVersionError) Error() string {
	if e.Reason != "" {
		return fmt.Sprintf("invalid version format %q: %s", e.Version, e.Reason)
	}
	return fmt.Sprintf("invalid version format: %s", e.Version)
}

// InvalidBumpTypeError indicates an invalid bump type was specified.
type InvalidBumpTypeError struct {
	BumpType string
}

func (e *InvalidBumpTypeError) Error() string {
	return fmt.Sprintf("invalid bump type: %s (expected: patch, minor, or major)", e.BumpType)
}

// ConfigError indicates a configuration-related error.
type ConfigError struct {
	Operation string
	Err       error
}

func (e *ConfigError) Error() string {
	return fmt.Sprintf("config %s failed: %v", e.Operation, e.Err)
}

func (e *ConfigError) Unwrap() error {
	return e.Err
}

// CommandError indicates a command execution error.
type CommandError struct {
	Command string
	Err     error
	Timeout bool
}

func (e *CommandError) Error() string {
	if e.Timeout {
		return fmt.Sprintf("command %q timed out: %v", e.Command, e.Err)
	}
	return fmt.Sprintf("command %q failed: %v", e.Command, e.Err)
}

func (e *CommandError) Unwrap() error {
	return e.Err
}

// PathValidationError indicates an invalid or dangerous path was provided.
type PathValidationError struct {
	Path   string
	Reason string
}

func (e *PathValidationError) Error() string {
	return fmt.Sprintf("invalid path %q: %s", e.Path, e.Reason)
}

// HookError indicates an error during hook execution.
type HookError struct {
	HookName string
	Err      error
}

func (e *HookError) Error() string {
	return fmt.Sprintf("hook %q failed: %v", e.HookName, e.Err)
}

func (e *HookError) Unwrap() error {
	return e.Err
}
