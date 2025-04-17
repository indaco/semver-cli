package plugins

import (
	"testing"

	"github.com/indaco/semver-cli/api/plugins"
	"github.com/indaco/semver-cli/internal/config"
	"github.com/urfave/cli/v3"
)

type dummyPlugin struct {
	pluginName string
	registered bool
}

func (d *dummyPlugin) Name() string {
	return d.pluginName
}

func (d *dummyPlugin) Register(cmd *cli.Command) {
	d.registered = true
	cmd.Description = "dummy registered"
}

func newDummyPlugin(name string) plugins.Plugin {
	return &dummyPlugin{pluginName: name}
}

func TestRegisterConfiguredPlugins(t *testing.T) {
	resetPluginSystem()

	RegisterFactory("dummy", func() plugins.Plugin {
		return newDummyPlugin("dummy")
	})

	cfg := &config.Config{
		Plugins: []config.PluginConfig{
			{Name: "dummy", Enabled: true},
		},
	}

	RegisterConfiguredPlugins(cfg)

	all := plugins.All()
	if len(all) != 1 {
		t.Fatalf("expected 1 plugin, got %d", len(all))
	}
	if all[0].Name() != "dummy" {
		t.Errorf("expected 'dummy', got %q", all[0].Name())
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

	if got := len(plugins.All()); got != 0 {
		t.Errorf("expected no plugin registered, got %d", got)
	}
}

func TestRegisterConfiguredPlugins_DisabledPlugin(t *testing.T) {
	resetPluginSystem()

	RegisterFactory("dummy", func() plugins.Plugin {
		return newDummyPlugin("dummy")
	})

	cfg := &config.Config{
		Plugins: []config.PluginConfig{
			{Name: "dummy", Enabled: false},
		},
	}

	RegisterConfiguredPlugins(cfg)

	if got := len(plugins.All()); got != 0 {
		t.Errorf("expected no plugin registered, got %d", got)
	}
}

func TestRegisterConfiguredPlugins_NilConfig(t *testing.T) {
	// Reset registry before and after test
	plugins.ResetPlugins()
	defer plugins.ResetPlugins()

	RegisterConfiguredPlugins(nil)

	if len(plugins.All()) != 0 {
		t.Errorf("expected no plugins registered, got %d", len(plugins.All()))
	}
}

/* ------------------------------------------------------------------------- */
/* HELPERS                                                                   */
/* ------------------------------------------------------------------------- */

func resetPluginSystem() {
	plugins.ResetPlugins()
	factories = map[string]Factory{}
}
