package changeloggenerator

import (
	"testing"

	"github.com/indaco/semver-cli/internal/config"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Enabled {
		t.Error("expected Enabled to be false by default")
	}
	if cfg.Mode != "versioned" {
		t.Errorf("Mode = %q, want 'versioned'", cfg.Mode)
	}
	if cfg.ChangesDir != ".changes" {
		t.Errorf("ChangesDir = %q, want '.changes'", cfg.ChangesDir)
	}
	if cfg.ChangelogPath != "CHANGELOG.md" {
		t.Errorf("ChangelogPath = %q, want 'CHANGELOG.md'", cfg.ChangelogPath)
	}
	if cfg.Repository == nil {
		t.Fatal("expected Repository to be non-nil")
	}
	if !cfg.Repository.AutoDetect {
		t.Error("expected Repository.AutoDetect to be true by default")
	}
	if len(cfg.Groups) == 0 {
		t.Error("expected Groups to have default values")
	}
	if len(cfg.ExcludePatterns) == 0 {
		t.Error("expected ExcludePatterns to have default values")
	}
	if cfg.Contributors == nil {
		t.Fatal("expected Contributors to be non-nil")
	}
	if !cfg.Contributors.Enabled {
		t.Error("expected Contributors.Enabled to be true by default")
	}
}

func TestDefaultGroups(t *testing.T) {
	groups := DefaultGroups()

	if len(groups) == 0 {
		t.Fatal("expected at least one default group")
	}

	// Check that feat and fix are in the groups
	patterns := make(map[string]string)
	for _, g := range groups {
		patterns[g.Pattern] = g.Label
	}

	if _, ok := patterns["^feat"]; !ok {
		t.Error("expected '^feat' pattern in default groups")
	}
	if _, ok := patterns["^fix"]; !ok {
		t.Error("expected '^fix' pattern in default groups")
	}
}

func TestDefaultExcludePatterns(t *testing.T) {
	patterns := DefaultExcludePatterns()

	if len(patterns) == 0 {
		t.Fatal("expected at least one default exclude pattern")
	}

	// Check for common patterns
	found := make(map[string]bool)
	for _, p := range patterns {
		found[p] = true
	}

	if !found["^Merge"] {
		t.Error("expected '^Merge' in default exclude patterns")
	}
	if !found["^WIP"] {
		t.Error("expected '^WIP' in default exclude patterns")
	}
}

func TestFromConfigStruct_Nil(t *testing.T) {
	cfg := FromConfigStruct(nil)

	if cfg == nil {
		t.Fatal("expected non-nil config from nil input")
	}
	if cfg.Mode != "versioned" {
		t.Errorf("Mode = %q, want 'versioned'", cfg.Mode)
	}
}

func TestFromConfigStruct_Full(t *testing.T) {
	input := &config.ChangelogGeneratorConfig{
		Enabled:       true,
		Mode:          "unified",
		ChangesDir:    "custom-changes",
		ChangelogPath: "CHANGES.md",
		Repository: &config.RepositoryConfig{
			Provider: "gitlab",
			Host:     "gitlab.com",
			Owner:    "mygroup",
			Repo:     "myproject",
		},
		Groups: []config.CommitGroupConfig{
			{Pattern: "^feat", Label: "Features", Icon: "rocket", Order: 0},
			{Pattern: "^fix", Label: "Bug Fixes", Order: 1},
		},
		ExcludePatterns: []string{"^WIP", "^SKIP"},
		Contributors: &config.ContributorsConfig{
			Enabled: false,
			Format:  "custom format",
		},
	}

	cfg := FromConfigStruct(input)

	if !cfg.Enabled {
		t.Error("expected Enabled to be true")
	}
	if cfg.Mode != "unified" {
		t.Errorf("Mode = %q, want 'unified'", cfg.Mode)
	}
	if cfg.ChangesDir != "custom-changes" {
		t.Errorf("ChangesDir = %q, want 'custom-changes'", cfg.ChangesDir)
	}
	if cfg.ChangelogPath != "CHANGES.md" {
		t.Errorf("ChangelogPath = %q, want 'CHANGES.md'", cfg.ChangelogPath)
	}

	// Repository
	if cfg.Repository == nil {
		t.Fatal("expected Repository to be non-nil")
	}
	if cfg.Repository.Provider != "gitlab" {
		t.Errorf("Repository.Provider = %q, want 'gitlab'", cfg.Repository.Provider)
	}
	if cfg.Repository.Host != "gitlab.com" {
		t.Errorf("Repository.Host = %q, want 'gitlab.com'", cfg.Repository.Host)
	}
	if cfg.Repository.Owner != "mygroup" {
		t.Errorf("Repository.Owner = %q, want 'mygroup'", cfg.Repository.Owner)
	}

	// Groups
	if len(cfg.Groups) != 2 {
		t.Errorf("expected 2 groups, got %d", len(cfg.Groups))
	}
	if cfg.Groups[0].Icon != "rocket" {
		t.Errorf("Groups[0].Icon = %q, want 'rocket'", cfg.Groups[0].Icon)
	}

	// ExcludePatterns
	if len(cfg.ExcludePatterns) != 2 {
		t.Errorf("expected 2 exclude patterns, got %d", len(cfg.ExcludePatterns))
	}

	// Contributors
	if cfg.Contributors == nil {
		t.Fatal("expected Contributors to be non-nil")
	}
	if cfg.Contributors.Enabled {
		t.Error("expected Contributors.Enabled to be false")
	}
	if cfg.Contributors.Format != "custom format" {
		t.Errorf("Contributors.Format = %q, want 'custom format'", cfg.Contributors.Format)
	}
}

func TestFromConfigStruct_Defaults(t *testing.T) {
	// Test with minimal config - should use defaults
	input := &config.ChangelogGeneratorConfig{
		Enabled: true,
	}

	cfg := FromConfigStruct(input)

	// Should use default mode
	if cfg.Mode != "versioned" {
		t.Errorf("Mode = %q, want 'versioned' (default)", cfg.Mode)
	}

	// Should use default changes dir
	if cfg.ChangesDir != ".changes" {
		t.Errorf("ChangesDir = %q, want '.changes' (default)", cfg.ChangesDir)
	}

	// Should have default groups
	if len(cfg.Groups) == 0 {
		t.Error("expected default groups")
	}

	// Should have default exclude patterns
	if len(cfg.ExcludePatterns) == 0 {
		t.Error("expected default exclude patterns")
	}

	// Repository should default to AutoDetect
	if cfg.Repository == nil {
		t.Fatal("expected Repository to be non-nil")
	}
	if !cfg.Repository.AutoDetect {
		t.Error("expected Repository.AutoDetect to be true by default")
	}
}

func TestGroupConfig(t *testing.T) {
	g := GroupConfig{
		Pattern: "^feat",
		Label:   "Features",
		Icon:    "rocket",
		Order:   0,
	}

	if g.Pattern != "^feat" {
		t.Errorf("Pattern = %q, want '^feat'", g.Pattern)
	}
	if g.Label != "Features" {
		t.Errorf("Label = %q, want 'Features'", g.Label)
	}
	if g.Icon != "rocket" {
		t.Errorf("Icon = %q, want 'rocket'", g.Icon)
	}
	if g.Order != 0 {
		t.Errorf("Order = %d, want 0", g.Order)
	}
}

func TestContributorsConfig(t *testing.T) {
	c := ContributorsConfig{
		Enabled: true,
		Format:  "custom",
	}

	if !c.Enabled {
		t.Error("expected Enabled to be true")
	}
	if c.Format != "custom" {
		t.Errorf("Format = %q, want 'custom'", c.Format)
	}
}

func TestRepositoryConfig(t *testing.T) {
	r := RepositoryConfig{
		Provider:   "github",
		Host:       "github.com",
		Owner:      "owner",
		Repo:       "repo",
		AutoDetect: false,
	}

	if r.Provider != "github" {
		t.Errorf("Provider = %q, want 'github'", r.Provider)
	}
	if r.Host != "github.com" {
		t.Errorf("Host = %q, want 'github.com'", r.Host)
	}
	if r.Owner != "owner" {
		t.Errorf("Owner = %q, want 'owner'", r.Owner)
	}
	if r.Repo != "repo" {
		t.Errorf("Repo = %q, want 'repo'", r.Repo)
	}
	if r.AutoDetect {
		t.Error("expected AutoDetect to be false")
	}
}
