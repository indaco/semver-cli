package main

import (
	"context"
	"fmt"

	"github.com/indaco/semver-cli/cmd/semver/bumpcmd"
	"github.com/indaco/semver-cli/cmd/semver/doctorcmd"
	"github.com/indaco/semver-cli/cmd/semver/extensioncmd"
	"github.com/indaco/semver-cli/cmd/semver/initcmd"
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
				Name:  "no-auto-init",
				Usage: "Disable auto-initialization of the .version file",
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
			showcmd.Run(),
			setcmd.Run(),
			bumpcmd.Run(cfg),
			precmd.Run(),
			doctorcmd.Run(),
			initcmd.Run(),
			extensioncmd.Run(),
		},
	}
}
