package pluginmanager

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/indaco/semver-cli/internal/config"
	"github.com/indaco/semver-cli/internal/plugins"
	"github.com/indaco/semver-cli/internal/testutils"
)

func TestRegisterLocalPlugin_Success(t *testing.T) {
	tmpDir := t.TempDir()
	pluginDir := filepath.Join(tmpDir, "myplugin")
	if err := os.Mkdir(pluginDir, 0755); err != nil {
		t.Fatal(err)
	}

	manifestContent := `
name: test-plugin
version: 1.0.0
description: A test plugin
author: John Doe
repository: https://github.com/test/plugin
entry: plugin.templ
`
	if err := os.WriteFile(filepath.Join(pluginDir, "plugin.yaml"), []byte(manifestContent), 0644); err != nil {
		t.Fatal(err)
	}

	cfgPath := filepath.Join(tmpDir, ".semver.yaml")
	if err := os.WriteFile(cfgPath, []byte("path: .version\n"), 0644); err != nil {
		t.Fatal(err)
	}

	// Override .semver-plugins dir for test
	originalCopyDir := CopyDirFn
	defer func() { CopyDirFn = originalCopyDir }()

	CopyDirFn = func(src, dst string) error {
		if !strings.Contains(src, "myplugin") || !strings.Contains(dst, "test-plugin") {
			t.Errorf("unexpected copy src=%q dst=%q", src, dst)
		}
		return nil
	}

	err := RegisterLocalPlugin(pluginDir, cfgPath, tmpDir)
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
}

func TestRegisterLocalPlugin_InvalidPath(t *testing.T) {
	tmpDir := os.TempDir()
	err := RegisterLocalPlugin("/nonexistent/path", ".semver.yaml", tmpDir)
	if err == nil || !strings.Contains(err.Error(), "plugin path error") {
		t.Errorf("expected plugin path error, got: %v", err)
	}
}

func TestRegisterLocalPlugin_NotDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	file := filepath.Join(tmpDir, "file.txt")
	_ = os.WriteFile(file, []byte("test"), 0644)

	err := RegisterLocalPlugin(file, ".semver.yaml", tmpDir)
	if err == nil || !strings.Contains(err.Error(), "must be a directory") {
		t.Errorf("expected directory error, got: %v", err)
	}
}

func TestRegisterLocalPlugin_InvalidManifest(t *testing.T) {
	tmpDir := t.TempDir()
	pluginDir := filepath.Join(tmpDir, "invalidplugin")
	_ = os.Mkdir(pluginDir, 0755)
	_ = os.WriteFile(filepath.Join(pluginDir, "plugin.yaml"), []byte("invalid: yaml:::"), 0644)

	err := RegisterLocalPlugin(pluginDir, ".semver.yaml", tmpDir)
	if err == nil || !strings.Contains(err.Error(), "failed to load plugin manifest") {
		t.Errorf("expected manifest load error, got: %v", err)
	}
}

func TestRegisterLocalPlugin_CopyDirFails(t *testing.T) {
	tmpDir := os.TempDir()
	// Setup mock plugin directory
	pluginDir := setupPluginDir(t, "mock-plugin", "1.0.0")

	// Create the config file
	configPath := testutils.WriteTempConfig(t, "plugins: []\n")

	// Temporarily override CopyDirFn to simulate failure
	originalCopyDirFn := CopyDirFn
	CopyDirFn = func(src, dst string) error {
		return fmt.Errorf("simulated copy failure")
	}
	defer func() {
		// Restore original CopyDir function
		CopyDirFn = originalCopyDirFn
	}()

	// Call RegisterLocalPlugin which should now fail due to the simulated copy error
	err := RegisterLocalPlugin(pluginDir, configPath, tmpDir)
	if err == nil {
		t.Fatal("expected error when copying, got nil")
	}

	if !strings.Contains(err.Error(), "simulated copy failure") {
		t.Fatalf("expected simulated copy error, got: %v", err)
	}
}

func TestRegisterLocalPlugin_DefaultConfigPath(t *testing.T) {
	content := "path: .version"
	tmpConfigPath := testutils.WriteTempConfig(t, content)
	tmpDir := filepath.Dir(tmpConfigPath)
	tmpPluginDir := setupPluginDir(t, "mock-plugin", "1.0.0")

	origDir, err := os.Getwd() // Get the original working directory to restore later
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}

	if err := os.Chdir(tmpDir); err != nil { // Change to the directory of the temporary config file
		t.Fatalf("failed to change directory to %s: %v", tmpDir, err)
	}
	t.Cleanup(func() { // Ensure we restore the original working directory after the test
		if err := os.Chdir(origDir); err != nil {
			t.Fatalf("failed to restore working directory: %v", err)
		}
	})

	// Register the plugin for the first time
	err = RegisterLocalPlugin(tmpPluginDir, tmpConfigPath, tmpDir)
	if err != nil {
		t.Fatalf("expected no error on first plugin registration, got: %v", err)
	}

	// Register the plugin again
	err = RegisterLocalPlugin(tmpPluginDir, tmpConfigPath, tmpDir)
	if err != nil {
		t.Fatalf("expected no error on second plugin registration, got: %v", err)
	}

	// Check if the .semver.yaml file exists before loading it
	if _, err := os.Stat(tmpConfigPath); os.IsNotExist(err) {
		t.Fatalf(".semver.yaml file does not exist at %s", tmpConfigPath)
	}

	// Ensure the config file has the plugin registered
	cfg, err := config.LoadConfigFn()
	if err != nil {
		t.Fatalf("expected no error loading config, got: %v", err)
	}

	// Guard check for nil config
	if cfg == nil {
		t.Fatal("config is nil after loading")
	}

	// Check that there's exactly one plugin
	if len(cfg.Plugins) != 1 {
		t.Fatalf("expected 1 plugin in config, got: %d", len(cfg.Plugins))
	}

	// Ensure that the default config path has been used if configPath is empty
	if cfg.Path != ".version" {
		t.Errorf("expected config path to be .version, got: %s", cfg.Path)
	}

	// Test that the default config path is used when no configPath is passed
	err = RegisterLocalPlugin(tmpPluginDir, "", tmpDir)
	if err != nil {
		t.Fatalf("expected no error on second plugin registration with empty configPath, got: %v", err)
	}

	// Verify the path has been set to the default value when configPath is empty
	cfg, err = config.LoadConfigFn()
	if err != nil {
		t.Fatalf("expected no error loading config, got: %v", err)
	}

	// Ensure the path is still the default
	if cfg.Path != ".version" {
		t.Errorf("expected config path to be .version, got: %s", cfg.Path)
	}
}

func TestRegisterLocalPlugin_DefaultConfigPathUsed_CurrentWorkingDir(t *testing.T) {
	content := "path: .version"
	tmpConfigPath := testutils.WriteTempConfig(t, content)
	tmpDir := filepath.Dir(tmpConfigPath)
	tmpPluginDir := setupPluginDir(t, "mock-plugin", "1.0.0")

	// Create .semver-plugins directory in the current working directory (not temp folder)
	semverPluginsDir := ".semver-plugins"
	if err := os.MkdirAll(semverPluginsDir, 0755); err != nil {
		t.Fatalf("failed to create .semver-plugins directory: %v", err)
	}

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}

	// Change to the directory of the temporary config file
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory to %s: %v", tmpDir, err)
	}
	t.Cleanup(func() {
		// Ensure we restore the original working directory after the test
		if err := os.Chdir(origDir); err != nil {
			t.Fatalf("failed to restore working directory: %v", err)
		}

		// Cleanup: remove the .semver-plugins directory in the current directory
		if err := os.RemoveAll(semverPluginsDir); err != nil {
			t.Fatalf("failed to remove .semver-plugins directory: %v", err)
		}
	})

	// Register the plugin with the current directory as the plugin path ("" or ".")
	err = RegisterLocalPlugin(tmpPluginDir, tmpConfigPath, "")
	if err != nil {
		t.Fatalf("expected no error on plugin registration, got: %v", err)
	}

	// Ensure the plugin was copied into the current .semver-plugins folder
	pluginPath := filepath.Join(semverPluginsDir, "mock-plugin")
	if _, err := os.Stat(pluginPath); os.IsNotExist(err) {
		t.Fatalf("plugin folder does not exist at %s", pluginPath)
	}

	// Ensure the config file has the plugin registered
	cfg, err := config.LoadConfigFn()
	if err != nil {
		t.Fatalf("expected no error loading config, got: %v", err)
	}

	if len(cfg.Plugins) != 1 {
		t.Fatalf("expected 1 plugin in config, got: %d", len(cfg.Plugins))
	}
}

func TestRegisterLocalPlugin_DefaultConfigPathUsed_OtherDir(t *testing.T) {
	content := "path: .version"
	tmpConfigPath := testutils.WriteTempConfig(t, content)
	tmpDir := filepath.Dir(tmpConfigPath)
	tmpPluginDir := setupPluginDir(t, "mock-plugin", "1.0.0")

	// Set up a temporary directory for the plugin
	tmpPluginFolder := t.TempDir()

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}

	// Change to the directory of the temporary config file
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory to %s: %v", tmpDir, err)
	}
	t.Cleanup(func() {
		// Ensure we restore the original working directory after the test
		if err := os.Chdir(origDir); err != nil {
			t.Fatalf("failed to restore working directory: %v", err)
		}
	})

	// Register the plugin with the temporary plugin folder
	err = RegisterLocalPlugin(tmpPluginDir, tmpConfigPath, tmpPluginFolder)
	if err != nil {
		t.Fatalf("expected no error on plugin registration, got: %v", err)
	}

	// Ensure the plugin was copied into the temporary plugin folder
	pluginPath := filepath.Join(tmpPluginFolder, ".semver-plugins", "mock-plugin")
	if _, err := os.Stat(pluginPath); os.IsNotExist(err) {
		t.Fatalf("plugin folder does not exist at %s", pluginPath)
	}

	// Ensure the config file has the plugin registered
	cfg, err := config.LoadConfigFn()
	if err != nil {
		t.Fatalf("expected no error loading config, got: %v", err)
	}

	if len(cfg.Plugins) != 1 {
		t.Fatalf("expected 1 plugin in config, got: %d", len(cfg.Plugins))
	}
}

func TestRegisterLocalPlugin_ValidConfigPath(t *testing.T) {
	// Set up temporary directories
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".semver.yaml")

	// Create a mock config file at the given path
	if err := os.WriteFile(configPath, []byte("path: .version"), 0644); err != nil {
		t.Fatalf("failed to create mock .semver.yaml: %v", err)
	}

	// Create a mock plugin directory
	pluginDir := t.TempDir()
	pluginPath := filepath.Join(pluginDir, "plugin.yaml")

	// Create a mock plugin manifest file
	pluginManifest := []byte(`
name: mock-plugin
version: "1.0.0"
description: Mock plugin
author: Test Author
repository: https://github.com/test/repo
entry: mock-entry
`)

	if err := os.WriteFile(pluginPath, pluginManifest, 0644); err != nil {
		t.Fatalf("failed to create mock plugin.yaml: %v", err)
	}

	// Call the RegisterLocalPlugin function
	err := RegisterLocalPlugin(pluginDir, configPath, tmpDir)
	if err != nil {
		t.Fatalf("expected no error during plugin registration, got: %v", err)
	}

	// Check that the plugin was successfully registered (the test should pass if no error is returned)
}

func TestRegisterLocalPlugin_InvalidConfigPath(t *testing.T) {
	// Set up temporary directories
	tmpDir := t.TempDir()

	// Use a non-existent config path for testing
	nonExistentConfigPath := filepath.Join(tmpDir, "nonexistent-config.yaml")

	// Create a mock plugin directory
	pluginDir := t.TempDir()
	pluginPath := filepath.Join(pluginDir, "plugin.yaml")

	// Create a mock plugin manifest file
	pluginManifest := []byte(`
name: mock-plugin
version: "1.0.0"
description: Mock plugin
author: Test Author
repository: https://github.com/test/repo
entry: mock-entry
`)

	if err := os.WriteFile(pluginPath, pluginManifest, 0644); err != nil {
		t.Fatalf("failed to create mock plugin.yaml: %v", err)
	}

	// Call the RegisterLocalPlugin function with an invalid config path
	err := RegisterLocalPlugin(pluginDir, nonExistentConfigPath, tmpDir)
	if err == nil {
		t.Fatal("expected error due to non-existent config file, got nil")
	}

	// Check that the error message contains "config file not found"
	expectedErr := fmt.Sprintf("config file not found at %s", nonExistentConfigPath)
	if !strings.Contains(err.Error(), expectedErr) {
		t.Errorf("expected error to contain %q, got: %v", expectedErr, err)
	}
}

func TestRegisterLocalPlugin_InvalidConfigPathResolution(t *testing.T) {
	// Create a temporary plugin directory
	tmpPluginDir := setupPluginDir(t, "mock-plugin", "1.0.0")

	// Simulate an invalid config path
	invalidConfigPath := "/invalid/path/to/.semver.yaml"

	// Try registering the plugin with the invalid config path
	err := RegisterLocalPlugin(tmpPluginDir, invalidConfigPath, os.TempDir())
	if err == nil {
		t.Fatal("expected error due to invalid config path resolution, got nil")
	}

	// Check if the error message is about the config file not being found
	expectedErr := "config file not found"
	if !strings.Contains(err.Error(), expectedErr) {
		t.Errorf("expected error message to contain %q, got: %v", expectedErr, err)
	}
}

func TestRegisterLocalPlugin_AddPluginToConfigError(t *testing.T) {
	// Set up the initial config
	tmpConfigPath := testutils.WriteTempConfig(t, `path: .version`)
	tmpPluginDir := setupPluginDir(t, "mock-plugin", "1.0.0")

	// Simulate the error returned by AddPluginToConfig
	localPath := t.TempDir()
	cfgPath := tmpConfigPath // Path to the config file

	// Mock the LoadPluginManifest function to return a mock manifest
	mockManifest := &plugins.PluginManifest{
		Name:        "mock-plugin",
		Version:     "1.0.0",
		Description: "Mock Plugin",
		Author:      "Test Author",
		Repository:  "https://github.com/test/repo",
		Entry:       "mock-entry",
	}

	// Mock LoadPluginManifest to return the mock manifest
	plugins.LoadPluginManifestFn = func(path string) (*plugins.PluginManifest, error) {
		return mockManifest, nil
	}

	// Simulate AddPluginToConfig error by overriding the function
	originalAddPluginToConfig := AddPluginToConfigFn
	defer func() {
		AddPluginToConfigFn = originalAddPluginToConfig // Restore original after test
	}()
	AddPluginToConfigFn = func(path string, plugin config.PluginConfig) error {
		return fmt.Errorf("failed to update config: some error")
	}

	// Attempt to register the plugin
	err := RegisterLocalPlugin(localPath, cfgPath, tmpPluginDir)

	// Check that we get the expected error
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "failed to update config") {
		t.Errorf("unexpected error: %v", err)
	}
}

func setupPluginDir(t *testing.T, name, version string) string {
	t.Helper()

	dir := t.TempDir()
	manifestContent := fmt.Sprintf(`name: %s
version: %s
description: test plugin
author: test
repository: https://example.com/%s.git
entry: plugin.templ
`, name, version, name)

	if err := os.WriteFile(filepath.Join(dir, "plugin.yaml"), []byte(manifestContent), 0644); err != nil {
		t.Fatalf("failed to write plugin.yaml: %v", err)
	}

	return dir
}
