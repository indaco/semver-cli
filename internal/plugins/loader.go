package plugins

import (
	"github.com/indaco/semver-cli/internal/config"
	"github.com/indaco/semver-cli/internal/plugins/commitparser"
	"github.com/indaco/semver-cli/internal/plugins/tagmanager"
)

func RegisterBuiltinPlugins(cfg *config.Config) {
	if cfg == nil || cfg.Plugins == nil {
		return
	}

	if cfg.Plugins.CommitParser {
		commitparser.Register()
	}

	if cfg.Plugins.TagManager != nil && cfg.Plugins.TagManager.Enabled {
		tmCfg := &tagmanager.Config{
			Enabled:    true,
			AutoCreate: cfg.Plugins.TagManager.GetAutoCreate(),
			Prefix:     cfg.Plugins.TagManager.GetPrefix(),
			Annotate:   cfg.Plugins.TagManager.GetAnnotate(),
			Push:       cfg.Plugins.TagManager.Push,
		}
		tagmanager.Register(tmCfg)
	}
}
