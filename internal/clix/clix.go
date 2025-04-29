package clix

import (
	"fmt"
	"os"

	"github.com/indaco/semver-cli/internal/semver"
	"github.com/urfave/cli/v3"
)

var FromCommandFn = fromCommand

// fromCommand extracts the --path and --no-auto-init flags from a cli.Command,
// and passes them to GetOrInitVersionFile.
func fromCommand(cmd *cli.Command) (bool, error) {
	return getOrInitVersionFile(cmd.String("path"), cmd.Bool("no-auto-init"))
}

// getOrInitVersionFile initializes the version file at the given path
// or checks for its existence based on the noAutoInit flag.
// It returns true if the file was created, false if it already existed.
func getOrInitVersionFile(path string, noAutoInit bool) (bool, error) {
	if noAutoInit {
		if _, err := os.Stat(path); err != nil {
			return false, cli.Exit(fmt.Sprintf("version file not found at %s", path), 1)
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
