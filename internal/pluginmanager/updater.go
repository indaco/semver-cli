package pluginmanager

import (
	"os"

	"github.com/goccy/go-yaml"
	"github.com/indaco/semver-cli/internal/config"
)

var (
	marshalFunc         = yaml.Marshal
	AddPluginToConfigFn = AddPluginToConfig
)

// AddPluginToConfig appends a plugin entry to the YAML config at the given path.
// It avoids duplicates and preserves existing fields.
func AddPluginToConfig(path string, plugin config.PluginConfig) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var cfg config.Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return err
	}

	// Avoid duplicates
	for _, p := range cfg.Plugins {
		if p.Name == plugin.Name {
			return nil
		}
	}

	cfg.Plugins = append(cfg.Plugins, plugin)

	out, err := marshalFunc(cfg)
	if err != nil {
		return err
	}

	return os.WriteFile(path, out, 0644)
}
