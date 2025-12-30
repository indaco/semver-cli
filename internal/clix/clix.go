package clix

import (
	stderrors "errors"
	"fmt"
	"os"

	"github.com/indaco/semver-cli/internal/apperrors"
	"github.com/indaco/semver-cli/internal/semver"
	"github.com/urfave/cli/v3"
)

var FromCommandFn = fromCommand

// fromCommand extracts the --path and --strict flags from a cli.Command,
// and passes them to GetOrInitVersionFile.
func fromCommand(cmd *cli.Command) (bool, error) {
	return getOrInitVersionFile(cmd.String("path"), cmd.Bool("strict"))
}

// GetOrInitVersionFile initializes the version file at the given path
// or checks for its existence based on the strict flag.
// It returns true if the file was created, false if it already existed.
// Returns a typed error (*apperrors.VersionFileNotFoundError) instead of cli.Exit.
func GetOrInitVersionFile(path string, strict bool) (bool, error) {
	if strict {
		if _, err := os.Stat(path); err != nil {
			return false, &apperrors.VersionFileNotFoundError{Path: path}
		}
		return false, nil
	}

	created, err := semver.InitializeVersionFileWithFeedback(path)
	if err != nil {
		return false, err
	}
	if created {
		fmt.Printf("Auto-initialized %s with default version\n", path)
	}
	return created, nil
}

// getOrInitVersionFile is the internal implementation that wraps errors for CLI display.
//
// Deprecated: Use GetOrInitVersionFile and handle errors at the CLI layer.
func getOrInitVersionFile(path string, strict bool) (bool, error) {
	created, err := GetOrInitVersionFile(path, strict)
	if err != nil {
		// Convert typed errors to CLI exits for backward compatibility
		var vfErr *apperrors.VersionFileNotFoundError
		if stderrors.As(err, &vfErr) {
			return false, cli.Exit(vfErr.Error(), 1)
		}
		return false, err
	}
	return created, nil
}
