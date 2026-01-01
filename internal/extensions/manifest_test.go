package extensions

import (
	"strings"
	"testing"
)

func TestExtensionManifest_Validate(t *testing.T) {
	base := ExtensionManifest{
		Name:        "commit-parser",
		Version:     "0.1.0",
		Description: "Parses conventional commits",
		Author:      "indaco",
		Repository:  "https://github.com/indaco/semver-commit-parser",
		Entry:       "github.com/indaco/semver-commit/parser",
	}

	tests := []struct {
		field    string
		modify   func(m *ExtensionManifest)
		expected string
	}{
		{"missing name", func(m *ExtensionManifest) { m.Name = "" }, "missing 'name'"},
		{"missing version", func(m *ExtensionManifest) { m.Version = "" }, "missing 'version'"},
		{"missing description", func(m *ExtensionManifest) { m.Description = "" }, "missing 'description'"},
		{"missing author", func(m *ExtensionManifest) { m.Author = "" }, "missing 'author'"},
		{"missing repository", func(m *ExtensionManifest) { m.Repository = "" }, "missing 'repository'"},
		{"missing entry", func(m *ExtensionManifest) { m.Entry = "" }, "missing 'entry'"},
	}

	for _, tt := range tests {
		t.Run(tt.field, func(t *testing.T) {
			m := base
			tt.modify(&m)

			err := m.ValidateManifest()
			if err == nil || !strings.Contains(err.Error(), tt.expected) {
				t.Errorf("expected error to contain %q, got %v", tt.expected, err)
			}
		})
	}

	t.Run("valid manifest", func(t *testing.T) {
		err := base.ValidateManifest()
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})
}

/* ------------------------------------------------------------------------- */
/* ADDITIONAL TABLE-DRIVEN TESTS FOR MANIFEST VALIDATION                   */
/* ------------------------------------------------------------------------- */

// TestExtensionManifest_ValidateManifest_TableDriven tests comprehensive validation scenarios
func TestExtensionManifest_ValidateManifest_TableDriven(t *testing.T) {
	tests := []struct {
		name        string
		manifest    ExtensionManifest
		wantErr     bool
		wantErrText string
	}{
		{
			name: "complete valid manifest",
			manifest: ExtensionManifest{
				Name:        "test-extension",
				Version:     "1.0.0",
				Description: "A test extension",
				Author:      "Test Author",
				Repository:  "https://github.com/test/repo",
				Entry:       "hook.sh",
				Hooks:       []string{"pre-bump", "post-bump"},
			},
			wantErr: false,
		},
		{
			name: "valid manifest without hooks",
			manifest: ExtensionManifest{
				Name:        "simple-ext",
				Version:     "0.1.0",
				Description: "Simple extension",
				Author:      "Developer",
				Repository:  "https://gitlab.com/dev/simple",
				Entry:       "run.sh",
				Hooks:       nil,
			},
			wantErr: false,
		},
		{
			name: "valid manifest with empty hooks array",
			manifest: ExtensionManifest{
				Name:        "no-hooks",
				Version:     "2.0.0",
				Description: "Extension without hooks",
				Author:      "Author",
				Repository:  "https://github.com/author/nohooks",
				Entry:       "script.py",
				Hooks:       []string{},
			},
			wantErr: false,
		},
		{
			name: "missing name only",
			manifest: ExtensionManifest{
				Name:        "",
				Version:     "1.0.0",
				Description: "Missing name",
				Author:      "Author",
				Repository:  "https://github.com/test/repo",
				Entry:       "hook.sh",
			},
			wantErr:     true,
			wantErrText: "missing 'name'",
		},
		{
			name: "missing version only",
			manifest: ExtensionManifest{
				Name:        "test",
				Version:     "",
				Description: "Missing version",
				Author:      "Author",
				Repository:  "https://github.com/test/repo",
				Entry:       "hook.sh",
			},
			wantErr:     true,
			wantErrText: "missing 'version'",
		},
		{
			name: "missing description only",
			manifest: ExtensionManifest{
				Name:        "test",
				Version:     "1.0.0",
				Description: "",
				Author:      "Author",
				Repository:  "https://github.com/test/repo",
				Entry:       "hook.sh",
			},
			wantErr:     true,
			wantErrText: "missing 'description'",
		},
		{
			name: "missing author only",
			manifest: ExtensionManifest{
				Name:        "test",
				Version:     "1.0.0",
				Description: "Test",
				Author:      "",
				Repository:  "https://github.com/test/repo",
				Entry:       "hook.sh",
			},
			wantErr:     true,
			wantErrText: "missing 'author'",
		},
		{
			name: "missing repository only",
			manifest: ExtensionManifest{
				Name:        "test",
				Version:     "1.0.0",
				Description: "Test",
				Author:      "Author",
				Repository:  "",
				Entry:       "hook.sh",
			},
			wantErr:     true,
			wantErrText: "missing 'repository'",
		},
		{
			name: "missing entry only",
			manifest: ExtensionManifest{
				Name:        "test",
				Version:     "1.0.0",
				Description: "Test",
				Author:      "Author",
				Repository:  "https://github.com/test/repo",
				Entry:       "",
			},
			wantErr:     true,
			wantErrText: "missing 'entry'",
		},
		{
			name: "all fields empty",
			manifest: ExtensionManifest{
				Name:        "",
				Version:     "",
				Description: "",
				Author:      "",
				Repository:  "",
				Entry:       "",
			},
			wantErr:     true,
			wantErrText: "missing 'name'", // Should fail on first check
		},
		{
			name: "whitespace only fields",
			manifest: ExtensionManifest{
				Name:        "   ",
				Version:     "1.0.0",
				Description: "Test",
				Author:      "Author",
				Repository:  "https://github.com/test/repo",
				Entry:       "hook.sh",
			},
			wantErr: false, // Currently doesn't trim/validate whitespace
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.manifest.ValidateManifest()

			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateManifest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.wantErrText != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErrText) {
					t.Errorf("expected error containing %q, got: %v", tt.wantErrText, err)
				}
			}
		})
	}
}

// TestExtensionManifest_Fields tests individual field properties
func TestExtensionManifest_Fields(t *testing.T) {
	tests := []struct {
		name     string
		manifest ExtensionManifest
		checkFn  func(t *testing.T, m ExtensionManifest)
	}{
		{
			name: "hooks with all valid types",
			manifest: ExtensionManifest{
				Name:        "multi-hook",
				Version:     "1.0.0",
				Description: "Multiple hooks",
				Author:      "Author",
				Repository:  "https://github.com/test/repo",
				Entry:       "hook.sh",
				Hooks:       []string{"pre-bump", "post-bump", "pre-release", "validate"},
			},
			checkFn: func(t *testing.T, m ExtensionManifest) {
				if len(m.Hooks) != 4 {
					t.Errorf("expected 4 hooks, got %d", len(m.Hooks))
				}
			},
		},
		{
			name: "single hook",
			manifest: ExtensionManifest{
				Name:        "single-hook",
				Version:     "1.0.0",
				Description: "Single hook",
				Author:      "Author",
				Repository:  "https://github.com/test/repo",
				Entry:       "hook.sh",
				Hooks:       []string{"pre-bump"},
			},
			checkFn: func(t *testing.T, m ExtensionManifest) {
				if len(m.Hooks) != 1 {
					t.Errorf("expected 1 hook, got %d", len(m.Hooks))
				}
				if m.Hooks[0] != "pre-bump" {
					t.Errorf("expected hook 'pre-bump', got %q", m.Hooks[0])
				}
			},
		},
		{
			name: "various version formats",
			manifest: ExtensionManifest{
				Name:        "version-test",
				Version:     "1.2.3-alpha.1+build.456",
				Description: "Version format test",
				Author:      "Author",
				Repository:  "https://github.com/test/repo",
				Entry:       "hook.sh",
			},
			checkFn: func(t *testing.T, m ExtensionManifest) {
				if m.Version != "1.2.3-alpha.1+build.456" {
					t.Errorf("version not preserved correctly: %s", m.Version)
				}
			},
		},
		{
			name: "different entry formats",
			manifest: ExtensionManifest{
				Name:        "entry-test",
				Version:     "1.0.0",
				Description: "Entry format test",
				Author:      "Author",
				Repository:  "https://github.com/test/repo",
				Entry:       "./scripts/main.py",
			},
			checkFn: func(t *testing.T, m ExtensionManifest) {
				if m.Entry != "./scripts/main.py" {
					t.Errorf("entry path not preserved: %s", m.Entry)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.manifest.ValidateManifest()
			if err != nil {
				t.Fatalf("unexpected validation error: %v", err)
			}
			tt.checkFn(t, tt.manifest)
		})
	}
}
