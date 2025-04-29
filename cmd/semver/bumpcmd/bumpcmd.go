package bumpcmd

import (
	"github.com/indaco/semver-cli/internal/config"
	"github.com/urfave/cli/v3"
)

// Run returns the "bump" parent command.
func Run(cfg *config.Config) *cli.Command {
	return &cli.Command{
		Name:      "bump",
		Usage:     "Bump semantic version (patch, minor, major)",
		UsageText: "semver bump <subcommand> [--flags]",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "pre",
				Usage: "Optional pre-release label",
			},
			&cli.StringFlag{
				Name:  "meta",
				Usage: "Optional build metadata",
			},
			&cli.BoolFlag{
				Name:  "preserve-meta",
				Usage: "Preserve existing build metadata when bumping",
			},
		},
		Commands: []*cli.Command{
			patchCmd(),
			minorCmd(),
			majorCmd(),
			releaseCmd(),
			nextCmd(cfg),
		},
	}
}
