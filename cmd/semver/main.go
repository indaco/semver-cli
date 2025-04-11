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
	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}

	defaultPath := ".version"
	if cfg != nil && cfg.Path != "" {
		defaultPath = config.NormalizeVersionPath(cfg.Path)
	}

	app := newCLI(defaultPath)

	// Register statically loaded plugins
	for _, plugin := range plugins.All() {
		plugin.Register(app)
	}

	return app.Run(context.Background(), args)
}
