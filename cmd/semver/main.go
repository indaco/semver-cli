package main

import (
	"context"
	"log"
	"os"

	"github.com/indaco/semver-cli/internal/config"
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

	app := newCLI(cfg)
	return app.Run(context.Background(), args)
}
