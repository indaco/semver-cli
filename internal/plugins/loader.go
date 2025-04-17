package plugins

import (
	"github.com/indaco/semver-cli/api"
	"github.com/indaco/semver-cli/internal/config"
)

// Factory is a function that returns a new Plugin instance.
type Factory func() api.Plugin

var factories = map[string]Factory{}

// RegisterFactory registers a plugin factory under a given name.
func RegisterFactory(name string, fn Factory) {
	factories[name] = fn
}

func RegisterConfiguredPlugins(cfg *config.Config) {
	if cfg == nil {
		return
	}

	for _, pluginCfg := range cfg.Plugins {
		if !pluginCfg.Enabled {
			continue
		}

		factory, ok := factories[pluginCfg.Name]
		if !ok {
			// Optionally: fmt.Printf("Warning: plugin %q not found\n", pluginCfg.Name)
			continue
		}

		Register(factory())
	}
}
