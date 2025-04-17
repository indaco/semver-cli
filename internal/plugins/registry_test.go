package plugins

import (
	"testing"

	"github.com/urfave/cli/v3"
)

type mockPlugin struct {
	registered bool
}

func (m *mockPlugin) Name() string {
	return "mock"
}

func (m *mockPlugin) Register(cmd *cli.Command) {
	m.registered = true
	cmd.Description = "plugin was here"
}

func TestRegisterAndAll(t *testing.T) {
	// Reset registry
	resetPlugins()
	defer resetPlugins()

	p := &mockPlugin{}
	Register(p)

	all := All()
	if len(all) != 1 {
		t.Fatalf("expected 1 plugin, got %d", len(all))
	}

	if all[0].Name() != "mock" {
		t.Errorf("expected plugin name 'mock', got %q", all[0].Name())
	}
}

func TestPluginRegisterModifiesCommand(t *testing.T) {
	// Reset registry
	resetPlugins()
	defer resetPlugins()

	p := &mockPlugin{}
	Register(p)

	root := &cli.Command{}
	p.Register(root)

	if root.Description != "plugin was here" {
		t.Errorf("expected root.Description to be set, got %q", root.Description)
	}
}

/* ------------------------------------------------------------------------- */
/* HELPERS                                                                   */
/* ------------------------------------------------------------------------- */

// resetPlugins empties the internal registry between tests
func resetPlugins() {
	registry = nil
}
