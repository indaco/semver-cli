package config

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"github.com/goccy/go-yaml"
)

type PluginConfig struct {
	CommitParser     bool                    `yaml:"commit-parser"`
	TagManager       *TagManagerConfig       `yaml:"tag-manager,omitempty"`
	VersionValidator *VersionValidatorConfig `yaml:"version-validator,omitempty"`
}

// TagManagerConfig holds configuration for the tag manager plugin.
type TagManagerConfig struct {
	// Enabled controls whether the plugin is active.
	Enabled bool `yaml:"enabled"`

	// AutoCreate automatically creates tags after version bumps.
	AutoCreate *bool `yaml:"auto-create,omitempty"`

	// Prefix is the tag prefix (default: "v").
	Prefix string `yaml:"prefix,omitempty"`

	// Annotate creates annotated tags instead of lightweight tags.
	Annotate *bool `yaml:"annotate,omitempty"`

	// Push automatically pushes tags to remote after creation.
	Push bool `yaml:"push,omitempty"`
}

// GetAutoCreate returns the auto-create setting with default true.
func (c *TagManagerConfig) GetAutoCreate() bool {
	if c.AutoCreate == nil {
		return true
	}
	return *c.AutoCreate
}

// GetAnnotate returns the annotate setting with default true.
func (c *TagManagerConfig) GetAnnotate() bool {
	if c.Annotate == nil {
		return true
	}
	return *c.Annotate
}

// GetPrefix returns the prefix with default "v".
func (c *TagManagerConfig) GetPrefix() string {
	if c.Prefix == "" {
		return "v"
	}
	return c.Prefix
}

// VersionValidatorConfig holds configuration for the version validator plugin.
type VersionValidatorConfig struct {
	// Enabled controls whether the plugin is active.
	Enabled bool `yaml:"enabled"`

	// Rules defines the validation rules to apply.
	Rules []ValidationRule `yaml:"rules,omitempty"`
}

// ValidationRule defines a single validation rule.
type ValidationRule struct {
	// Type is the rule type (e.g., "pre-release-format", "major-version-max").
	Type string `yaml:"type"`

	// Pattern is a regex pattern for format validation rules.
	Pattern string `yaml:"pattern,omitempty"`

	// Value is a numeric limit for max-version rules.
	Value int `yaml:"value,omitempty"`

	// Enabled controls whether this specific rule is active.
	Enabled bool `yaml:"enabled,omitempty"`

	// Branch is a glob pattern for branch-constraint rules.
	Branch string `yaml:"branch,omitempty"`

	// Allowed lists allowed bump types for branch-constraint rules.
	Allowed []string `yaml:"allowed,omitempty"`
}

type ExtensionConfig struct {
	Name    string `yaml:"name"`
	Path    string `yaml:"path"`
	Enabled bool   `yaml:"enabled"`
}

type PreReleaseHookConfig struct {
	Command string `yaml:"command,omitempty"`
}

// DiscoveryConfig configures automatic module discovery behavior.
type DiscoveryConfig struct {
	// Enabled controls whether auto-discovery is active (default: true).
	Enabled *bool `yaml:"enabled,omitempty"`

	// Recursive enables searching subdirectories (default: true).
	Recursive *bool `yaml:"recursive,omitempty"`

	// MaxDepth limits directory traversal depth (default: 10).
	MaxDepth *int `yaml:"max_depth,omitempty"`

	// Exclude lists paths/patterns to skip during discovery.
	Exclude []string `yaml:"exclude,omitempty"`
}

// ModuleConfig defines an explicitly configured module.
type ModuleConfig struct {
	// Name is the module identifier.
	Name string `yaml:"name"`

	// Path is the path to the module's .version file.
	Path string `yaml:"path"`

	// Enabled controls whether this module is active (default: true).
	Enabled *bool `yaml:"enabled,omitempty"`
}

// WorkspaceConfig configures multi-module/monorepo behavior.
type WorkspaceConfig struct {
	// Discovery configures automatic module discovery.
	Discovery *DiscoveryConfig `yaml:"discovery,omitempty"`

	// Modules explicitly defines modules (overrides discovery if non-empty).
	Modules []ModuleConfig `yaml:"modules,omitempty"`
}

type Config struct {
	Path            string                            `yaml:"path"`
	Plugins         *PluginConfig                     `yaml:"plugins,omitempty"`
	Extensions      []ExtensionConfig                 `yaml:"extensions,omitempty"`
	PreReleaseHooks []map[string]PreReleaseHookConfig `yaml:"pre-release-hooks,omitempty"`
	Workspace       *WorkspaceConfig                  `yaml:"workspace,omitempty"`
}

var (
	LoadConfigFn = loadConfig
	SaveConfigFn = saveConfig
	marshalFn    = yaml.Marshal
	openFileFn   = os.OpenFile
	writeFileFn  = func(file *os.File, data []byte) (int, error) {
		return file.Write(data)
	}
)

func loadConfig() (*Config, error) {
	// Highest priority: ENV variable
	if envPath := os.Getenv("SEMVER_PATH"); envPath != "" {
		return &Config{Path: envPath}, nil
	}

	// Second priority: YAML file
	data, err := os.ReadFile(".semver.yaml")
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // fallback to default
		}
		return nil, err
	}

	var cfg Config
	decoder := yaml.NewDecoder(bytes.NewReader(data), yaml.Strict())
	if err := decoder.Decode(&cfg); err != nil {
		return nil, err
	}

	if cfg.Path == "" {
		cfg.Path = ".version"
	}

	if cfg.Plugins == nil {
		cfg.Plugins = &PluginConfig{CommitParser: true}
	}

	return &cfg, nil
}

// NormalizeVersionPath ensures the path is a file, not just a directory.
func NormalizeVersionPath(path string) string {
	info, err := os.Stat(path)
	if err == nil && info.IsDir() {
		return filepath.Join(path, ".version")
	}

	// If it doesn't exist or is already a file, return as-is
	return path
}

// ConfigFilePerm defines secure file permissions for config files (owner read/write only).
const ConfigFilePerm = 0600

func saveConfig(cfg *Config) error {
	file, err := openFileFn(".semver.yaml", os.O_RDWR|os.O_CREATE|os.O_TRUNC, ConfigFilePerm)
	if err != nil {
		return fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

	data, err := marshalFn(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if _, err := writeFileFn(file, data); err != nil {
		return fmt.Errorf("failed to write config data: %w", err)
	}

	return nil
}

// DefaultExcludePatterns returns the default patterns to exclude during module discovery.
var DefaultExcludePatterns = []string{
	"node_modules",
	".git",
	"vendor",
	"tmp",
	"build",
	"dist",
	".cache",
	"__pycache__",
}

// DiscoveryDefaults returns a DiscoveryConfig with default values.
func DiscoveryDefaults() *DiscoveryConfig {
	enabled := true
	recursive := true
	maxDepth := 10

	return &DiscoveryConfig{
		Enabled:   &enabled,
		Recursive: &recursive,
		MaxDepth:  &maxDepth,
		Exclude:   DefaultExcludePatterns,
	}
}

// GetDiscoveryConfig returns the discovery configuration with defaults applied.
// If workspace or discovery is not configured, returns default discovery settings.
func (c *Config) GetDiscoveryConfig() *DiscoveryConfig {
	if c.Workspace == nil || c.Workspace.Discovery == nil {
		return DiscoveryDefaults()
	}

	cfg := c.Workspace.Discovery
	defaults := DiscoveryDefaults()

	// Apply defaults for nil pointer fields
	result := &DiscoveryConfig{
		Enabled:   cfg.Enabled,
		Recursive: cfg.Recursive,
		MaxDepth:  cfg.MaxDepth,
		Exclude:   cfg.Exclude,
	}

	if result.Enabled == nil {
		result.Enabled = defaults.Enabled
	}
	if result.Recursive == nil {
		result.Recursive = defaults.Recursive
	}
	if result.MaxDepth == nil {
		result.MaxDepth = defaults.MaxDepth
	}

	return result
}

// GetExcludePatterns returns the merged list of default and configured exclude patterns.
// Configured patterns are appended to defaults, allowing for extension.
func (c *Config) GetExcludePatterns() []string {
	discovery := c.GetDiscoveryConfig()

	// Start with defaults
	patterns := make([]string, len(DefaultExcludePatterns))
	copy(patterns, DefaultExcludePatterns)

	// Add configured patterns if they differ from defaults
	if c.Workspace != nil && c.Workspace.Discovery != nil && len(c.Workspace.Discovery.Exclude) > 0 {
		// Use a map to avoid duplicates
		seen := make(map[string]bool)
		for _, p := range DefaultExcludePatterns {
			seen[p] = true
		}

		for _, p := range c.Workspace.Discovery.Exclude {
			if !seen[p] {
				patterns = append(patterns, p)
				seen[p] = true
			}
		}
	} else if discovery.Exclude != nil {
		// If using defaults, return them directly
		return discovery.Exclude
	}

	return patterns
}

// HasExplicitModules returns true if modules are explicitly defined in the workspace configuration.
func (c *Config) HasExplicitModules() bool {
	return c.Workspace != nil && len(c.Workspace.Modules) > 0
}

// IsModuleEnabled checks if a specific module is enabled by name.
// Returns false if the module is not found or workspace is not configured.
func (c *Config) IsModuleEnabled(name string) bool {
	if !c.HasExplicitModules() {
		return false
	}

	for _, module := range c.Workspace.Modules {
		if module.Name == name {
			return module.IsEnabled()
		}
	}

	return false
}

// IsEnabled returns true if the module is enabled.
// Modules are enabled by default if the Enabled field is nil.
func (m *ModuleConfig) IsEnabled() bool {
	if m.Enabled == nil {
		return true
	}
	return *m.Enabled
}
