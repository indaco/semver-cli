package bumpcmd

import (
	"context"

	"github.com/indaco/semver-cli/internal/clix"
	"github.com/indaco/semver-cli/internal/hooks"
	"github.com/indaco/semver-cli/internal/semver"
	"github.com/urfave/cli/v3"
)

// minorCmd returns the "minor" subcommand.
func minorCmd() *cli.Command {
	return &cli.Command{
		Name:      "minor",
		Usage:     "Increment minor version and reset patch",
		UsageText: "semver bump minor [--pre label] [--meta data] [--preserve-meta] [--skip-hooks]",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "skip-hooks",
				Usage: "Skip pre-release hooks",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return runBumpMinor(cmd)
		},
	}
}

// runBumpMinor increments the minor version and resets patch.
func runBumpMinor(cmd *cli.Command) error {
	path := cmd.String("path")
	pre := cmd.String("pre")
	meta := cmd.String("meta")
	isPreserveMeta := cmd.Bool("preserve-meta")
	isSkipHooks := cmd.Bool("skip-hooks")

	if _, err := clix.FromCommandFn(cmd); err != nil {
		return err
	}

	if err := hooks.RunPreReleaseHooksFn(isSkipHooks); err != nil {
		return err
	}

	return semver.UpdateVersion(path, "minor", pre, meta, isPreserveMeta)
}
