package doctorcmd

import (
	"context"
	"fmt"

	"github.com/indaco/semver-cli/internal/semver"
	"github.com/urfave/cli/v3"
)

// Run returns the "pre" command.
func Run() *cli.Command {
	return &cli.Command{
		Name:    "doctor",
		Aliases: []string{"validate"},
		Usage:   "Validate the .version file",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return runDoctorCmd(cmd)
		},
	}
}

// runDoctorCmd checks that the .version file is valid.
func runDoctorCmd(cmd *cli.Command) error {
	path := cmd.String("path")
	_, err := semver.ReadVersion(path)
	if err != nil {
		return fmt.Errorf("invalid version file at %s: %w", path, err)
	}
	fmt.Printf("Valid version file at %s\n", path)
	return nil
}
