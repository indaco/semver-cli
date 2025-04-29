package extensioncmd

import (
	"context"
	"fmt"

	"github.com/indaco/semver-cli/api/v0/extensions"
	"github.com/indaco/semver-cli/internal/config"
	"github.com/urfave/cli/v3"
)

// listCmd returns the "list" subcommand.
func listCmd() *cli.Command {
	return &cli.Command{
		Name:  "list",
		Usage: "List installed extensions",
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return runExtenstionList()
		},
	}
}

// runExtenstionList lists installed extensions.
func runExtenstionList() error {
	// Load the configuration file
	cfg, err := config.LoadConfigFn()
	if err != nil {
		// Print the error to stdout and return
		fmt.Println("failed to load configuration:", err)
		return nil
	}

	// If there are no plugins, notify the user
	if len(cfg.Extensions) == 0 {
		fmt.Println("No extensions registered.")
		return nil
	}

	// Create a lookup map of metadata
	metadataMap := map[string]extensions.Extension{}
	for _, meta := range extensions.AllExtensions() {
		metadataMap[meta.Name()] = meta
	}

	fmt.Println("List of Registered Extensions:")
	fmt.Println()
	fmt.Println("  NAME              VERSION     ENABLED   DESCRIPTION")
	fmt.Println("  ----------------------------------------------------------")

	for _, p := range cfg.Extensions {
		meta, ok := metadataMap[p.Name]
		version := "?"
		desc := "(no metadata)"
		if ok {
			version = meta.Version()
			desc = meta.Description()
		}

		fmt.Printf("  %-17s %-10s %-9v %s\n", p.Name, version, p.Enabled, desc)
	}

	fmt.Println()

	return nil
}
