package auditlog

import (
	"encoding/json"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/goccy/go-yaml"
)

// MockGitOps implements GitOperations for testing.
type MockGitOps struct {
	AuthorFunc    func() (string, error)
	CommitSHAFunc func() (string, error)
	BranchFunc    func() (string, error)
}

func (m *MockGitOps) GetAuthor() (string, error) {
	if m.AuthorFunc != nil {
		return m.AuthorFunc()
	}
	return "Test User <test@example.com>", nil
}

func (m *MockGitOps) GetCommitSHA() (string, error) {
	if m.CommitSHAFunc != nil {
		return m.CommitSHAFunc()
	}
	return "abc1234567890def", nil
}

func (m *MockGitOps) GetBranch() (string, error) {
	if m.BranchFunc != nil {
		return m.BranchFunc()
	}
	return "main", nil
}

// MockFileOps implements FileOperations for testing.
type MockFileOps struct {
	data   map[string][]byte
	exists map[string]bool
}

func NewMockFileOps() *MockFileOps {
	return &MockFileOps{
		data:   make(map[string][]byte),
		exists: make(map[string]bool),
	}
}

func (m *MockFileOps) ReadFile(path string) ([]byte, error) {
	if data, ok := m.data[path]; ok {
		return data, nil
	}
	return nil, os.ErrNotExist
}

func (m *MockFileOps) WriteFile(path string, data []byte, perm os.FileMode) error {
	m.data[path] = data
	m.exists[path] = true
	return nil
}

func (m *MockFileOps) FileExists(path string) bool {
	return m.exists[path]
}

func TestNewAuditLog(t *testing.T) {
	tests := []struct {
		name           string
		config         *Config
		expectedPath   string
		expectedFormat string
	}{
		{
			name:           "nil config uses defaults",
			config:         nil,
			expectedPath:   ".version-history.json",
			expectedFormat: "json",
		},
		{
			name: "custom config",
			config: &Config{
				Enabled: true,
				Path:    "custom.json",
				Format:  "json",
			},
			expectedPath:   "custom.json",
			expectedFormat: "json",
		},
		{
			name: "yaml format",
			config: &Config{
				Enabled: true,
				Path:    "history.yaml",
				Format:  "yaml",
			},
			expectedPath:   "history.yaml",
			expectedFormat: "yaml",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plugin := NewAuditLog(tt.config)
			if plugin == nil {
				t.Fatal("expected plugin to be non-nil")
			}
			if plugin.GetConfig().GetPath() != tt.expectedPath {
				t.Errorf("expected path %q, got %q", tt.expectedPath, plugin.GetConfig().GetPath())
			}
			if plugin.GetConfig().GetFormat() != tt.expectedFormat {
				t.Errorf("expected format %q, got %q", tt.expectedFormat, plugin.GetConfig().GetFormat())
			}
		})
	}
}

func TestAuditLogPlugin_Metadata(t *testing.T) {
	plugin := NewAuditLog(DefaultConfig())

	if plugin.Name() != "audit-log" {
		t.Errorf("expected name 'audit-log', got %q", plugin.Name())
	}

	if plugin.Version() != "v0.1.0" {
		t.Errorf("expected version 'v0.1.0', got %q", plugin.Version())
	}

	if plugin.Description() == "" {
		t.Error("expected non-empty description")
	}
}

func TestAuditLogPlugin_IsEnabled(t *testing.T) {
	tests := []struct {
		name    string
		enabled bool
	}{
		{"enabled", true},
		{"disabled", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := DefaultConfig()
			cfg.Enabled = tt.enabled
			plugin := NewAuditLog(cfg)

			if plugin.IsEnabled() != tt.enabled {
				t.Errorf("expected IsEnabled() = %v, got %v", tt.enabled, plugin.IsEnabled())
			}
		})
	}
}

func TestAuditLogPlugin_RecordEntry_Disabled(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Enabled = false

	plugin := NewAuditLogWithOps(cfg, &MockGitOps{}, NewMockFileOps())

	entry := &Entry{
		PreviousVersion: "1.0.0",
		NewVersion:      "1.0.1",
		BumpType:        "patch",
	}

	err := plugin.RecordEntry(entry)
	if err != nil {
		t.Errorf("expected no error when disabled, got %v", err)
	}
}

func TestAuditLogPlugin_RecordEntry_JSON(t *testing.T) {
	cfg := &Config{
		Enabled:          true,
		Path:             ".version-history.json",
		Format:           "json",
		IncludeAuthor:    true,
		IncludeTimestamp: true,
		IncludeCommitSHA: true,
		IncludeBranch:    true,
	}

	mockGit := &MockGitOps{}
	mockFile := NewMockFileOps()
	plugin := NewAuditLogWithOps(cfg, mockGit, mockFile)

	// Fix time for consistent testing
	fixedTime := time.Date(2026, 1, 4, 12, 0, 0, 0, time.UTC)
	plugin.timeFunc = func() time.Time { return fixedTime }

	entry := &Entry{
		PreviousVersion: "1.0.0",
		NewVersion:      "1.0.1",
		BumpType:        "patch",
	}

	err := plugin.RecordEntry(entry)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify file was written
	data, ok := mockFile.data[cfg.Path]
	if !ok {
		t.Fatal("expected file to be written")
	}

	var logFile AuditLogFile
	if err := json.Unmarshal(data, &logFile); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if len(logFile.Entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(logFile.Entries))
	}

	actualEntry := logFile.Entries[0]
	if actualEntry.PreviousVersion != "1.0.0" {
		t.Errorf("expected previous version '1.0.0', got %q", actualEntry.PreviousVersion)
	}
	if actualEntry.NewVersion != "1.0.1" {
		t.Errorf("expected new version '1.0.1', got %q", actualEntry.NewVersion)
	}
	if actualEntry.BumpType != "patch" {
		t.Errorf("expected bump type 'patch', got %q", actualEntry.BumpType)
	}
	if actualEntry.Author != "Test User <test@example.com>" {
		t.Errorf("expected author 'Test User <test@example.com>', got %q", actualEntry.Author)
	}
	if actualEntry.CommitSHA != "abc1234567890def" {
		t.Errorf("expected commit SHA 'abc1234567890def', got %q", actualEntry.CommitSHA)
	}
	if actualEntry.Branch != "main" {
		t.Errorf("expected branch 'main', got %q", actualEntry.Branch)
	}
	if actualEntry.Timestamp != fixedTime.UTC().Format(time.RFC3339) {
		t.Errorf("expected timestamp %q, got %q", fixedTime.UTC().Format(time.RFC3339), actualEntry.Timestamp)
	}
}

func TestAuditLogPlugin_RecordEntry_YAML(t *testing.T) {
	cfg := &Config{
		Enabled:          true,
		Path:             ".version-history.yaml",
		Format:           "yaml",
		IncludeAuthor:    true,
		IncludeTimestamp: true,
		IncludeCommitSHA: true,
		IncludeBranch:    true,
	}

	mockGit := &MockGitOps{}
	mockFile := NewMockFileOps()
	plugin := NewAuditLogWithOps(cfg, mockGit, mockFile)

	entry := &Entry{
		PreviousVersion: "2.0.0",
		NewVersion:      "3.0.0",
		BumpType:        "major",
	}

	err := plugin.RecordEntry(entry)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify file was written
	data, ok := mockFile.data[cfg.Path]
	if !ok {
		t.Fatal("expected file to be written")
	}

	var logFile AuditLogFile
	if err := yaml.Unmarshal(data, &logFile); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if len(logFile.Entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(logFile.Entries))
	}

	actualEntry := logFile.Entries[0]
	if actualEntry.PreviousVersion != "2.0.0" {
		t.Errorf("expected previous version '2.0.0', got %q", actualEntry.PreviousVersion)
	}
	if actualEntry.NewVersion != "3.0.0" {
		t.Errorf("expected new version '3.0.0', got %q", actualEntry.NewVersion)
	}
	if actualEntry.BumpType != "major" {
		t.Errorf("expected bump type 'major', got %q", actualEntry.BumpType)
	}
}

func TestAuditLogPlugin_RecordEntry_MultipleEntries(t *testing.T) {
	cfg := &Config{
		Enabled:          true,
		Path:             ".version-history.json",
		Format:           "json",
		IncludeTimestamp: true,
	}

	mockGit := &MockGitOps{}
	mockFile := NewMockFileOps()
	plugin := NewAuditLogWithOps(cfg, mockGit, mockFile)

	// Create entries at different times
	times := []time.Time{
		time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC),
		time.Date(2026, 1, 2, 12, 0, 0, 0, time.UTC),
		time.Date(2026, 1, 3, 12, 0, 0, 0, time.UTC),
	}

	entries := []*Entry{
		{PreviousVersion: "1.0.0", NewVersion: "1.0.1", BumpType: "patch"},
		{PreviousVersion: "1.0.1", NewVersion: "1.1.0", BumpType: "minor"},
		{PreviousVersion: "1.1.0", NewVersion: "2.0.0", BumpType: "major"},
	}

	for i, entry := range entries {
		plugin.timeFunc = func() time.Time { return times[i] }
		if err := plugin.RecordEntry(entry); err != nil {
			t.Fatalf("unexpected error on entry %d: %v", i, err)
		}
	}

	// Verify entries are sorted newest first
	data := mockFile.data[cfg.Path]
	var logFile AuditLogFile
	if err := json.Unmarshal(data, &logFile); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if len(logFile.Entries) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(logFile.Entries))
	}

	// Newest first
	if logFile.Entries[0].NewVersion != "2.0.0" {
		t.Errorf("expected first entry to be newest (2.0.0), got %q", logFile.Entries[0].NewVersion)
	}
	if logFile.Entries[2].NewVersion != "1.0.1" {
		t.Errorf("expected last entry to be oldest (1.0.1), got %q", logFile.Entries[2].NewVersion)
	}
}

func TestAuditLogPlugin_RecordEntry_GitError(t *testing.T) {
	cfg := &Config{
		Enabled:          true,
		Path:             ".version-history.json",
		Format:           "json",
		IncludeAuthor:    true,
		IncludeCommitSHA: true,
		IncludeBranch:    true,
	}

	mockGit := &MockGitOps{
		AuthorFunc: func() (string, error) {
			return "", errors.New("git error")
		},
		CommitSHAFunc: func() (string, error) {
			return "", errors.New("git error")
		},
		BranchFunc: func() (string, error) {
			return "", errors.New("git error")
		},
	}
	mockFile := NewMockFileOps()
	plugin := NewAuditLogWithOps(cfg, mockGit, mockFile)

	entry := &Entry{
		PreviousVersion: "1.0.0",
		NewVersion:      "1.0.1",
		BumpType:        "patch",
	}

	// Should not fail even if git operations fail
	err := plugin.RecordEntry(entry)
	if err != nil {
		t.Errorf("expected no error when git operations fail, got %v", err)
	}

	// Verify entry was still written with basic info
	data := mockFile.data[cfg.Path]
	var logFile AuditLogFile
	if err := json.Unmarshal(data, &logFile); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if len(logFile.Entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(logFile.Entries))
	}

	// Git fields should be empty
	actualEntry := logFile.Entries[0]
	if actualEntry.Author != "" {
		t.Errorf("expected empty author when git fails, got %q", actualEntry.Author)
	}
	if actualEntry.CommitSHA != "" {
		t.Errorf("expected empty commit SHA when git fails, got %q", actualEntry.CommitSHA)
	}
	if actualEntry.Branch != "" {
		t.Errorf("expected empty branch when git fails, got %q", actualEntry.Branch)
	}
}

func TestAuditLogPlugin_RecordEntry_SelectiveMetadata(t *testing.T) {
	tests := []struct {
		name          string
		includeAuthor bool
		includeSHA    bool
		includeBranch bool
		includeTime   bool
	}{
		{"all disabled", false, false, false, false},
		{"only author", true, false, false, false},
		{"only sha", false, true, false, false},
		{"only branch", false, false, true, false},
		{"only timestamp", false, false, false, true},
		{"author and sha", true, true, false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry := recordEntryWithConfig(t, tt.includeAuthor, tt.includeSHA, tt.includeBranch, tt.includeTime)
			verifyMetadataInclusion(t, entry, tt.includeAuthor, tt.includeSHA, tt.includeBranch, tt.includeTime)
		})
	}
}

func recordEntryWithConfig(t *testing.T, includeAuthor, includeSHA, includeBranch, includeTime bool) Entry {
	t.Helper()
	cfg := &Config{
		Enabled:          true,
		Path:             ".version-history.json",
		Format:           "json",
		IncludeAuthor:    includeAuthor,
		IncludeCommitSHA: includeSHA,
		IncludeBranch:    includeBranch,
		IncludeTimestamp: includeTime,
	}

	mockGit := &MockGitOps{}
	mockFile := NewMockFileOps()
	plugin := NewAuditLogWithOps(cfg, mockGit, mockFile)

	entry := &Entry{
		PreviousVersion: "1.0.0",
		NewVersion:      "1.0.1",
		BumpType:        "patch",
	}

	if err := plugin.RecordEntry(entry); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data := mockFile.data[cfg.Path]
	var logFile AuditLogFile
	if err := json.Unmarshal(data, &logFile); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	return logFile.Entries[0]
}

func verifyMetadataInclusion(t *testing.T, entry Entry, includeAuthor, includeSHA, includeBranch, includeTime bool) {
	t.Helper()
	verifyField(t, "author", entry.Author, includeAuthor)
	verifyField(t, "commit SHA", entry.CommitSHA, includeSHA)
	verifyField(t, "branch", entry.Branch, includeBranch)
	verifyField(t, "timestamp", entry.Timestamp, includeTime)
}

func verifyField(t *testing.T, fieldName, fieldValue string, shouldInclude bool) {
	t.Helper()
	if shouldInclude && fieldValue == "" {
		t.Errorf("expected %s to be included", fieldName)
	}
	if !shouldInclude && fieldValue != "" {
		t.Errorf("expected %s to be excluded, got %q", fieldName, fieldValue)
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Enabled {
		t.Error("expected default enabled to be false")
	}
	if cfg.GetPath() != ".version-history.json" {
		t.Errorf("expected default path '.version-history.json', got %q", cfg.GetPath())
	}
	if cfg.GetFormat() != "json" {
		t.Errorf("expected default format 'json', got %q", cfg.GetFormat())
	}
	if !cfg.IncludeAuthor {
		t.Error("expected default include-author to be true")
	}
	if !cfg.IncludeTimestamp {
		t.Error("expected default include-timestamp to be true")
	}
	if !cfg.IncludeCommitSHA {
		t.Error("expected default include-commit-sha to be true")
	}
	if !cfg.IncludeBranch {
		t.Error("expected default include-branch to be true")
	}
}

func TestConfig_GetPath(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{"empty path uses default", "", ".version-history.json"},
		{"custom path", "custom.json", "custom.json"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{Path: tt.path}
			if cfg.GetPath() != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, cfg.GetPath())
			}
		})
	}
}

func TestConfig_GetFormat(t *testing.T) {
	tests := []struct {
		name     string
		format   string
		expected string
	}{
		{"empty format uses default", "", "json"},
		{"json format", "json", "json"},
		{"yaml format", "yaml", "yaml"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{Format: tt.format}
			if cfg.GetFormat() != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, cfg.GetFormat())
			}
		})
	}
}
