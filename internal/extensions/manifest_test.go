package extensions

import (
	"strings"
	"testing"
)

func TestExtensionManifest_Validate(t *testing.T) {
	base := ExtensionManifest{
		Name:        "commit-parser",
		Version:     "0.1.0",
		Description: "Parses conventional commits",
		Author:      "indaco",
		Repository:  "https://github.com/indaco/semver-commit-parser",
		Entry:       "github.com/indaco/semver-commit/parser",
	}

	tests := []struct {
		field    string
		modify   func(m *ExtensionManifest)
		expected string
	}{
		{"missing name", func(m *ExtensionManifest) { m.Name = "" }, "missing 'name'"},
		{"missing version", func(m *ExtensionManifest) { m.Version = "" }, "missing 'version'"},
		{"missing description", func(m *ExtensionManifest) { m.Description = "" }, "missing 'description'"},
		{"missing author", func(m *ExtensionManifest) { m.Author = "" }, "missing 'author'"},
		{"missing repository", func(m *ExtensionManifest) { m.Repository = "" }, "missing 'repository'"},
		{"missing entry", func(m *ExtensionManifest) { m.Entry = "" }, "missing 'entry'"},
	}

	for _, tt := range tests {
		t.Run(tt.field, func(t *testing.T) {
			m := base
			tt.modify(&m)

			err := m.ValidateManifest()
			if err == nil || !strings.Contains(err.Error(), tt.expected) {
				t.Errorf("expected error to contain %q, got %v", tt.expected, err)
			}
		})
	}

	t.Run("valid manifest", func(t *testing.T) {
		err := base.ValidateManifest()
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})
}
