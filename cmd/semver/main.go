package main

import (
	"context"
	"log"
	"os"

	"github.com/indaco/semver-cli/internal/config"
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
		defaultPath = cfg.Path
	}

	app := newCLI(defaultPath)
	return app.Run(context.Background(), args)
}
