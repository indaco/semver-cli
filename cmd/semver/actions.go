package main

import (
	"context"
	"fmt"
	"os"

	"github.com/indaco/semver-cli/internal/semver"
	"github.com/urfave/cli/v3"
)

func initVersion() func(ctx context.Context, cmd *cli.Command) error {
	return func(ctx context.Context, cmd *cli.Command) error {
		path := cmd.String("path")

		created, err := semver.InitializeVersionFileWithFeedback(path)
		if err != nil {
			return err
		}

		version, err := semver.ReadVersion(path)
		if err != nil {
			return fmt.Errorf("failed to read version file at %s: %w", path, err)
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

		if _, err := getOrInitVersionFile(cmd); err != nil {
			return err
		}

		return semver.UpdateVersion(path, "patch")
	}
}

// bumpMinor increments the minor version and resets patch.
func bumpMinor() func(ctx context.Context, cmd *cli.Command) error {
	return func(ctx context.Context, cmd *cli.Command) error {
		path := cmd.String("path")

		if _, err := getOrInitVersionFile(cmd); err != nil {
			return err
		}

		return semver.UpdateVersion(path, "minor")
	}
}

// bumpMajor increments the major version and resets minor/patch.
func bumpMajor() func(ctx context.Context, cmd *cli.Command) error {
	return func(ctx context.Context, cmd *cli.Command) error {
		path := cmd.String("path")

		if _, err := getOrInitVersionFile(cmd); err != nil {
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

		if _, err := getOrInitVersionFile(cmd); err != nil {
			return err
		}

		version, err := semver.ReadVersion(path)
		if err != nil {
			return fmt.Errorf("failed to read version: %w", err)
		}

		if inc {
			version.PreRelease = semver.IncrementPreRelease(version.PreRelease, label)
		} else {
			if version.PreRelease == "" {
				version.Patch++
			}
			version.PreRelease = label
		}

		if err := semver.SaveVersion(path, version); err != nil {
			return fmt.Errorf("failed to save version: %w", err)
		}

		return nil
	}
}

func showVersion() func(ctx context.Context, cmd *cli.Command) error {
	return func(ctx context.Context, cmd *cli.Command) error {
		_, err := getOrInitVersionFile(cmd)
		if err != nil {
			return err
		}

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

func validateVersion() func(ctx context.Context, cmd *cli.Command) error {
	return func(ctx context.Context, cmd *cli.Command) error {
		path := cmd.String("path")
		_, err := semver.ReadVersion(path)
		if err != nil {
			return fmt.Errorf("invalid version file at %s: %w", path, err)
		}
		fmt.Printf("Valid version file at %s\n", path)
		return nil
	}
}

// getOrInitVersionFile handles .version file initialization or returns an error
// if auto-init is disabled and the file is missing.
func getOrInitVersionFile(cmd *cli.Command) (created bool, err error) {
	path := cmd.String("path")
	noAutoInit := cmd.Bool("no-auto-init")

	if noAutoInit {
		if _, err := os.Stat(path); err != nil {
			return false, cli.Exit(fmt.Sprintf("version file not found at %s", path), 1)
		}
		return false, nil
	}

	created, err = semver.InitializeVersionFileWithFeedback(path)
	if err != nil {
		return false, err
	}
	if created {
		fmt.Printf("Auto-initialized %s with default version\n", path)
	}
	return created, nil
}
