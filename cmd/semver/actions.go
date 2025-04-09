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

		created, _, err := semver.InitializeVersionFileWithFeedback(path)
		if err != nil {
			return err
		}
		if created {
			fmt.Printf("Auto-initialized %s with default version\n", path)
		}

		return semver.UpdateVersion(path, "patch")
	}
}

// bumpMinor increments the minor version and resets patch.
func bumpMinor() func(ctx context.Context, cmd *cli.Command) error {
	return func(ctx context.Context, cmd *cli.Command) error {
		path := cmd.String("path")

		created, _, err := semver.InitializeVersionFileWithFeedback(path)
		if err != nil {
			return err
		}
		if created {
			fmt.Printf("Auto-initialized %s with default version\n", path)
		}

		return semver.UpdateVersion(path, "minor")
	}
}

// bumpMajor increments the major version and resets minor/patch.
func bumpMajor() func(ctx context.Context, cmd *cli.Command) error {
	return func(ctx context.Context, cmd *cli.Command) error {
		path := cmd.String("path")

		created, _, err := semver.InitializeVersionFileWithFeedback(path)
		if err != nil {
			return err
		}
		if created {
			fmt.Printf("Auto-initialized %s with default version\n", path)
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

		created, version, err := semver.InitializeVersionFileWithFeedback(path)
		if err != nil {
			return err
		}
		if created {
			fmt.Printf("Auto-initialized %s with default version\n", path)
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
			return fmt.Errorf("failed to read version file at %s: %w", path, err)
		}

		fmt.Println(version.String())
		return nil
	}
}

func setVersion() func(ctx context.Context, cmd *cli.Command) error {
	return func(ctx context.Context, cmd *cli.Command) error {
		path := cmd.String("path")
		args := cmd.Args()

		if args.Len() < 1 {
			return cli.Exit("missing required version argument", 1)
		}

		raw := args.Get(0)
		pre := cmd.String("pre")

		version, err := semver.ParseVersion(raw)
		if err != nil {
			return fmt.Errorf("invalid version: %w", err)
		}
		version.PreRelease = pre

		if err := semver.SaveVersion(path, version); err != nil {
			return fmt.Errorf("failed to save version: %w", err)
		}

		fmt.Printf("Set version to %s in %s\n", version.String(), path)
		return nil
	}
}
