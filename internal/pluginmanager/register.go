package pluginmanager

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/indaco/semver-cli/internal/config"
	"github.com/indaco/semver-cli/internal/plugins"
)

var RegisterLocalPluginFn = registerLocalPlugin

func registerLocalPlugin(localPath, configPath, pluginDirectory string) error {
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

	// 3. Determine pluginDirectory if not set
	if pluginDirectory == "" {
		pluginDirectory = "."
	}
	destPath := filepath.Join(pluginDirectory, ".semver-plugins", manifest.Name)

	// 4. Resolve and validate config path
	if configPath == "" {
		configPath = ".semver.yaml"
	}
	absConfigPath, _ := filepath.Abs(configPath)

	if _, err := os.Stat(absConfigPath); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, `
💡 To enable plugin support, create a .semver.yaml file in your project root. For example:

    echo "plugins: []" > .semver.yaml

Then run this command again.
`)
		return fmt.Errorf("config file not found at %s", absConfigPath)
	}

	// 5. Copy the plugin files to the destination directory
	if err := copyDirFn(localPath, destPath); err != nil {
		return fmt.Errorf("failed to copy plugin files: %w", err)
	}

	// 6. Update the config
	pluginCfg := config.PluginConfig{
		Name:    manifest.Name,
		Path:    destPath,
		Enabled: true,
	}

	// 7. Add the plugin to the config file
	if err := AddPluginToConfigFn(absConfigPath, pluginCfg); err != nil {
		return fmt.Errorf("failed to update config: %w", err)
	}

	// 8. Success message
	fmt.Printf("✅ Plugin %q registered successfully.\n", manifest.Name)
	return nil
}
