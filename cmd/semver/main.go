package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/indaco/semver-cli/internal/config"
	"github.com/indaco/semver-cli/internal/hooks"
	"github.com/indaco/semver-cli/internal/plugins"
)

func main() {
	if err := runCLI(os.Args); err != nil {
		log.Fatal(err)
	}
}

func runCLI(args []string) error {
	cfg, err := config.LoadConfigFn()
	if err != nil {
		return err
	}
	if cfg == nil {
		cfg = &config.Config{}
	}

	// Normalize or fallback
	cfg.Path = config.NormalizeVersionPath(cfg.Path)
	if cfg.Path == "" {
		cfg.Path = ".version"
	}

	plugins.RegisterBuiltinPlugins(cfg)

	if err := hooks.LoadPreReleaseHooksFromConfig(cfg); err != nil {
		return fmt.Errorf("failed to load pre-release hooks: %w", err)
	}

	app := newCLI(cfg)
	return app.Run(context.Background(), args)
}
