package extensioncmd

import (
	"context"

	extensionmanager "github.com/indaco/semver-cli/internal/extension-manager"
	"github.com/urfave/cli/v3"
)

// installCmd returns the "install" subcommand.
func installCmd() *cli.Command {
	return &cli.Command{
		Name:  "install",
		Usage: "Install an extension from a remote repo or local path",
		MutuallyExclusiveFlags: []cli.MutuallyExclusiveFlags{
			{},
		},
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "url", Usage: "Git URL to clone"},
			&cli.StringFlag{Name: "path", Usage: "Local path to copy from"},
			&cli.StringFlag{Name: "extension-dir", Usage: "Directory to store extensions in", Value: "."},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return runExtenstionInstall(cmd)
		},
	}
}

// runExtenstionInstall installs an extension from local or remote.
func runExtenstionInstall(cmd *cli.Command) error {
	localPath := cmd.String("path")
	if localPath == "" {
		return cli.Exit("missing --path (or --url) for extension registration", 1)
	}

	// Get the extension directory (use the provided flag or default to current directory)
	extensionDirectory := cmd.String("extension-dir")

	// Proceed with normal extension registration
	return extensionmanager.RegisterLocalExtensionFn(localPath, ".semver.yaml", extensionDirectory)
}
