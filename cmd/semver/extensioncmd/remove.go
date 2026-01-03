package extensioncmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/indaco/semver-cli/internal/config"
	"github.com/urfave/cli/v3"
)

// listCmd returns the "list" subcommand.
func removeCmd() *cli.Command {
	return &cli.Command{
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
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return runExtenstionRemove(cmd)
		},
	}
}

// runExtenstionRemove removes an installed extension.
func runExtenstionRemove(cmd *cli.Command) error {
	extensionName := cmd.String("name")
	if extensionName == "" {
		return fmt.Errorf("please provide an extension name to remove")
	}

	cfg, err := config.LoadConfigFn()
	if err != nil {
		fmt.Println("failed to load configuration:", err)
		return nil
	}

	var extensionToRemove *config.ExtensionConfig
	for i, extension := range cfg.Extensions {
		if extension.Name == extensionName {
			extensionToRemove = &cfg.Extensions[i]
			break
		}
	}

	if extensionToRemove == nil {
		fmt.Printf("extension %q not found\n", extensionName)
		return nil
	}

	// Disable the plugin in the configuration (set Enabled to false)
	extensionToRemove.Enabled = false

	// Save the updated config back to the file
	if err := config.SaveConfigFn(cfg); err != nil {
		fmt.Println("failed to save updated configuration:", err)
		return nil
	}

	// Check if --delete-folder flag is set to remove the extension folder
	isDeleteFolder := cmd.Bool("delete-folder")
	if isDeleteFolder {
		// Remove the extension directory from ".semver-extensions"
		extensionDir := filepath.Join(".semver-extensions", extensionName)
		if err := os.RemoveAll(extensionDir); err != nil {
			return fmt.Errorf("failed to remove extension directory: %w", err)
		}
		fmt.Printf("Extension %q and its directory removed successfully.\n", extensionName)
	} else {
		fmt.Printf("Extension %q removed, but its directory is preserved.\n", extensionName)
	}

	return nil
}
