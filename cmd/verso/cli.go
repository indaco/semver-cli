package main

import (
	"context"
	"fmt"

	"github.com/indaco/verso/cmd/verso/bumpcmd"
	"github.com/indaco/verso/cmd/verso/doctorcmd"
	"github.com/indaco/verso/cmd/verso/extensioncmd"
	"github.com/indaco/verso/cmd/verso/initcmd"
	"github.com/indaco/verso/cmd/verso/modulescmd"
	"github.com/indaco/verso/cmd/verso/precmd"
	"github.com/indaco/verso/cmd/verso/setcmd"
	"github.com/indaco/verso/cmd/verso/showcmd"
	"github.com/indaco/verso/internal/config"
	"github.com/indaco/verso/internal/console"
	"github.com/indaco/verso/internal/version"
	"github.com/urfave/cli/v3"
)

var noColorFlag bool

// newCLI builds and returns the root CLI command,
// configuring all subcommands and flags for the verso cli.
func newCLI(cfg *config.Config) *cli.Command {
	return &cli.Command{
		Name:    "verso",
		Version: fmt.Sprintf("v%s", version.GetVersion()),
		Usage:   "Version orchestrator for semantic versioning",
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
