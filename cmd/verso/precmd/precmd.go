package precmd

import (
	"context"
	"fmt"

	"github.com/indaco/verso/internal/clix"
	"github.com/indaco/verso/internal/semver"
	"github.com/urfave/cli/v3"
)

// Run returns the "pre" command.
func Run() *cli.Command {
	return &cli.Command{
		Name:      "pre",
		Usage:     "Set pre-release label (e.g., alpha, beta.1)",
		UsageText: "verso pre --label <label> [--inc]",
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
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return runPreCmd(cmd)
		},
	}
}

// runPreCmd sets or increments the pre-release label.
func runPreCmd(cmd *cli.Command) error {
	path := cmd.String("path")
	label := cmd.String("label")
	isInc := cmd.Bool("inc")

	if _, err := clix.FromCommandFn(cmd); err != nil {
		return err
	}

	version, err := semver.ReadVersion(path)
	if err != nil {
		return fmt.Errorf("failed to read version: %w", err)
	}

	if isInc {
		version.PreRelease = semver.IncrementPreRelease(version.PreRelease, label)
	} else {
		if version.PreRelease == "" {
			version.Patch++
		}
		version.PreRelease = label
	}

	if err := semver.SaveVersion(path, version); err != nil {
		return fmt.Errorf("failed to save version: %w", err)
	}

	return nil
}
