package modulescmd

import (
	"context"
	"fmt"
	"os"

	"github.com/indaco/semver-cli/internal/config"
	"github.com/indaco/semver-cli/internal/core"
	"github.com/indaco/semver-cli/internal/workspace"
	"github.com/urfave/cli/v3"
)

// discoverCmd returns the "discover" subcommand for testing discovery settings.
func discoverCmd() *cli.Command {
	return &cli.Command{
		Name:  "discover",
		Usage: "Test module discovery without running operations",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "dry-run",
				Usage: "Show what would be discovered",
				Value: true,
			},
		},
		Action: runDiscover,
	}
}

func runDiscover(ctx context.Context, cmd *cli.Command) error {
	cfg, err := config.LoadConfigFn()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	if cfg == nil {
		cfg = &config.Config{}
	}

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	fs := core.NewOSFileSystem()
	detector := workspace.NewDetector(fs, cfg)

	// Show discovery settings
	discovery := cfg.GetDiscoveryConfig()
	fmt.Println("Discovery settings:")
	fmt.Printf("  Enabled: %v\n", *discovery.Enabled)
	fmt.Printf("  Recursive: %v\n", *discovery.Recursive)
	fmt.Printf("  Max depth: %d\n", *discovery.MaxDepth)
	fmt.Printf("  Exclude patterns: %v\n", cfg.GetExcludePatterns())
	fmt.Println()

	// Detect context
	detectedCtx, err := detector.DetectContext(cwd)
	if err != nil {
		return fmt.Errorf("failed to detect context: %w", err)
	}

	fmt.Printf("Detection mode: %s\n", detectedCtx.Mode)
	fmt.Println()

	switch detectedCtx.Mode {
	case workspace.SingleModule:
		fmt.Printf("Single module found:\n")
		fmt.Printf("  Path: %s\n", detectedCtx.Path)

	case workspace.MultiModule:
		fmt.Printf("Multiple modules found (%d):\n", len(detectedCtx.Modules))
		for _, mod := range detectedCtx.Modules {
			version := mod.CurrentVersion
			if version == "" {
				version = "unknown"
			}
			fmt.Printf("  - %s (%s)\n", mod.Name, version)
			fmt.Printf("    Path: %s\n", mod.RelPath)
		}

	case workspace.NoModules:
		fmt.Println("No modules found in workspace")
	}

	return nil
}
