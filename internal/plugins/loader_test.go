package plugins

import (
	"errors"
	"testing"

	"github.com/indaco/semver-cli/api/v0/plugins"
	"github.com/indaco/semver-cli/internal/config"
)

/* ------------------------------------------------------------------------- */
/* MOCK PLUGINS                                                              */
/* ------------------------------------------------------------------------- */

// mockPlugin implements both metadata and CommitParser capability
type mockPlugin struct {
	metadataName string
}

func (m mockPlugin) Name() string        { return m.metadataName }
func (m mockPlugin) Description() string { return "mock description" }
func (m mockPlugin) Version() string     { return "v0.0.1" }

// CommitParser capability
func (m mockPlugin) Parse(commits []string) (string, error) {
	if len(commits) == 0 {
		return "", errors.New("no commits")
	}
	return "patch", nil
}

/* ------------------------------------------------------------------------- */
/* TESTS                                                                     */
/* ------------------------------------------------------------------------- */

func TestRegisterConfiguredPlugins_WithMetadataAndCommitParser(t *testing.T) {
	resetPluginSystem()

	RegisterFactory("mock", func() any {
		return mockPlugin{metadataName: "mock"}
	})

	cfg := &config.Config{
		Plugins: []config.PluginConfig{
			{Name: "mock", Enabled: true},
		},
	}

	RegisterConfiguredPlugins(cfg)

	// Check metadata
	meta := plugins.AllPlugins()
	if len(meta) != 1 {
		t.Fatalf("expected 1 plugin metadata, got %d", len(meta))
	}
	if meta[0].Name() != "mock" {
		t.Errorf("expected name 'mock', got %q", meta[0].Name())
	}

	// Check capabilities
	parser := plugins.GetCommitParser()
	if parser == nil {
		t.Fatalf("expected 1 commit parser, got 0")
	}
	if parser.Name() != "mock" {
		t.Errorf("expected parser name 'mock', got %q", parser.Name())
	}
}

func TestRegisterConfiguredPlugins_UnknownPlugin(t *testing.T) {
	resetPluginSystem()

	cfg := &config.Config{
		Plugins: []config.PluginConfig{
			{Name: "unknown", Enabled: true},
		},
	}

	RegisterConfiguredPlugins(cfg)

	if got := len(plugins.AllPlugins()); got != 0 {
		t.Errorf("expected no plugin metadata registered, got %d", got)
	}
}

func TestRegisterConfiguredPlugins_DisabledPlugin(t *testing.T) {
	resetPluginSystem()

	RegisterFactory("mock", func() any {
		return mockPlugin{metadataName: "mock"}
	})

	cfg := &config.Config{
		Plugins: []config.PluginConfig{
			{Name: "mock", Enabled: false},
		},
	}

	RegisterConfiguredPlugins(cfg)

	if got := len(plugins.AllPlugins()); got != 0 {
		t.Errorf("expected no plugin metadata registered, got %d", got)
	}
}

func TestRegisterConfiguredPlugins_NilConfig(t *testing.T) {
	resetPluginSystem()

	RegisterConfiguredPlugins(nil)

	if got := len(plugins.AllPlugins()); got != 0 {
		t.Errorf("expected no plugin metadata registered, got %d", got)
	}
}

/* ------------------------------------------------------------------------- */
/* HELPERS                                                                   */
/* ------------------------------------------------------------------------- */

func resetPluginSystem() {
	factories = map[string]Factory{}
	plugins.ResetCommitParser()
	plugins.ResetPlugin()
}
