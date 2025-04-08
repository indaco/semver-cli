package main

import (
	"fmt"

	"github.com/indaco/semver-cli/internal/version"
	"github.com/urfave/cli/v3"
)

// newCLI builds and returns the root CLI command,
// configuring all subcommands and flags for the semver tool.
func newCLI(defaultPath string) (*cli.Command, error) {

	return &cli.Command{
		Name:    "semver",
		Version: fmt.Sprintf("v%s", version.GetVersion()),
		Usage:   "Manage semantic versioning with a .version file",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "path",
				Aliases: []string{"p"},
				Usage:   "Path to .version file",
				Value:   defaultPath,
			},
		},
		Commands: []*cli.Command{
			{
				Name:   "patch",
				Usage:  "Increment patch version",
				Action: bumpPatch(),
			},
			{
				Name:   "minor",
				Usage:  "Increment minor version and reset patch",
				Action: bumpMinor(),
			},
			{
				Name:   "major",
				Usage:  "Increment major version and reset minor and patch",
				Action: bumpMajor(),
			},
			{
				Name:  "pre",
				Usage: "Set pre-release label (e.g., alpha, beta.1)",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "label",
						Usage:    "Pre-release label to set",
						Required: true,
					},
					&cli.BoolFlag{
						Name:  "inc",
						Usage: "Increment numeric suffix if it exists or add '.1'",
					},
				},
				Action: setPreRelease(),
			},
			{
				Name:   "show",
				Usage:  "Display current version",
				Action: showVersion(),
			},
		},
	}, nil

}
