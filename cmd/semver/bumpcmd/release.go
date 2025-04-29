package bumpcmd

import (
	"context"
	"fmt"

	"github.com/indaco/semver-cli/internal/clix"
	"github.com/indaco/semver-cli/internal/hooks"
	"github.com/indaco/semver-cli/internal/semver"
	"github.com/urfave/cli/v3"
)

// releaseCmd returns the "release" subcommand.
func releaseCmd() *cli.Command {
	return &cli.Command{
		Name:      "release",
		Usage:     "Promote pre-release to final version (e.g. 1.2.3-alpha → 1.2.3)",
		UsageText: "semver bump release [--preserve-meta]",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "skip-hooks",
				Usage: "Skip pre-release hooks",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return runBumpRelease(cmd)
		},
	}
}

// runBumpRelease promotes a pre-release version to a final release.
func runBumpRelease(cmd *cli.Command) error {
	path := cmd.String("path")
	isPreserveMeta := cmd.Bool("preserve-meta")
	isSkipHooks := cmd.Bool("skip-hooks")

	if _, err := clix.FromCommand(cmd); err != nil {
		return err
	}

	version, err := semver.ReadVersion(path)
	if err != nil {
		return fmt.Errorf("failed to read version: %w", err)
	}

	if err := hooks.RunPreReleaseHooks(isSkipHooks); err != nil {
		return err
	}

	version.PreRelease = ""
	if !isPreserveMeta {
		version.Build = ""
	}

	if err := semver.SaveVersion(path, version); err != nil {
		return fmt.Errorf("failed to save version: %w", err)
	}

	fmt.Printf("Promoted to release version: %s\n", version.String())
	return nil
}
