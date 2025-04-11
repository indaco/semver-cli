package plugins

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"github.com/goccy/go-yaml"
)

// LoadPluginManifest loads and validates a plugin.yaml file from the given directory.
func LoadPluginManifest(dir string) (*PluginManifest, error) {
	manifestPath := filepath.Join(dir, "plugin.yaml")

	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read manifest: %w", err)
	}

	var manifest PluginManifest
	decoder := yaml.NewDecoder(bytes.NewReader(data), yaml.Strict())

	if err := decoder.Decode(&manifest); err != nil {
		return nil, fmt.Errorf("failed to parse manifest: %w", err)
	}

	if err := manifest.ValidateManifest(); err != nil {
		return nil, fmt.Errorf("invalid plugin manifest: %w", err)
	}

	return &manifest, nil
}
