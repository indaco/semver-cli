package extensions

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeExtensionYAML(t *testing.T, dir, content string) string {
	t.Helper()
	path := filepath.Join(dir, "extension.yaml")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write extension.yaml: %v", err)
	}
	return path
}

func TestLoadExtensionManifest_Valid(t *testing.T) {
	dir := t.TempDir()
	content := `
name: test
version: 0.1.0
description: test plugin
author: me
repository: https://example.com/repo
entry: actions.json
`
	writeExtensionYAML(t, dir, content)

	m, err := LoadExtensionManifestFn(dir)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if m.Name != "test" {
		t.Errorf("expected name 'test', got %q", m.Name)
	}
}

func TestLoadExtensionManifest_MissingFile(t *testing.T) {
	dir := t.TempDir()
	_, err := LoadExtensionManifestFn(dir)
	if err == nil || !strings.Contains(err.Error(), "no such file") {
		t.Fatalf("expected file not found error, got %v", err)
	}
}

func TestLoadExtensionManifest_InvalidYAML(t *testing.T) {
	dir := t.TempDir()
	content := ": this is not valid yaml"
	writeExtensionYAML(t, dir, content)

	_, err := LoadExtensionManifestFn(dir)
	if err == nil || !strings.Contains(err.Error(), "failed to parse manifest:") {
		t.Fatalf("expected YAML parse error, got %v", err)
	}
}

func TestLoadExtensionManifest_InvalidManifest(t *testing.T) {
	dir := t.TempDir()
	content := `
name: ""
version: ""
description: ""
author: ""
repository: ""
entry: ""
`
	writeExtensionYAML(t, dir, content)

	_, err := LoadExtensionManifestFn(dir)
	if err == nil || !strings.Contains(err.Error(), "plugin manifest: missing") {
		t.Fatalf("expected validation error, got %v", err)
	}
}
