package plugins

import (
	"testing"

	"github.com/indaco/semver-cli/internal/config"
	"github.com/indaco/semver-cli/internal/plugins/commitparser"
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
