package plugins

import (
	"testing"

	"github.com/indaco/semver-cli/internal/config"
	"github.com/indaco/semver-cli/internal/plugins/commitparser"
	"github.com/indaco/semver-cli/internal/plugins/tagmanager"
	"github.com/indaco/semver-cli/internal/plugins/versionvalidator"
)

func TestRegisterConfiguredPlugins_WithCommitParser(t *testing.T) {
	commitparser.ResetCommitParser()

	cfg := &config.Config{
		Plugins: &config.PluginConfig{
			CommitParser: true,
		},
	}

	RegisterBuiltinPlugins(cfg)

	p := commitparser.GetCommitParserFn()
	if p == nil {
		t.Fatal("expected commit parser to be registered, got nil")
	}

	if p.Name() != "commit-parser" {
		t.Errorf("expected name 'commit-parser', got %q", p.Name())
	}
}

func TestRegisterConfiguredPlugins_DisabledCommitParser(t *testing.T) {
	commitparser.ResetCommitParser()

	cfg := &config.Config{
		Plugins: &config.PluginConfig{
			CommitParser: false,
		},
	}

	RegisterBuiltinPlugins(cfg)

	if p := commitparser.GetCommitParserFn(); p != nil {
		t.Errorf("expected no commit parser to be registered, got %q", p.Name())
	}
}

func TestRegisterConfiguredPlugins_NilConfig(t *testing.T) {
	commitparser.ResetCommitParser()

	RegisterBuiltinPlugins(nil)

	if p := commitparser.GetCommitParserFn(); p != nil {
		t.Errorf("expected no commit parser to be registered, got %q", p.Name())
	}
}

func TestRegisterConfiguredPlugins_NilPluginsField(t *testing.T) {
	commitparser.ResetCommitParser()

	cfg := &config.Config{
		Plugins: nil, // explicit nil
	}

	RegisterBuiltinPlugins(cfg)

	if p := commitparser.GetCommitParserFn(); p != nil {
		t.Errorf("expected no commit parser to be registered, got %q", p.Name())
	}
}

func TestRegisterConfiguredPlugins_WithTagManager(t *testing.T) {
	tagmanager.ResetTagManager()
	defer tagmanager.ResetTagManager()

	autoCreate := true
	annotate := true
	cfg := &config.Config{
		Plugins: &config.PluginConfig{
			TagManager: &config.TagManagerConfig{
				Enabled:    true,
				AutoCreate: &autoCreate,
				Prefix:     "v",
				Annotate:   &annotate,
				Push:       false,
			},
		},
	}

	RegisterBuiltinPlugins(cfg)

	tm := tagmanager.GetTagManagerFn()
	if tm == nil {
		t.Fatal("expected tag manager to be registered, got nil")
	}

	if tm.Name() != "tag-manager" {
		t.Errorf("expected name 'tag-manager', got %q", tm.Name())
	}
}

func TestRegisterConfiguredPlugins_TagManagerDisabled(t *testing.T) {
	tagmanager.ResetTagManager()
	defer tagmanager.ResetTagManager()

	cfg := &config.Config{
		Plugins: &config.PluginConfig{
			TagManager: &config.TagManagerConfig{
				Enabled: false,
			},
		},
	}

	RegisterBuiltinPlugins(cfg)

	if tm := tagmanager.GetTagManagerFn(); tm != nil {
		t.Errorf("expected no tag manager to be registered when disabled")
	}
}

func TestRegisterConfiguredPlugins_TagManagerNil(t *testing.T) {
	tagmanager.ResetTagManager()
	defer tagmanager.ResetTagManager()

	cfg := &config.Config{
		Plugins: &config.PluginConfig{
			TagManager: nil,
		},
	}

	RegisterBuiltinPlugins(cfg)

	if tm := tagmanager.GetTagManagerFn(); tm != nil {
		t.Errorf("expected no tag manager to be registered when nil")
	}
}

func TestRegisterConfiguredPlugins_WithVersionValidator(t *testing.T) {
	versionvalidator.Unregister()
	defer versionvalidator.Unregister()

	cfg := &config.Config{
		Plugins: &config.PluginConfig{
			VersionValidator: &config.VersionValidatorConfig{
				Enabled: true,
				Rules: []config.ValidationRule{
					{Type: "major-version-max", Value: 10},
					{Type: "pre-release-format", Pattern: "^(alpha|beta)$"},
				},
			},
		},
	}

	RegisterBuiltinPlugins(cfg)

	vv := versionvalidator.GetVersionValidatorFn()
	if vv == nil {
		t.Fatal("expected version validator to be registered, got nil")
	}

	if vv.Name() != "version-validator" {
		t.Errorf("expected name 'version-validator', got %q", vv.Name())
	}
}

func TestRegisterConfiguredPlugins_VersionValidatorDisabled(t *testing.T) {
	versionvalidator.Unregister()
	defer versionvalidator.Unregister()

	cfg := &config.Config{
		Plugins: &config.PluginConfig{
			VersionValidator: &config.VersionValidatorConfig{
				Enabled: false,
			},
		},
	}

	RegisterBuiltinPlugins(cfg)

	if vv := versionvalidator.GetVersionValidatorFn(); vv != nil {
		t.Errorf("expected no version validator to be registered when disabled")
	}
}

func TestRegisterConfiguredPlugins_VersionValidatorNil(t *testing.T) {
	versionvalidator.Unregister()
	defer versionvalidator.Unregister()

	cfg := &config.Config{
		Plugins: &config.PluginConfig{
			VersionValidator: nil,
		},
	}

	RegisterBuiltinPlugins(cfg)

	if vv := versionvalidator.GetVersionValidatorFn(); vv != nil {
		t.Errorf("expected no version validator to be registered when nil")
	}
}

func TestConvertValidationRules(t *testing.T) {
	configRules := []config.ValidationRule{
		{Type: "major-version-max", Value: 10},
		{Type: "pre-release-format", Pattern: "^alpha$"},
		{Type: "branch-constraint", Branch: "release/*", Allowed: []string{"patch"}, Enabled: true},
	}

	rules := convertValidationRules(configRules)

	if len(rules) != 3 {
		t.Fatalf("expected 3 rules, got %d", len(rules))
	}

	if rules[0].Type != versionvalidator.RuleMajorVersionMax {
		t.Errorf("expected rule type 'major-version-max', got %q", rules[0].Type)
	}
	if rules[0].Value != 10 {
		t.Errorf("expected value 10, got %d", rules[0].Value)
	}

	if rules[1].Pattern != "^alpha$" {
		t.Errorf("expected pattern '^alpha$', got %q", rules[1].Pattern)
	}

	if rules[2].Branch != "release/*" {
		t.Errorf("expected branch 'release/*', got %q", rules[2].Branch)
	}
	if len(rules[2].Allowed) != 1 || rules[2].Allowed[0] != "patch" {
		t.Errorf("expected allowed [patch], got %v", rules[2].Allowed)
	}
}

func TestConvertValidationRules_Empty(t *testing.T) {
	rules := convertValidationRules(nil)

	if len(rules) != 0 {
		t.Errorf("expected 0 rules for nil input, got %d", len(rules))
	}

	rules = convertValidationRules([]config.ValidationRule{})

	if len(rules) != 0 {
		t.Errorf("expected 0 rules for empty input, got %d", len(rules))
	}
}

func TestRegisterConfiguredPlugins_AllPlugins(t *testing.T) {
	commitparser.ResetCommitParser()
	tagmanager.ResetTagManager()
	versionvalidator.Unregister()
	defer func() {
		commitparser.ResetCommitParser()
		tagmanager.ResetTagManager()
		versionvalidator.Unregister()
	}()

	autoCreate := true
	cfg := &config.Config{
		Plugins: &config.PluginConfig{
			CommitParser: true,
			TagManager: &config.TagManagerConfig{
				Enabled:    true,
				AutoCreate: &autoCreate,
			},
			VersionValidator: &config.VersionValidatorConfig{
				Enabled: true,
				Rules:   []config.ValidationRule{{Type: "major-version-max", Value: 5}},
			},
		},
	}

	RegisterBuiltinPlugins(cfg)

	if p := commitparser.GetCommitParserFn(); p == nil {
		t.Error("expected commit parser to be registered")
	}
	if tm := tagmanager.GetTagManagerFn(); tm == nil {
		t.Error("expected tag manager to be registered")
	}
	if vv := versionvalidator.GetVersionValidatorFn(); vv == nil {
		t.Error("expected version validator to be registered")
	}
}
