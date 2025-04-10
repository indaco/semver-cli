package main

import (
	"fmt"

	"github.com/indaco/semver-cli/internal/version"
	"github.com/urfave/cli/v3"
)

// newCLI builds and returns the root CLI command,
// configuring all subcommands and flags for the semver cli.
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
			&cli.BoolFlag{
				Name:  "no-auto-init",
				Usage: "Disable auto-initialization of the .version file",
			},
		},
		Commands: []*cli.Command{
			{
				Name:      "show",
				Usage:     "Display current version",
				UsageText: "semver show",
				Action:    showVersionCmd(),
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
					&cli.StringFlag{
						Name:  "meta",
						Usage: "Optional build metadata",
					},
				},
				Action: setVersionCmd(),
			},
			{
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
					{
						Name:      "patch",
						Usage:     "Increment patch version",
						UsageText: "semver bump patch [--pre label] [--meta data] [--preserve-meta]",
						Action:    bumpPatchCmd(),
					},
					{
						Name:      "minor",
						Usage:     "Increment minor version and reset patch",
						UsageText: "semver bump minor [--pre label] [--meta data] [--preserve-meta]",
						Action:    bumpMinorCmd(),
					},
					{
						Name:      "major",
						Usage:     "Increment major version and reset minor and patch",
						UsageText: "semver bump major [--pre label] [--meta data] [--preserve-meta]",
						Action:    bumpMajorCmd(),
					},
					{
						Name:      "release",
						Usage:     "Promote pre-release to final version (e.g. 1.2.3-alpha → 1.2.3)",
						UsageText: "semver bump release [--preserve-meta]",
						Action:    bumpReleaseCmd(),
					},
					{
						Name:      "next",
						Usage:     "Smart bump to the next version based on current state",
						UsageText: "semver bump next [--preserve-meta]",
						Action:    bumpNextCmd(),
					},
				},
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
				Action: setPreReleaseCmd(),
			},
			{
				Name:   "validate",
				Usage:  "Validate the .version file",
				Action: validateVersionCmd(),
			},
			{
				Name:      "init",
				Usage:     "Initialize a .version file (auto-detects Git tag or starts from 0.1.0)",
				UsageText: "semver init",
				Action:    initVersionCmd(),
			},
		},
	}
}
