package extensions

import (
	"testing"

	"github.com/indaco/semver-cli/internal/testutils"
)

func TestRegisterExtensionMeta(t *testing.T) {
	// Reset before and after
	ResetExtension()
	defer ResetExtension()

	p := testutils.MockExtension{
		NameValue:        "mock",
		VersionValue:     "v1.0.0",
		DescriptionValue: "A mock plugin for testing",
	}
	RegisterExtension(p)

	all := AllExtensions()
	if len(all) != 1 {
		t.Fatalf("expected 1 plugin, got %d", len(all))
	}

	if all[0].Name() != "mock" {
		t.Errorf("expected plugin name 'mock', got %q", all[0].Name())
	}
	if all[0].Description() != "A mock plugin for testing" {
		t.Errorf("unexpected description: %q", all[0].Description())
	}
	if all[0].Version() != "v1.0.0" {
		t.Errorf("unexpected version: %q", all[0].Version())
	}
}
