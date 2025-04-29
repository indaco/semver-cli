package plugins

import (
	"errors"
)

// PluginManifest defines the metadata and entry point for a semver plugin.
// This structure is expected to be defined in a plugin's `plugin.yaml` file.
//
// All fields are required:
// - Name: A unique plugin identifier (e.g. "commit-parser")
// - Version: The plugin's version (e.g. "0.1.0")
// - Description: A brief explanation of what the plugin does
// - Author: Name or handle of the plugin author
// - Repository: URL of the plugin's source repository
// - Entry: Go package path of the entry point (e.g. "github.com/user/semver-plugin-commit/parser")
type PluginManifest struct {
	Name        string `yaml:"name"`
	Version     string `yaml:"version"`
	Description string `yaml:"description"`
	Author      string `yaml:"author"`
	Repository  string `yaml:"repository"`
	Entry       string `yaml:"entry"`
}

// ValidateManifest ensures all required fields are present
func (m *PluginManifest) ValidateManifest() error {
	if m.Name == "" {
		return errors.New("plugin manifest: missing 'name'")
	}
	if m.Version == "" {
		return errors.New("plugin manifest: missing 'version'")
	}
	if m.Description == "" {
		return errors.New("plugin manifest: missing 'description'")
	}
	if m.Author == "" {
		return errors.New("plugin manifest: missing 'author'")
	}
	if m.Repository == "" {
		return errors.New("plugin manifest: missing 'repository'")
	}
	if m.Entry == "" {
		return errors.New("plugin manifest: missing 'entry'")
	}
	return nil
}
