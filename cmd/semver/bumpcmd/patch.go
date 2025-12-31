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

// patchCmd returns the "patch" subcommand.
func patchCmd(cfg *config.Config) *cli.Command {
	return &cli.Command{
		Name:      "patch",
		Usage:     "Increment patch version",
		UsageText: "semver bump patch [--pre label] [--meta data] [--preserve-meta] [--skip-hooks] [--all] [--module name]",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "skip-hooks",
				Usage: "Skip pre-release hooks",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return runBumpPatch(ctx, cmd, cfg)
		},
	}
}

// runBumpPatch executes the patch bump logic.
func runBumpPatch(ctx context.Context, cmd *cli.Command, cfg *config.Config) error {
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
		return semver.UpdateVersion(execCtx.Path, "patch", pre, meta, isPreserveMeta)
	}

	// Handle multi-module mode
	return runMultiModuleBump(ctx, cmd, execCtx, operations.BumpPatch, pre, meta, isPreserveMeta)
}
