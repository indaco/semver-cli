package bumpcmd

import (
	"context"

	"github.com/indaco/semver-cli/internal/clix"
	"github.com/indaco/semver-cli/internal/config"
	"github.com/indaco/semver-cli/internal/hooks"
	"github.com/indaco/semver-cli/internal/operations"
	"github.com/indaco/semver-cli/internal/semver"
	"github.com/urfave/cli/v3"
)

// majorCmd returns the "major" subcommand.
func majorCmd(cfg *config.Config) *cli.Command {
	return &cli.Command{
		Name:      "major",
		Usage:     "Increment major version and reset minor and patch",
		UsageText: "semver bump major [--pre label] [--meta data] [--preserve-meta] [--skip-hooks] [--all] [--module name]",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "skip-hooks",
				Usage: "Skip pre-release hooks",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return runBumpMajor(ctx, cmd, cfg)
		},
	}
}

// runBumpMajor increments the major version and resets minor and patch.
func runBumpMajor(ctx context.Context, cmd *cli.Command, cfg *config.Config) error {
	pre := cmd.String("pre")
	meta := cmd.String("meta")
	isPreserveMeta := cmd.Bool("preserve-meta")
	isSkipHooks := cmd.Bool("skip-hooks")

	// Run pre-release hooks first (before any version operations)
	if err := hooks.RunPreReleaseHooksFn(isSkipHooks); err != nil {
		return err
	}

	// Get execution context to determine single vs multi-module mode
	execCtx, err := clix.GetExecutionContext(ctx, cmd, cfg)
	if err != nil {
		return err
	}

	// Handle single-module mode
	if execCtx.IsSingleModule() {
		if _, err := clix.FromCommandFn(cmd); err != nil {
			return err
		}
		return semver.UpdateVersion(execCtx.Path, "major", pre, meta, isPreserveMeta)
	}

	// Handle multi-module mode
	return runMultiModuleBump(ctx, cmd, execCtx, operations.BumpMajor, pre, meta, isPreserveMeta)
}
