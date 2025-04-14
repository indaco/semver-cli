package pluginmanager

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/indaco/semver-cli/internal/config"
	"github.com/indaco/semver-cli/internal/plugins"
)

var RegisterLocalPluginFn = RegisterLocalPlugin

func RegisterLocalPlugin(localPath, configPath, pluginDirectory string) error {
	// 1. Validate source path (ensure it's a directory)
	info, err := os.Stat(localPath)
	if err != nil {
		return fmt.Errorf("plugin path error: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("plugin path must be a directory")
	}

	// 2. Load and validate the plugin manifest
	manifest, err := plugins.LoadPluginManifestFn(localPath)
	if err != nil {
		return fmt.Errorf("failed to load plugin manifest: %w", err)
	}

	// 3. Determine the destination directory for plugins
	// If pluginDirectory is provided, use it. Otherwise, default to current directory.
	if pluginDirectory == "" {
		pluginDirectory = "."
	}

	destPath := filepath.Join(pluginDirectory, ".semver-plugins", manifest.Name)

	// 4. Copy the plugin files to the destination directory
	if err := CopyDirFn(localPath, destPath); err != nil {
		return fmt.Errorf("failed to copy plugin files: %w", err)
	}

	// 5. Ensure configPath is set, default to ".semver.yaml" if empty
	if configPath == "" {
		configPath = ".semver.yaml"
	}

	// 6. Resolve the config path to an absolute path
	absConfigPath, _ := filepath.Abs(configPath)

	// 7. Ensure the config file exists before trying to update it
	if _, err := os.Stat(absConfigPath); os.IsNotExist(err) {
		return fmt.Errorf("config file not found at %s", absConfigPath)
	}

	// 8. Prepare the plugin config entry
	pluginCfg := config.PluginConfig{
		Name:    manifest.Name,
		Path:    destPath,
		Enabled: true,
	}

	// 9. Add the plugin to the config file
	if err := AddPluginToConfigFn(absConfigPath, pluginCfg); err != nil {
		return fmt.Errorf("failed to update config: %w", err)
	}

	// 10. Success message
	fmt.Printf("✅ Plugin %q registered successfully.\n", manifest.Name)
	return nil
}
