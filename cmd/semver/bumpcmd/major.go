package bumpcmd

import (
	"context"

	"github.com/indaco/semver-cli/internal/clix"
	"github.com/indaco/semver-cli/internal/hooks"
	"github.com/indaco/semver-cli/internal/semver"
	"github.com/urfave/cli/v3"
)

// majorCmd returns the "major" subcommand.
func majorCmd() *cli.Command {
	return &cli.Command{
		Name:      "major",
		Usage:     "Increment major version and reset minor and patch",
		UsageText: "semver bump major [--pre label] [--meta data] [--preserve-meta] [--skip-hooks]",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "skip-hooks",
				Usage: "Skip pre-release hooks",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return runBumpMajor(cmd)
		},
	}
}

// runBumpMajor increments the minor version and resets patch.
func runBumpMajor(cmd *cli.Command) error {
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

	return semver.UpdateVersion(path, "major", pre, meta, isPreserveMeta)
}
