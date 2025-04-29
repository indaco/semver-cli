package extensions

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"github.com/goccy/go-yaml"
)

var LoadExtensionManifestFn = loadExtensionManifest

// loadExtensionManifest loads and validates a extension.yaml file from the given directory.
func loadExtensionManifest(dir string) (*ExtensionManifest, error) {
	manifestPath := filepath.Join(dir, "extension.yaml")

	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read manifest: %w", err)
	}

	var manifest ExtensionManifest
	decoder := yaml.NewDecoder(bytes.NewReader(data), yaml.Strict())

	if err := decoder.Decode(&manifest); err != nil {
		return nil, fmt.Errorf("failed to parse manifest: %w", err)
	}

	if err := manifest.ValidateManifest(); err != nil {
		return nil, fmt.Errorf("invalid plugin manifest: %w", err)
	}

	return &manifest, nil
}
