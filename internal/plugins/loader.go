package plugins

import (
	"github.com/indaco/semver-cli/internal/config"
	"github.com/indaco/semver-cli/internal/plugins/changeloggenerator"
	"github.com/indaco/semver-cli/internal/plugins/changelogparser"
	"github.com/indaco/semver-cli/internal/plugins/commitparser"
	"github.com/indaco/semver-cli/internal/plugins/dependencycheck"
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

	if cfg.Plugins.DependencyCheck != nil && cfg.Plugins.DependencyCheck.Enabled {
		dcCfg := convertDependencyCheckConfig(cfg.Plugins.DependencyCheck)
		dependencycheck.Register(dcCfg)
	}

	if cfg.Plugins.ChangelogParser != nil && cfg.Plugins.ChangelogParser.Enabled {
		clCfg := convertChangelogParserConfig(cfg.Plugins.ChangelogParser)
		changelogparser.Register(clCfg)
	}

	if cfg.Plugins.ChangelogGenerator != nil && cfg.Plugins.ChangelogGenerator.Enabled {
		changeloggenerator.Register(cfg.Plugins.ChangelogGenerator)
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

// convertDependencyCheckConfig converts config to dependencycheck config.
func convertDependencyCheckConfig(cfg *config.DependencyCheckConfig) *dependencycheck.Config {
	files := make([]dependencycheck.FileConfig, len(cfg.Files))
	for i, f := range cfg.Files {
		files[i] = dependencycheck.FileConfig{
			Path:    f.Path,
			Field:   f.Field,
			Format:  f.Format,
			Pattern: f.Pattern,
		}
	}
	return &dependencycheck.Config{
		Enabled:  cfg.Enabled,
		AutoSync: cfg.AutoSync,
		Files:    files,
	}
}

// convertChangelogParserConfig converts config to changelogparser config.
func convertChangelogParserConfig(cfg *config.ChangelogParserConfig) *changelogparser.Config {
	return &changelogparser.Config{
		Enabled:                  cfg.Enabled,
		Path:                     cfg.Path,
		RequireUnreleasedSection: cfg.RequireUnreleasedSection,
		InferBumpType:            cfg.InferBumpType,
		Priority:                 cfg.Priority,
	}
}
