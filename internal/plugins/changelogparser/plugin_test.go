package changelogparser

import (
	"errors"
	"os"
	"strings"
	"testing"
)

func TestNewChangelogParser_Plugin(t *testing.T) {
	t.Run("with nil config", func(t *testing.T) {
		plugin := NewChangelogParser(nil)
		if plugin == nil {
			t.Fatal("expected non-nil plugin")
		}
		if plugin.config.Path != "CHANGELOG.md" {
			t.Errorf("expected default path 'CHANGELOG.md', got %s", plugin.config.Path)
		}
		if plugin.config.Priority != "changelog" {
			t.Errorf("expected default priority 'changelog', got %s", plugin.config.Priority)
		}
	})

	t.Run("with custom config", func(t *testing.T) {
		cfg := &Config{
			Enabled:                  true,
			Path:                     "docs/CHANGES.md",
			RequireUnreleasedSection: false,
			InferBumpType:            true,
			Priority:                 "commits",
		}
		plugin := NewChangelogParser(cfg)
		if plugin.config.Path != "docs/CHANGES.md" {
			t.Errorf("expected path 'docs/CHANGES.md', got %s", plugin.config.Path)
		}
		if plugin.config.Priority != "commits" {
			t.Errorf("expected priority 'commits', got %s", plugin.config.Priority)
		}
	})

	t.Run("with empty path applies default", func(t *testing.T) {
		cfg := &Config{
			Enabled: true,
			Path:    "",
		}
		plugin := NewChangelogParser(cfg)
		if plugin.config.Path != "CHANGELOG.md" {
			t.Errorf("expected default path 'CHANGELOG.md', got %s", plugin.config.Path)
		}
	})
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.Enabled {
		t.Error("expected Enabled to be false by default")
	}
	if cfg.Path != "CHANGELOG.md" {
		t.Errorf("expected default path 'CHANGELOG.md', got %s", cfg.Path)
	}
	if !cfg.RequireUnreleasedSection {
		t.Error("expected RequireUnreleasedSection to be true by default")
	}
	if !cfg.InferBumpType {
		t.Error("expected InferBumpType to be true by default")
	}
	if cfg.Priority != "changelog" {
		t.Errorf("expected default priority 'changelog', got %s", cfg.Priority)
	}
}

func TestPluginMetadata(t *testing.T) {
	plugin := NewChangelogParser(nil)

	if plugin.Name() != "changelog-parser" {
		t.Errorf("expected name 'changelog-parser', got %s", plugin.Name())
	}

	if plugin.Description() == "" {
		t.Error("expected non-empty description")
	}

	if plugin.Version() != "v0.1.0" {
		t.Errorf("expected version 'v0.1.0', got %s", plugin.Version())
	}
}

func TestIsEnabled(t *testing.T) {
	tests := []struct {
		name    string
		enabled bool
	}{
		{"enabled", true},
		{"disabled", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{Enabled: tt.enabled}
			plugin := NewChangelogParser(cfg)

			if plugin.IsEnabled() != tt.enabled {
				t.Errorf("expected IsEnabled() = %v, got %v", tt.enabled, plugin.IsEnabled())
			}
		})
	}
}

func TestGetConfig(t *testing.T) {
	cfg := &Config{
		Enabled:  true,
		Path:     "custom/CHANGELOG.md",
		Priority: "commits",
	}
	plugin := NewChangelogParser(cfg)

	gotCfg := plugin.GetConfig()
	if gotCfg.Path != "custom/CHANGELOG.md" {
		t.Errorf("expected path 'custom/CHANGELOG.md', got %s", gotCfg.Path)
	}
}

func TestInferBumpType_Plugin(t *testing.T) {
	// Save original and restore after test
	origOpenFile := openFileFn
	defer func() { openFileFn = origOpenFile }()

	t.Run("disabled plugin", func(t *testing.T) {
		cfg := &Config{Enabled: false, InferBumpType: true}
		plugin := NewChangelogParser(cfg)

		_, err := plugin.InferBumpType()
		if err == nil {
			t.Error("expected error when plugin disabled, got nil")
		}
		if !strings.Contains(err.Error(), "not enabled") {
			t.Errorf("expected 'not enabled' error, got: %v", err)
		}
	})

	t.Run("inference disabled", func(t *testing.T) {
		cfg := &Config{Enabled: true, InferBumpType: false}
		plugin := NewChangelogParser(cfg)

		_, err := plugin.InferBumpType()
		if err == nil {
			t.Error("expected error when inference disabled, got nil")
		}
		if !strings.Contains(err.Error(), "inference disabled") {
			t.Errorf("expected 'inference disabled' error, got: %v", err)
		}
	})

	t.Run("successful inference - major", func(t *testing.T) {
		changelog := `# Changelog

## [Unreleased]

### Removed
- Old API
`
		openFileFn = mockOpenFile(changelog)

		cfg := &Config{
			Enabled:       true,
			Path:          "CHANGELOG.md",
			InferBumpType: true,
		}
		plugin := NewChangelogParser(cfg)

		bumpType, err := plugin.InferBumpType()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if bumpType != "major" {
			t.Errorf("expected 'major', got %s", bumpType)
		}
	})

	t.Run("successful inference - minor", func(t *testing.T) {
		changelog := `# Changelog

## [Unreleased]

### Added
- New feature
`
		openFileFn = mockOpenFile(changelog)

		cfg := &Config{
			Enabled:       true,
			Path:          "CHANGELOG.md",
			InferBumpType: true,
		}
		plugin := NewChangelogParser(cfg)

		bumpType, err := plugin.InferBumpType()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if bumpType != "minor" {
			t.Errorf("expected 'minor', got %s", bumpType)
		}
	})

	t.Run("successful inference - patch", func(t *testing.T) {
		changelog := `# Changelog

## [Unreleased]

### Fixed
- Bug fix
`
		openFileFn = mockOpenFile(changelog)

		cfg := &Config{
			Enabled:       true,
			Path:          "CHANGELOG.md",
			InferBumpType: true,
		}
		plugin := NewChangelogParser(cfg)

		bumpType, err := plugin.InferBumpType()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if bumpType != "patch" {
			t.Errorf("expected 'patch', got %s", bumpType)
		}
	})

	t.Run("parse error", func(t *testing.T) {
		openFileFn = func(name string) (*os.File, error) {
			return nil, os.ErrNotExist
		}

		cfg := &Config{
			Enabled:       true,
			Path:          "CHANGELOG.md",
			InferBumpType: true,
		}
		plugin := NewChangelogParser(cfg)

		_, err := plugin.InferBumpType()
		if err == nil {
			t.Error("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "failed to parse") {
			t.Errorf("expected 'failed to parse' error, got: %v", err)
		}
	})

	t.Run("no entries in unreleased section", func(t *testing.T) {
		changelog := `# Changelog

## [Unreleased]

## [1.0.0] - 2024-01-01
`
		openFileFn = mockOpenFile(changelog)

		cfg := &Config{
			Enabled:       true,
			Path:          "CHANGELOG.md",
			InferBumpType: true,
		}
		plugin := NewChangelogParser(cfg)

		_, err := plugin.InferBumpType()
		if err == nil {
			t.Error("expected error for empty unreleased section, got nil")
		}
		if !strings.Contains(err.Error(), "failed to infer bump type") {
			t.Errorf("expected 'failed to infer bump type' error, got: %v", err)
		}
	})
}

func TestValidateHasEntries(t *testing.T) {
	// Save original and restore after test
	origOpenFile := openFileFn
	defer func() { openFileFn = origOpenFile }()

	t.Run("disabled plugin", func(t *testing.T) {
		cfg := &Config{Enabled: false, RequireUnreleasedSection: true}
		plugin := NewChangelogParser(cfg)

		err := plugin.ValidateHasEntries()
		if err != nil {
			t.Errorf("expected nil error when plugin disabled, got: %v", err)
		}
	})

	t.Run("validation disabled", func(t *testing.T) {
		cfg := &Config{Enabled: true, RequireUnreleasedSection: false}
		plugin := NewChangelogParser(cfg)

		err := plugin.ValidateHasEntries()
		if err != nil {
			t.Errorf("expected nil error when validation disabled, got: %v", err)
		}
	})

	t.Run("valid changelog with entries", func(t *testing.T) {
		changelog := `# Changelog

## [Unreleased]

### Added
- New feature
`
		openFileFn = mockOpenFile(changelog)

		cfg := &Config{
			Enabled:                  true,
			Path:                     "CHANGELOG.md",
			RequireUnreleasedSection: true,
		}
		plugin := NewChangelogParser(cfg)

		err := plugin.ValidateHasEntries()
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("empty unreleased section", func(t *testing.T) {
		changelog := `# Changelog

## [Unreleased]

## [1.0.0] - 2024-01-01
`
		openFileFn = mockOpenFile(changelog)

		cfg := &Config{
			Enabled:                  true,
			Path:                     "CHANGELOG.md",
			RequireUnreleasedSection: true,
		}
		plugin := NewChangelogParser(cfg)

		err := plugin.ValidateHasEntries()
		if err == nil {
			t.Error("expected error for empty unreleased section, got nil")
		}
		if !strings.Contains(err.Error(), "no entries") {
			t.Errorf("expected 'no entries' error, got: %v", err)
		}
	})

	t.Run("missing unreleased section", func(t *testing.T) {
		changelog := `# Changelog

## [1.0.0] - 2024-01-01

### Added
- Old feature
`
		openFileFn = mockOpenFile(changelog)

		cfg := &Config{
			Enabled:                  true,
			Path:                     "CHANGELOG.md",
			RequireUnreleasedSection: true,
		}
		plugin := NewChangelogParser(cfg)

		err := plugin.ValidateHasEntries()
		if err == nil {
			t.Error("expected error for missing unreleased section, got nil")
		}
		if !strings.Contains(err.Error(), "changelog validation failed") {
			t.Errorf("expected 'changelog validation failed' error, got: %v", err)
		}
	})

	t.Run("file not found", func(t *testing.T) {
		openFileFn = func(name string) (*os.File, error) {
			return nil, os.ErrNotExist
		}

		cfg := &Config{
			Enabled:                  true,
			Path:                     "CHANGELOG.md",
			RequireUnreleasedSection: true,
		}
		plugin := NewChangelogParser(cfg)

		err := plugin.ValidateHasEntries()
		if err == nil {
			t.Error("expected error for missing file, got nil")
		}
	})
}

func TestShouldTakePrecedence(t *testing.T) {
	tests := []struct {
		name     string
		enabled  bool
		priority string
		expected bool
	}{
		{
			name:     "enabled with changelog priority",
			enabled:  true,
			priority: "changelog",
			expected: true,
		},
		{
			name:     "enabled with commits priority",
			enabled:  true,
			priority: "commits",
			expected: false,
		},
		{
			name:     "disabled with changelog priority",
			enabled:  false,
			priority: "changelog",
			expected: false,
		},
		{
			name:     "disabled with commits priority",
			enabled:  false,
			priority: "commits",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				Enabled:  tt.enabled,
				Priority: tt.priority,
			}
			plugin := NewChangelogParser(cfg)

			result := plugin.ShouldTakePrecedence()
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// mockOpenFile returns a function that creates a mock file from a string
func mockOpenFile(content string) func(string) (*os.File, error) {
	return func(name string) (*os.File, error) {
		// Create a temporary file with the content
		tmpFile, err := os.CreateTemp("", "changelog-test-*.md")
		if err != nil {
			return nil, err
		}

		if _, err := tmpFile.WriteString(content); err != nil {
			tmpFile.Close()
			os.Remove(tmpFile.Name())
			return nil, err
		}

		// Seek back to start
		if _, err := tmpFile.Seek(0, 0); err != nil {
			tmpFile.Close()
			os.Remove(tmpFile.Name())
			return nil, err
		}

		// The file will be cleaned up by the test cleanup
		return tmpFile, nil
	}
}

func TestPluginInterface(t *testing.T) {
	var _ ChangelogInferrer = (*ChangelogParserPlugin)(nil)
}

func TestRegistry(t *testing.T) {
	// Save original state
	origParser := defaultChangelogParser
	defer func() { defaultChangelogParser = origParser }()

	t.Run("register and get", func(t *testing.T) {
		ResetChangelogParser()

		cfg := &Config{Enabled: true}
		Register(cfg)

		parser := GetChangelogParserFn()
		if parser == nil {
			t.Fatal("expected parser to be registered")
		}

		plugin, ok := parser.(*ChangelogParserPlugin)
		if !ok {
			t.Fatal("expected ChangelogParserPlugin type")
		}

		if !plugin.IsEnabled() {
			t.Error("expected plugin to be enabled")
		}
	})

	t.Run("unregister", func(t *testing.T) {
		ResetChangelogParser()

		cfg := &Config{Enabled: true}
		Register(cfg)

		Unregister()

		parser := GetChangelogParserFn()
		if parser != nil {
			t.Error("expected parser to be nil after unregister")
		}
	})

	t.Run("reset", func(t *testing.T) {
		cfg := &Config{Enabled: true}
		Register(cfg)

		ResetChangelogParser()

		parser := GetChangelogParserFn()
		if parser != nil {
			t.Error("expected parser to be nil after reset")
		}
	})

	t.Run("duplicate registration warning", func(t *testing.T) {
		ResetChangelogParser()

		// Register first parser
		cfg1 := &Config{Enabled: true}
		Register(cfg1)

		// Attempt to register second parser
		cfg2 := &Config{Enabled: true}
		Register(cfg2)

		// Should still have the first parser
		parser := GetChangelogParserFn()
		if parser == nil {
			t.Fatal("expected first parser to remain registered")
		}
	})
}

func TestChangelogParserPlugin_ErrorScenarios(t *testing.T) {
	// Save original and restore after test
	origOpenFile := openFileFn
	defer func() { openFileFn = origOpenFile }()

	t.Run("infer with IO error", func(t *testing.T) {
		openFileFn = func(name string) (*os.File, error) {
			return nil, errors.New("disk full")
		}

		cfg := &Config{
			Enabled:       true,
			InferBumpType: true,
		}
		plugin := NewChangelogParser(cfg)

		_, err := plugin.InferBumpType()
		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("validate with IO error", func(t *testing.T) {
		openFileFn = func(name string) (*os.File, error) {
			return nil, errors.New("disk full")
		}

		cfg := &Config{
			Enabled:                  true,
			RequireUnreleasedSection: true,
		}
		plugin := NewChangelogParser(cfg)

		err := plugin.ValidateHasEntries()
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}
