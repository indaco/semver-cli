package commitparser

import (
	"testing"
)

/* ------------------------------------------------------------------------- */
/* PLUGIN INTERFACE TESTS                                                    */
/* ------------------------------------------------------------------------- */

func TestDefaultCommitParser_Name(t *testing.T) {
	var p CommitParserPlugin
	if got := p.Name(); got != "commit-parser" {
		t.Errorf("expected plugin name 'commit-parser', got '%s'", got)
	}
}

func TestDefaultCommitParser_Description(t *testing.T) {
	var p CommitParserPlugin
	if got := p.Description(); got != "Parses conventional commits to infer bump type" {
		t.Errorf("expected plugin description, got '%s'", got)
	}
}

func TestDefaultCommitParser_Version(t *testing.T) {
	var p CommitParserPlugin
	if got := p.Version(); got != "v0.1.0" {
		t.Errorf("expected plugin version 'v0.1.0', got '%s'", got)
	}
}

/* ------------------------------------------------------------------------- */
/* COMMIT PARSER TESTS                                                       */
/* ------------------------------------------------------------------------- */

func TestCommitParser_Parse(t *testing.T) {
	parser := newCommitParser()

	tests := []struct {
		name         string
		commits      []string
		expectedBump string
		expectError  bool
	}{
		{"Single feat commit", []string{"feat: add new feature"}, "minor", false},
		{"Single fix commit", []string{"fix: bug fix"}, "patch", false},
		{"Breaking change in body", []string{"chore: refactor\n\nBREAKING CHANGE: API"}, "major", false},
		{"Multiple types, breaking wins", []string{"fix: bug", "feat: thing", "BREAKING CHANGE: yes"}, "major", false},
		{"Multiple types, feat wins", []string{"fix: bug", "feat: thing"}, "minor", false},
		{"Only unrelated", []string{"docs: update", "chore: clean"}, "", true},
		{"Empty list", []string{}, "", true},
		{"Case insensitive", []string{"Feat: Add X"}, "minor", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.Parse(tt.commits)
			if tt.expectError {
				if err == nil {
					t.Fatalf("expected error, got none")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tt.expectedBump {
				t.Errorf("expected %q, got %q", tt.expectedBump, result)
			}
		})
	}
}

func TestRegisterCommitParser(t *testing.T) {
	called := false

	original := RegisterCommitParserFn
	RegisterCommitParserFn = func(_ CommitParser) {
		called = true
	}
	defer func() { RegisterCommitParserFn = original }()

	Register()

	if !called {
		t.Errorf("expected RegisterCommitParser to be called")
	}
}
