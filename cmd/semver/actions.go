package main

import (
	"context"
	"fmt"

	"github.com/indaco/semver-cli/internal/semver"
	"github.com/urfave/cli/v3"
)

// bumpPatch increments the patch version of the .version file.
func bumpPatch() func(ctx context.Context, cmd *cli.Command) error {
	return func(ctx context.Context, cmd *cli.Command) error {
		return semver.UpdateVersion(cmd.String("path"), "patch")
	}
}

// bumpMinor increments the minor version and resets patch.
func bumpMinor() func(ctx context.Context, cmd *cli.Command) error {
	return func(ctx context.Context, cmd *cli.Command) error {
		return semver.UpdateVersion(cmd.String("path"), "minor")
	}
}

// bumpMajor increments the major version and resets minor/patch.
func bumpMajor() func(ctx context.Context, cmd *cli.Command) error {
	return func(ctx context.Context, cmd *cli.Command) error {
		return semver.UpdateVersion(cmd.String("path"), "major")
	}
}

// setPreRelease sets or increments the pre-release label.
func setPreRelease() func(ctx context.Context, cmd *cli.Command) error {
	return func(ctx context.Context, cmd *cli.Command) error {
		path := cmd.String("path")
		label := cmd.String("label")
		inc := cmd.Bool("inc")

		version, err := semver.ReadVersion(path)
		if err != nil {
			return err
		}

		if inc {
			version.PreRelease = semver.IncrementPreRelease(version.PreRelease, label)
		} else {
			if version.PreRelease == "" {
				version.Patch++ // <-- Bump patch before applying label
			}
			version.PreRelease = label
		}

		return semver.WriteVersion(path, version)
	}
}

// showVersion prints the current version to stdout.
func showVersion() func(ctx context.Context, cmd *cli.Command) error {
	return func(ctx context.Context, cmd *cli.Command) error {
		version, err := semver.ReadVersion(cmd.String("path"))
		if err != nil {
			return err
		}
		fmt.Println(version.String())
		return nil
	}
}
