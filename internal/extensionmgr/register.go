package extensionmgr

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/indaco/semver-cli/internal/config"
	"github.com/indaco/semver-cli/internal/extensions"
)

var (
	userHomeDirFn            = os.UserHomeDir
	RegisterLocalExtensionFn = registerLocalExtension
)

func registerLocalExtension(localPath, configPath, extensionDirectory string) error {
	// 1. Validate source path (ensure it's a directory)
	info, err := os.Stat(localPath)
	if err != nil {
		return fmt.Errorf("extension path error: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("extension path must be a directory")
	}

	// 2. Load and validate the extension manifest
	manifest, err := extensions.LoadExtensionManifestFn(localPath)
	if err != nil {
		return fmt.Errorf("failed to load extension manifest: %w", err)
	}

	// 3. Resolve base extension directory
	baseDir := extensionDirectory
	if baseDir == "" || baseDir == "." {
		homeDir, err := userHomeDirFn()
		if err != nil {
			return fmt.Errorf("failed to get user home directory: %w", err)
		}
		baseDir = filepath.Join(homeDir, ".semver-extensions")
	} else {
		baseDir = filepath.Join(baseDir, ".semver-extensions")
	}

	destPath := filepath.Join(baseDir, manifest.Name)

	// 4. Resolve and validate config path
	if configPath == "" {
		configPath = ".semver.yaml"
	}
	absConfigPath, _ := filepath.Abs(configPath)

	if _, err := os.Stat(absConfigPath); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, `
To enable extension support, create a .semver.yaml file in your project root. For example:

    echo "extensions: []" > .semver.yaml

Then run this command again.
`)
		return fmt.Errorf("config file not found at %s", absConfigPath)
	}

	// 5. Copy the extension files to the destination directory
	if err := copyDirFn(localPath, destPath); err != nil {
		return fmt.Errorf("failed to copy extension files: %w", err)
	}

	// 6. Update the config
	extensionCfg := config.ExtensionConfig{
		Name:    manifest.Name,
		Path:    destPath,
		Enabled: true,
	}

	// 7. Add the extension to the config file
	if err := AddExtensionToConfigFn(absConfigPath, extensionCfg); err != nil {
		return fmt.Errorf("failed to update config: %w", err)
	}

	// 8. Success message
	fmt.Printf("Extension %q registered successfully.\n", manifest.Name)
	return nil
}
