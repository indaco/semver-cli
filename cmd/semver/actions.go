package main

import (
	"context"
	"fmt"

	"github.com/indaco/semver-cli/internal/semver"
	"github.com/urfave/cli/v3"
)

func initVersion() func(ctx context.Context, cmd *cli.Command) error {
	return func(ctx context.Context, cmd *cli.Command) error {
		path := cmd.String("path")
		created, version, err := semver.InitializeVersionFileWithFeedback(path)
		if err != nil {
			return err
		}

		if created {
			fmt.Printf("Initialized %s with version %s\n", path, version.String())
		} else {
			fmt.Printf("Version file already exists at %s\n", path)
		}
		return nil
	}
}

// bumpPatch increments the patch version of the .version file.
func bumpPatch() func(ctx context.Context, cmd *cli.Command) error {
	return func(ctx context.Context, cmd *cli.Command) error {
		path := cmd.String("path")
		if err := semver.InitializeVersionFile(path); err != nil {
			return err
		}
		return semver.UpdateVersion(path, "patch")
	}
}

// bumpMinor increments the minor version and resets patch.
func bumpMinor() func(ctx context.Context, cmd *cli.Command) error {
	return func(ctx context.Context, cmd *cli.Command) error {
		path := cmd.String("path")
		if err := semver.InitializeVersionFile(path); err != nil {
			return err
		}
		return semver.UpdateVersion(path, "minor")
	}
}

// bumpMajor increments the major version and resets minor/patch.
func bumpMajor() func(ctx context.Context, cmd *cli.Command) error {
	return func(ctx context.Context, cmd *cli.Command) error {
		path := cmd.String("path")
		if err := semver.InitializeVersionFile(path); err != nil {
			return err
		}
		return semver.UpdateVersion(path, "major")
	}
}

// setPreRelease sets or increments the pre-release label.
func setPreRelease() func(ctx context.Context, cmd *cli.Command) error {
	return func(ctx context.Context, cmd *cli.Command) error {
		path := cmd.String("path")
		label := cmd.String("label")
		inc := cmd.Bool("inc")

		if err := semver.InitializeVersionFile(path); err != nil {
			return err
		}

		version, err := semver.ReadVersion(path)
		if err != nil {
			return err
		}

		if inc {
			version.PreRelease = semver.IncrementPreRelease(version.PreRelease, label)
		} else {
			if version.PreRelease == "" {
				version.Patch++
			}
			version.PreRelease = label
		}

		return semver.SaveVersion(path, version)
	}
}

// showVersion prints the current version to stdout.
func showVersion() func(ctx context.Context, cmd *cli.Command) error {
	return func(ctx context.Context, cmd *cli.Command) error {
		path := cmd.String("path")

		version, err := semver.ReadVersion(path)
		if err != nil {
			return err // do not fallback or initialize
		}

		fmt.Println(version.String())
		return nil
	}
}
