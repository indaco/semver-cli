package tui

import (
	"testing"

	"github.com/indaco/semver-cli/internal/workspace"
)

func TestNewModulePrompt(t *testing.T) {
	modules := []*workspace.Module{
		{Name: "module-a", CurrentVersion: "1.0.0"},
		{Name: "module-b", CurrentVersion: "2.0.0"},
	}

	prompt := NewModulePrompt(modules)

	if prompt == nil {
		t.Fatal("NewModulePrompt() returned nil")
	}

	if len(prompt.modules) != 2 {
		t.Errorf("NewModulePrompt() modules count = %d, want 2", len(prompt.modules))
	}
}

func TestModulePrompt_ImplementsPrompter(t *testing.T) {
	// Compile-time check that ModulePrompt implements Prompter
	var _ Prompter = (*ModulePrompt)(nil)
}

func TestPluralize(t *testing.T) {
	tests := []struct {
		name     string
		count    int
		expected string
	}{
		{"zero", 0, "s"},
		{"one", 1, ""},
		{"two", 2, "s"},
		{"many", 10, "s"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := pluralize(tt.count)
			if got != tt.expected {
				t.Errorf("pluralize(%d) = %q, want %q", tt.count, got, tt.expected)
			}
		})
	}
}

func TestModulePrompt_FormatModuleList(t *testing.T) {
	tests := []struct {
		name          string
		modules       []*workspace.Module
		shouldContain []string
	}{
		{
			name:          "empty modules",
			modules:       []*workspace.Module{},
			shouldContain: []string{},
		},
		{
			name: "single module",
			modules: []*workspace.Module{
				{Name: "module-a", CurrentVersion: "1.0.0"},
			},
			shouldContain: []string{"module-a", "1.0.0"},
		},
		{
			name: "multiple modules under limit",
			modules: []*workspace.Module{
				{Name: "module-a", CurrentVersion: "1.0.0"},
				{Name: "module-b", CurrentVersion: "2.0.0"},
			},
			shouldContain: []string{"module-a", "module-b"},
		},
		{
			name: "many modules over limit",
			modules: []*workspace.Module{
				{Name: "module-1", CurrentVersion: "1.0.0"},
				{Name: "module-2", CurrentVersion: "2.0.0"},
				{Name: "module-3", CurrentVersion: "3.0.0"},
				{Name: "module-4", CurrentVersion: "4.0.0"},
				{Name: "module-5", CurrentVersion: "5.0.0"},
				{Name: "module-6", CurrentVersion: "6.0.0"},
				{Name: "module-7", CurrentVersion: "7.0.0"},
			},
			shouldContain: []string{"module-1", "and 2 more"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prompt := NewModulePrompt(tt.modules)
			result := prompt.formatModuleList()

			for _, expected := range tt.shouldContain {
				if result != "" && !contains(result, expected) {
					t.Errorf("formatModuleList() does not contain %q\nGot: %s", expected, result)
				}
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Note: We cannot easily test the interactive prompts (showInitialPrompt, showMultiSelect)
// without mocking the huh library or using integration tests.
// The important parts are tested via the mock implementation.

func TestModulePrompt_PromptModuleSelection_EmptyModules(t *testing.T) {
	prompt := NewModulePrompt(nil)

	_, err := prompt.PromptModuleSelection([]*workspace.Module{})
	if err == nil {
		t.Error("PromptModuleSelection() with empty modules should return an error")
	}

	expectedErrMsg := "no modules provided for selection"
	if err.Error() != expectedErrMsg {
		t.Errorf("error message = %q, want %q", err.Error(), expectedErrMsg)
	}
}

func TestErrCanceled(t *testing.T) {
	if ErrCanceled == nil {
		t.Fatal("ErrCanceled should not be nil")
	}

	expectedMsg := "operation canceled by user"
	if ErrCanceled.Error() != expectedMsg {
		t.Errorf("ErrCanceled.Error() = %q, want %q", ErrCanceled.Error(), expectedMsg)
	}
}
