package main

import (
	"fmt"

	"github.com/indaco/semver-cli/internal/version"
	"github.com/urfave/cli/v3"
)

// newCLI builds and returns the root CLI command,
// configuring all subcommands and flags for the semver tool.
func newCLI(defaultPath string) *cli.Command {
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
				Name:      "init",
				Usage:     "Initialize a .version file (auto-detects Git tag or starts from 0.1.0)",
				UsageText: "semver init",
				Action:    initVersion(),
			},
			{
				Name:      "patch",
				Usage:     "Increment patch version",
				UsageText: "semver patch",
				Action:    bumpPatch(),
			},
			{
				Name:      "minor",
				Usage:     "Increment minor version and reset patch",
				UsageText: "semver minor",
				Action:    bumpMinor(),
			},
			{
				Name:      "major",
				Usage:     "Increment major version and reset minor and patch",
				UsageText: "semver major",
				Action:    bumpMajor(),
			},
			{
				Name:      "pre",
				Usage:     "Set pre-release label (e.g., alpha, beta.1)",
				UsageText: "semver pre --label <label> [--inc]",
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
				Name:      "show",
				Usage:     "Display current version",
				UsageText: "semver show",
				Action:    showVersion(),
			},
			{
				Name:      "set",
				Usage:     "Set the version manually",
				UsageText: "semver set <version> [--pre label]",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "pre",
						Usage: "Optional pre-release label",
					},
				},
				Action: setVersion(),
			},
		},
	}
}
