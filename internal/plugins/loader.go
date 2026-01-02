package plugins

import (
	"github.com/indaco/semver-cli/internal/config"
	"github.com/indaco/semver-cli/internal/plugins/commitparser"
	"github.com/indaco/semver-cli/internal/plugins/tagmanager"
	"github.com/indaco/semver-cli/internal/plugins/versionvalidator"
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

	if cfg.Plugins.VersionValidator != nil && cfg.Plugins.VersionValidator.Enabled {
		vvCfg := &versionvalidator.Config{
			Enabled: true,
			Rules:   convertValidationRules(cfg.Plugins.VersionValidator.Rules),
		}
		versionvalidator.Register(vvCfg)
	}
}

// convertValidationRules converts config rules to versionvalidator rules.
func convertValidationRules(configRules []config.ValidationRule) []versionvalidator.Rule {
	rules := make([]versionvalidator.Rule, len(configRules))
	for i, r := range configRules {
		rules[i] = versionvalidator.Rule{
			Type:    versionvalidator.RuleType(r.Type),
			Pattern: r.Pattern,
			Value:   r.Value,
			Enabled: r.Enabled,
			Branch:  r.Branch,
			Allowed: r.Allowed,
		}
	}
	return rules
}
