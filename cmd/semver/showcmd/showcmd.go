package showcmd

import (
	"context"
	"fmt"

	"github.com/indaco/semver-cli/internal/clix"
	"github.com/indaco/semver-cli/internal/semver"
	"github.com/urfave/cli/v3"
)

// Run returns the "pre" command.
func Run() *cli.Command {
	return &cli.Command{
		Name:      "show",
		Usage:     "Display current version",
		UsageText: "semver show",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return runShowCmd(cmd)
		},
	}
}

// runShowCmd prints the current version.
func runShowCmd(cmd *cli.Command) error {
	if _, err := clix.FromCommandFn(cmd); err != nil {
		return err
	}

	path := cmd.String("path")
	version, err := semver.ReadVersion(path)
	if err != nil {
		return fmt.Errorf("failed to read version file at %s: %w", path, err)
	}

	fmt.Println(version.String())
	return nil
}
