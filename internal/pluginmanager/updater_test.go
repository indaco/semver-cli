package pluginmanager

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/goccy/go-yaml"
	"github.com/indaco/semver-cli/internal/config"
)

func TestAddPluginToConfig_Success(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".semver.yaml")

	initial := []byte("path: .version\nplugins: []\n")
	if err := os.WriteFile(configPath, initial, 0644); err != nil {
		t.Fatal(err)
	}

	plugin := config.PluginConfig{
		Name:    "commit-parser",
		Path:    ".semver-plugins/commit-parser",
		Enabled: true,
	}

	if err := AddPluginToConfig(configPath, plugin); err != nil {
		t.Fatalf("expected success, got: %v", err)
	}

	// Re-read and verify
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}

	var parsed config.Config
	if err := yaml.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to unmarshal updated config: %v", err)
	}

	if len(parsed.Plugins) != 1 {
		t.Fatalf("expected 1 plugin, got %d", len(parsed.Plugins))
	}

	got := parsed.Plugins[0]
	if got.Name != plugin.Name || got.Path != plugin.Path || !got.Enabled {
		t.Errorf("unexpected plugin entry: %+v", got)
	}
}

func TestAddPluginToConfig_Duplicate(t *testing.T) {
	tmpDir := t.TempDir()

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory to %s: %v", tmpDir, err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(origDir); err != nil {
			t.Fatalf("failed to restore working directory: %v", err)
		}
	})

	configPath := filepath.Join(tmpDir, ".semver.yaml")
	// Initial config with one plugin
	initial := []byte(`
path: .version
plugins:
  - name: commit-parser
    path: .semver-plugins/commit-parser
    enabled: true
`)
	if err := os.WriteFile(configPath, initial, 0644); err != nil {
		t.Fatal(err)
	}

	plugin := config.PluginConfig{
		Name:    "commit-parser",
		Path:    ".semver-plugins/commit-parser",
		Enabled: true,
	}

	// First registration (no error expected)
	err = AddPluginToConfig(configPath, plugin)
	if err != nil {
		t.Fatalf("unexpected error during first registration: %v", err)
	}

	// Second registration (no error expected, duplicates are silently skipped)
	err = AddPluginToConfig(configPath, plugin)
	if err != nil {
		t.Fatalf("unexpected error during second registration: %v", err)
	}

	// Ensure the config file has only one plugin
	cfg, err := config.LoadConfig()
	if err != nil {
		t.Fatalf("expected no error loading config, got: %v", err)
	}
	if len(cfg.Plugins) != 1 {
		t.Fatalf("expected 1 plugin in config, got: %d", len(cfg.Plugins))
	}
}

func TestAddPluginToConfig_ReadFileError(t *testing.T) {
	invalidPath := filepath.Join(t.TempDir(), "nonexistent.yaml")

	plugin := config.PluginConfig{
		Name:    "test",
		Path:    "some/path",
		Enabled: true,
	}

	err := AddPluginToConfig(invalidPath, plugin)
	if err == nil || !os.IsNotExist(err) {
		t.Fatalf("expected file not found error, got: %v", err)
	}
}

func TestAddPluginToConfig_UnmarshalError(t *testing.T) {
	tmp := t.TempDir()
	configPath := filepath.Join(tmp, ".semver.yaml")

	badYAML := []byte(": invalid yaml")
	if err := os.WriteFile(configPath, badYAML, 0644); err != nil {
		t.Fatal(err)
	}

	err := AddPluginToConfig(configPath, config.PluginConfig{
		Name:    "test",
		Path:    "some/path",
		Enabled: true,
	})

	if err == nil || !strings.Contains(err.Error(), "unexpected key name") {
		t.Fatalf("expected YAML unmarshal error, got: %v", err)
	}
}

func TestAddPluginToConfig_MarshalError(t *testing.T) {
	// Create a temporary file with a valid config
	tmp := t.TempDir()
	configPath := filepath.Join(tmp, ".semver.yaml")
	initial := []byte(`path: .version`)
	if err := os.WriteFile(configPath, initial, 0644); err != nil {
		t.Fatal(err)
	}

	// Backup the original yaml.Marshal
	originalMarshal := marshalFunc
	defer func() { marshalFunc = originalMarshal }()

	// Force yaml.Marshal to fail
	marshalFunc = func(v any) ([]byte, error) {
		return nil, errors.New("forced marshal failure")
	}

	err := AddPluginToConfig(configPath, config.PluginConfig{
		Name:    "fail-marshaling",
		Path:    ".semver-plugins/fail",
		Enabled: true,
	})

	if err == nil || !strings.Contains(err.Error(), "forced marshal failure") {
		t.Fatalf("expected marshal error, got: %v", err)
	}
}

func TestAddPluginToConfig_WriteFileError(t *testing.T) {
	tmp := t.TempDir()
	configPath := filepath.Join(tmp, ".semver.yaml")

	initial := []byte("path: .version\nplugins: []\n")
	if err := os.WriteFile(configPath, initial, 0444); err != nil {
		t.Fatal(err)
	}
	// Ensure cleanup restores perms so t.TempDir can delete
	t.Cleanup(func() {
		_ = os.Chmod(configPath, 0644)
	})

	err := AddPluginToConfig(configPath, config.PluginConfig{
		Name:    "test",
		Path:    "some/path",
		Enabled: true,
	})
	if err == nil || !strings.Contains(err.Error(), "permission denied") {
		t.Fatalf("expected write error, got: %v", err)
	}
}
