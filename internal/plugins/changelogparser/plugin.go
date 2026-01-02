package changelogparser

import (
	"errors"
	"fmt"
)

// ChangelogParser defines the interface for parsing changelog files
// to infer version bumps and validate changelog completeness.
type ChangelogInferrer interface {
	Name() string
	InferBumpType() (string, error)
	ValidateHasEntries() error
}

// ChangelogParserPlugin implements the ChangelogInferrer interface.
type ChangelogParserPlugin struct {
	config *Config
}

// Config holds configuration for the changelog parser plugin.
type Config struct {
	// Enabled controls whether the plugin is active.
	Enabled bool

	// Path is the path to the changelog file (default: "CHANGELOG.md").
	Path string

	// RequireUnreleasedSection enforces presence of Unreleased section.
	RequireUnreleasedSection bool

	// InferBumpType enables automatic bump type inference from changelog.
	InferBumpType bool

	// Priority determines which parser takes precedence: "changelog" or "commits"
	// When set to "changelog", changelog-based inference overrides commit-based inference.
	Priority string
}

// Ensure ChangelogParserPlugin implements the plugin interface.
var _ ChangelogInferrer = (*ChangelogParserPlugin)(nil)

func (p *ChangelogParserPlugin) Name() string { return "changelog-parser" }
func (p *ChangelogParserPlugin) Description() string {
	return "Parses CHANGELOG.md to infer bump type and validate changelog completeness"
}
func (p *ChangelogParserPlugin) Version() string { return "v0.1.0" }

// NewChangelogParser creates a new changelog parser plugin with the given configuration.
func NewChangelogParser(cfg *Config) *ChangelogParserPlugin {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	// Apply defaults
	if cfg.Path == "" {
		cfg.Path = "CHANGELOG.md"
	}
	if cfg.Priority == "" {
		cfg.Priority = "changelog"
	}
	return &ChangelogParserPlugin{config: cfg}
}

// DefaultConfig returns the default changelog parser configuration.
func DefaultConfig() *Config {
	return &Config{
		Enabled:                  false,
		Path:                     "CHANGELOG.md",
		RequireUnreleasedSection: true,
		InferBumpType:            true,
		Priority:                 "changelog",
	}
}

// IsEnabled returns whether the plugin is active.
func (p *ChangelogParserPlugin) IsEnabled() bool {
	return p.config.Enabled
}

// GetConfig returns the plugin configuration.
func (p *ChangelogParserPlugin) GetConfig() *Config {
	return p.config
}

// InferBumpType parses the changelog and infers the bump type.
func (p *ChangelogParserPlugin) InferBumpType() (string, error) {
	if !p.IsEnabled() || !p.config.InferBumpType {
		return "", errors.New("changelog parser not enabled or inference disabled")
	}

	parser := newChangelogFileParser(p.config.Path)
	section, err := parser.ParseUnreleased()
	if err != nil {
		return "", fmt.Errorf("failed to parse unreleased section: %w", err)
	}

	bumpType, err := section.InferBumpType()
	if err != nil {
		return "", fmt.Errorf("failed to infer bump type: %w", err)
	}

	return bumpType, nil
}

// ValidateHasEntries validates that the Unreleased section has entries.
func (p *ChangelogParserPlugin) ValidateHasEntries() error {
	if !p.IsEnabled() || !p.config.RequireUnreleasedSection {
		return nil
	}

	parser := newChangelogFileParser(p.config.Path)
	section, err := parser.ParseUnreleased()
	if err != nil {
		return fmt.Errorf("changelog validation failed: %w", err)
	}

	if !section.HasEntries {
		return errors.New("changelog validation failed: Unreleased section has no entries")
	}

	return nil
}

// ShouldTakePrecedence returns true if changelog parser should take precedence over commit parser.
func (p *ChangelogParserPlugin) ShouldTakePrecedence() bool {
	return p.IsEnabled() && p.config.Priority == "changelog"
}
