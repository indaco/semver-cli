package plugins

import (
	"github.com/indaco/semver-cli/api/plugins"
	"github.com/indaco/semver-cli/internal/config"
)

// Factory is a function that returns an instance that implements one or more plugin interfaces.
type Factory func() any

var factories = map[string]Factory{}

// RegisterFactory registers a plugin factory under a given name.
func RegisterFactory(name string, fn Factory) {
	factories[name] = fn
}

// RegisterConfiguredPlugins loads and registers plugins declared in the .semver.yaml file.
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
			continue
		}

		instance := factory()

		// Register metadata
		if meta, ok := instance.(plugins.Plugin); ok {
			plugins.RegisterPlugin(meta)
		}

		// Register capabilities
		if cp, ok := instance.(plugins.CommitParser); ok {
			plugins.RegisterCommitParser(cp)
		}

		// future:
		// if cg, ok := instance.(plugins.ChangelogGenerator); ok {
		//     plugins.RegisterChangelogGenerator(cg)
		// }
	}
}
