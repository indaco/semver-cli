package changeloggenerator

import (
	"testing"

	"github.com/indaco/semver-cli/internal/config"
)

func TestRegisterChangelogGenerator(t *testing.T) {
	// Reset before test
	ResetChangelogGenerator()
	defer ResetChangelogGenerator()

	// Create a plugin
	cfg := DefaultConfig()
	plugin := NewChangelogGenerator(cfg)

	// Register
	registerChangelogGenerator(plugin)

	// Verify registration
	got := getChangelogGenerator()
	if got == nil {
		t.Fatal("expected registered generator, got nil")
	}
	if got.Name() != plugin.Name() {
		t.Errorf("Name() = %q, want %q", got.Name(), plugin.Name())
	}
}

func TestRegisterChangelogGenerator_Duplicate(t *testing.T) {
	// Reset before test
	ResetChangelogGenerator()
	defer ResetChangelogGenerator()

	// Create two plugins
	cfg := DefaultConfig()
	plugin1 := NewChangelogGenerator(cfg)
	plugin2 := NewChangelogGenerator(cfg)

	// Register first
	registerChangelogGenerator(plugin1)

	// Register second (should be ignored with warning)
	registerChangelogGenerator(plugin2)

	// Verify first is still registered
	got := getChangelogGenerator()
	if got == nil {
		t.Fatal("expected registered generator")
	}
	// Both have same name, but first should still be there
	if got != plugin1 {
		t.Error("expected first registered plugin to remain")
	}
}

func TestGetChangelogGenerator_None(t *testing.T) {
	// Reset before test
	ResetChangelogGenerator()
	defer ResetChangelogGenerator()

	got := getChangelogGenerator()
	if got != nil {
		t.Errorf("expected nil, got %v", got)
	}
}

func TestResetChangelogGenerator(t *testing.T) {
	// Reset before test
	ResetChangelogGenerator()

	// Register
	cfg := DefaultConfig()
	plugin := NewChangelogGenerator(cfg)
	registerChangelogGenerator(plugin)

	// Verify registered
	if getChangelogGenerator() == nil {
		t.Fatal("expected registered generator")
	}

	// Reset
	ResetChangelogGenerator()

	// Verify cleared
	if getChangelogGenerator() != nil {
		t.Error("expected nil after reset")
	}
}

func TestRegister(t *testing.T) {
	// Reset before test
	ResetChangelogGenerator()
	defer ResetChangelogGenerator()

	cfg := &config.ChangelogGeneratorConfig{
		Enabled: true,
		Mode:    "versioned",
	}

	Register(cfg)

	got := getChangelogGenerator()
	if got == nil {
		t.Fatal("expected registered generator")
	}
	if !got.IsEnabled() {
		t.Error("expected generator to be enabled")
	}
}

func TestRegister_NilConfig(t *testing.T) {
	// Reset before test
	ResetChangelogGenerator()
	defer ResetChangelogGenerator()

	// Should use default config when nil
	Register(nil)

	got := getChangelogGenerator()
	if got == nil {
		t.Fatal("expected registered generator even with nil config")
	}
}

func TestFunctionVariables(t *testing.T) {
	// Test that function variables are properly set
	if RegisterChangelogGeneratorFn == nil {
		t.Error("RegisterChangelogGeneratorFn should not be nil")
	}
	if GetChangelogGeneratorFn == nil {
		t.Error("GetChangelogGeneratorFn should not be nil")
	}
}
