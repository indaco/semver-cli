package main

import (
	"fmt"

	"github.com/indaco/semver-cli/internal/config"
	"github.com/indaco/semver-cli/internal/version"
	"github.com/urfave/cli/v3"
)

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
						Name:  "next",
						Usage: "Smart bump logic (e.g. promote pre-release or bump patch)",
						UsageText: `semver bump next [--label patch|minor|major] [--meta data] [--preserve-meta] [--since ref] [--until ref] [--no-infer]

By default, semver tries to infer the bump type from recent commit messages using the built-in commit-parser plugin.
You can override this behavior with the --label flag, disable it explicitly with --no-infer, or disable the plugin via the config file (.semver.yaml).`,
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:  "label",
								Usage: "Optional bump label override (patch, minor, major)",
							},
							&cli.StringFlag{
								Name:  "meta",
								Usage: "Set build metadata (e.g. 'ci.123')",
							},
							&cli.BoolFlag{
								Name:  "preserve-meta",
								Usage: "Preserve existing build metadata instead of clearing it",
							},
							&cli.StringFlag{
								Name:  "since",
								Usage: "Start commit/tag for bump inference (default: last tag or HEAD~10)",
							},
							&cli.StringFlag{
								Name:  "until",
								Usage: "End commit/tag for bump inference (default: HEAD)",
							},
							&cli.BoolFlag{
								Name:  "no-infer",
								Usage: "Disable bump inference from commit messages (overrides config)",
							},
						},
						Action: bumpNextCmd(cfg),
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
			{
				Name:  "extension",
				Usage: "Manage extensions for semver-cli",
				Commands: []*cli.Command{
					{
						Name:  "install",
						Usage: "Install an extension from a remote repo or local path",
						Flags: []cli.Flag{
							&cli.StringFlag{Name: "url", Usage: "Git URL to clone"},
							&cli.StringFlag{Name: "path", Usage: "Local path to copy from"},
							&cli.StringFlag{Name: "extension-dir", Usage: "Directory to store extensions in", Value: "."},
						},
						Action: extensionInstallCmd(),
					},
					{
						Name:   "list",
						Usage:  "List installed extensions",
						Action: extensionListCmd(),
					},
					{
						Name:  "remove",
						Usage: "Remove a registered extension",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:  "name",
								Usage: "Name of the extension to remove",
							},
							&cli.BoolFlag{
								Name:  "delete-folder",
								Usage: "Delete the extension directory from the .semver-extensions folder",
							},
						},
						Action: extensionRemoveCmd(),
					},
				},
			},
		},
	}
}
