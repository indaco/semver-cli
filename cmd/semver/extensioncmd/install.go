package extensioncmd

import (
	"context"
	"fmt"

	"github.com/indaco/semver-cli/internal/extensionmgr"
	"github.com/urfave/cli/v3"
)

// installCmd returns the "install" subcommand.
func installCmd() *cli.Command {
	return &cli.Command{
		Name:  "install",
		Usage: "Install an extension from a remote repo or local path",
		Description: `Install an extension from a local path or remote Git repository.

Supported URL formats:
  - https://github.com/user/repo
  - https://gitlab.com/user/repo
  - github.com/user/repo
  - gitlab.com/user/repo

Examples:
  semver extension install --path ./my-extension
  semver extension install --url https://github.com/user/semver-ext-changelog
  semver extension install --url github.com/user/semver-ext-notify`,
		MutuallyExclusiveFlags: []cli.MutuallyExclusiveFlags{
			{
				Flags: [][]cli.Flag{
					{
						&cli.StringFlag{Name: "url", Usage: "Git URL to clone"},
						&cli.StringFlag{Name: "path", Usage: "Local path to copy from"},
					},
				},
			},
		},
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "extension-dir", Usage: "Directory to store extensions in", Value: "."},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return runExtensionInstall(cmd)
		},
	}
}

// runExtensionInstall installs an extension from local or remote.
func runExtensionInstall(cmd *cli.Command) error {
	localPath := cmd.String("path")
	urlStr := cmd.String("url")

	// Check that at least one source is provided
	if localPath == "" && urlStr == "" {
		return cli.Exit("missing --path or --url for extension installation", 1)
	}

	// Get the extension directory (use the provided flag or default to current directory)
	extensionDirectory := cmd.String("extension-dir")

	// Handle URL-based installation
	if urlStr != "" {
		// Validate git is available
		if err := extensionmgr.ValidateGitAvailable(); err != nil {
			return cli.Exit(fmt.Sprintf("Error: %v", err), 1)
		}

		// Install from URL
		if err := extensionmgr.InstallFromURL(urlStr, ".semver.yaml", extensionDirectory); err != nil {
			return cli.Exit(fmt.Sprintf("Failed to install extension from URL: %v", err), 1)
		}
		return nil
	}

	// Handle local path installation
	if localPath != "" {
		// Auto-detect if the path looks like a URL
		if extensionmgr.IsURL(localPath) {
			// Validate git is available
			if err := extensionmgr.ValidateGitAvailable(); err != nil {
				return cli.Exit(fmt.Sprintf("Error: %v", err), 1)
			}

			// Install from URL
			if err := extensionmgr.InstallFromURL(localPath, ".semver.yaml", extensionDirectory); err != nil {
				return cli.Exit(fmt.Sprintf("Failed to install extension from URL: %v", err), 1)
			}
			return nil
		}

		// Proceed with normal extension registration from local path
		if err := extensionmgr.RegisterLocalExtensionFn(localPath, ".semver.yaml", extensionDirectory); err != nil {
			return cli.Exit(fmt.Sprintf("Failed to install extension: %v", err), 1)
		}
		return nil
	}

	return cli.Exit("no installation source provided", 1)
}
