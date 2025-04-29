package plugins

import (
	"github.com/indaco/semver-cli/internal/config"
	commitparser "github.com/indaco/semver-cli/internal/plugins/commit-parser"
)

func RegisterBuiltinPlugins(cfg *config.Config) {
	if cfg == nil || cfg.Plugins == nil {
		return
	}

	if cfg.Plugins.CommitParser {
		commitparser.Register()
	}
}
