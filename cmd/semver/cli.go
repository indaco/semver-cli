package main

import (
	"context"
	"fmt"

	"github.com/indaco/semver-cli/cmd/semver/bumpcmd"
	"github.com/indaco/semver-cli/cmd/semver/doctorcmd"
	"github.com/indaco/semver-cli/cmd/semver/extensioncmd"
	"github.com/indaco/semver-cli/cmd/semver/initcmd"
	"github.com/indaco/semver-cli/cmd/semver/modulescmd"
	"github.com/indaco/semver-cli/cmd/semver/precmd"
	"github.com/indaco/semver-cli/cmd/semver/setcmd"
	"github.com/indaco/semver-cli/cmd/semver/showcmd"
	"github.com/indaco/semver-cli/internal/config"
	"github.com/indaco/semver-cli/internal/console"
	"github.com/indaco/semver-cli/internal/version"
	"github.com/urfave/cli/v3"
)

var noColorFlag bool

// newCLI builds and returns the root CLI command,
// configuring all subcommands and flags for the semver cli.
func newCLI(cfg *config.Config) *cli.Command {
	return &cli.Command{
		Name:    "semver",
		Version: fmt.Sprintf("v%s", version.GetVersion()),
		Usage:   "Manage semantic versioning with a .version file",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "path",
				Aliases: []string{"p"},
				Usage:   "Path to .version file",
				Value:   cfg.Path,
			},
			&cli.BoolFlag{
				Name:    "strict",
				Aliases: []string{"no-auto-init"},
				Usage:   "Fail if .version file is missing (disable auto-initialization)",
			},
			&cli.BoolFlag{
				Name:        "no-color",
				Usage:       "Disable colored output",
				Destination: &noColorFlag,
			},
			&cli.BoolFlag{
				Name:    "all",
				Aliases: []string{"a"},
				Usage:   "Operate on all discovered modules (multi-module mode)",
			},
			&cli.StringFlag{
				Name:    "module",
				Aliases: []string{"m"},
				Usage:   "Operate on specific module by name (multi-module mode)",
			},
			&cli.StringSliceFlag{
				Name:  "modules",
				Usage: "Operate on multiple modules (comma-separated names)",
			},
			&cli.StringFlag{
				Name:  "pattern",
				Usage: "Operate on modules matching glob pattern (e.g., 'services/*')",
			},
			&cli.BoolFlag{
				Name:    "yes",
				Aliases: []string{"y"},
				Usage:   "Auto-select all modules without prompting (implies --all)",
			},
			&cli.BoolFlag{
				Name:  "non-interactive",
				Usage: "Disable interactive prompts (CI mode)",
			},
			&cli.BoolFlag{
				Name:  "parallel",
				Usage: "Execute operations in parallel across modules",
			},
			&cli.BoolFlag{
				Name:  "fail-fast",
				Usage: "Stop execution on first error (default: true)",
				Value: true,
			},
			&cli.BoolFlag{
				Name:  "continue-on-error",
				Usage: "Continue execution even if some modules fail",
			},
			&cli.BoolFlag{
				Name:    "quiet",
				Aliases: []string{"q"},
				Usage:   "Suppress module-level output, show summary only",
			},
			&cli.StringFlag{
				Name:  "format",
				Usage: "Output format: text, json, table",
				Value: "text",
			},
		},
		Before: func(ctx context.Context, cmd *cli.Command) (context.Context, error) {
			console.SetNoColor(noColorFlag)
			return ctx, nil
		},
		Commands: []*cli.Command{
			showcmd.Run(cfg),
			setcmd.Run(cfg),
			bumpcmd.Run(cfg),
			precmd.Run(),
			doctorcmd.Run(),
			initcmd.Run(),
			extensioncmd.Run(),
			modulescmd.Run(),
		},
	}
}
