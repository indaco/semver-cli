package setcmd

import (
	"context"
	"fmt"

	"github.com/indaco/semver-cli/internal/semver"
	"github.com/urfave/cli/v3"
)

// Run returns the "set" command.
func Run() *cli.Command {
	return &cli.Command{
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
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return runSetCmd(cmd)
		},
	}
}

// runSetCmd manually sets the version.
func runSetCmd(cmd *cli.Command) error {
	path := cmd.String("path")
	args := cmd.Args()

	if args.Len() < 1 {
		return cli.Exit("missing required version argument", 1)
	}

	raw := args.Get(0)
	pre := cmd.String("pre")
	meta := cmd.String("meta")

	version, err := semver.ParseVersion(raw)
	if err != nil {
		return fmt.Errorf("invalid version: %w", err)
	}
	version.PreRelease = pre
	version.Build = meta

	if err := semver.SaveVersion(path, version); err != nil {
		return fmt.Errorf("failed to save version: %w", err)
	}

	fmt.Printf("Set version to %s in %s\n", version.String(), path)
	return nil
}
