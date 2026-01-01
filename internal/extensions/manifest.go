package extensions

import (
	"errors"
)

// ExtensionManifest defines the metadata and entry point for a semver extension.
// This structure is expected to be defined in a extension's `extension.yaml` file.
//
// All fields except Hooks are required:
// - Name: A unique extension identifier (e.g. "changelog-generator")
// - Version: The extension's version (e.g. "0.1.0")
// - Description: A brief explanation of what the extension does
// - Author: Name or handle of the extension author
// - Repository: URL of the extension's source repository
// - Entry: Path to the executable script or binary (relative to extension directory)
// - Hooks: List of hook points this extension supports (optional)
type ExtensionManifest struct {
	Name        string   `yaml:"name"`
	Version     string   `yaml:"version"`
	Description string   `yaml:"description"`
	Author      string   `yaml:"author"`
	Repository  string   `yaml:"repository"`
	Entry       string   `yaml:"entry"`
	Hooks       []string `yaml:"hooks,omitempty"`
}

// ValidateManifest ensures all required fields are present
func (m *ExtensionManifest) ValidateManifest() error {
	if m.Name == "" {
		return errors.New("extension manifest: missing 'name'")
	}
	if m.Version == "" {
		return errors.New("extension manifest: missing 'version'")
	}
	if m.Description == "" {
		return errors.New("extension manifest: missing 'description'")
	}
	if m.Author == "" {
		return errors.New("extension manifest: missing 'author'")
	}
	if m.Repository == "" {
		return errors.New("extension manifest: missing 'repository'")
	}
	if m.Entry == "" {
		return errors.New("extension manifest: missing 'entry'")
	}
	return nil
}
