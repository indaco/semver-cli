package bumpcmd

import (
	"context"

	"github.com/indaco/semver-cli/internal/clix"
	"github.com/indaco/semver-cli/internal/hooks"
	"github.com/indaco/semver-cli/internal/semver"
	"github.com/urfave/cli/v3"
)

// patchCmd returns the "patch" subcommand.
func patchCmd() *cli.Command {
	return &cli.Command{
		Name:      "patch",
		Usage:     "Increment patch version",
		UsageText: "semver bump patch [--pre label] [--meta data] [--preserve-meta] [--skip-hooks]",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "skip-hooks",
				Usage: "Skip pre-release hooks",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return runBumpPatch(cmd)
		},
	}
}

// runBumpPatch executes the patch bump logic.
func runBumpPatch(cmd *cli.Command) error {
	path := cmd.String("path")
	pre := cmd.String("pre")
	meta := cmd.String("meta")
	isPreserveMeta := cmd.Bool("preserve-meta")
	isSkipHooks := cmd.Bool("skip-hooks")

	if _, err := clix.FromCommand(cmd); err != nil {
		return err
	}

	if err := hooks.RunPreReleaseHooks(isSkipHooks); err != nil {
		return err
	}
	return semver.UpdateVersion(path, "patch", pre, meta, isPreserveMeta)
}
