package extensioncmd

import (
	"github.com/urfave/cli/v3"
)

// Run returns the "pre" command.
func Run() *cli.Command {
	return &cli.Command{
		Name:  "extension",
		Usage: "Manage extensions for semver-cli",
		Commands: []*cli.Command{
			installCmd(),
			listCmd(),
			removeCmd(),
		},
	}
}
