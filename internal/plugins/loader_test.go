package plugins

import (
	"testing"

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

func newDummyPlugin(name string) Plugin {
	return &dummyPlugin{pluginName: name}
}

func TestRegisterConfiguredPlugins(t *testing.T) {
	resetPluginSystem()

	RegisterFactory("dummy", func() Plugin {
		return newDummyPlugin("dummy")
	})

	cfg := &config.Config{
		Plugins: []config.PluginConfig{
			{Name: "dummy", Enabled: true},
		},
	}

	RegisterConfiguredPlugins(cfg)

	all := All()
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

	if got := len(All()); got != 0 {
		t.Errorf("expected no plugin registered, got %d", got)
	}
}

func TestRegisterConfiguredPlugins_DisabledPlugin(t *testing.T) {
	resetPluginSystem()

	RegisterFactory("dummy", func() Plugin {
		return newDummyPlugin("dummy")
	})

	cfg := &config.Config{
		Plugins: []config.PluginConfig{
			{Name: "dummy", Enabled: false},
		},
	}

	RegisterConfiguredPlugins(cfg)

	if got := len(All()); got != 0 {
		t.Errorf("expected no plugin registered, got %d", got)
	}
}

func TestRegisterConfiguredPlugins_NilConfig(t *testing.T) {
	// Reset registry before and after test
	resetPlugins()
	defer resetPlugins()

	RegisterConfiguredPlugins(nil)

	if len(All()) != 0 {
		t.Errorf("expected no plugins registered, got %d", len(All()))
	}
}

/* ------------------------------------------------------------------------- */
/* HELPERS                                                                   */
/* ------------------------------------------------------------------------- */

func resetPluginSystem() {
	registry = nil
	factories = map[string]Factory{}
}
