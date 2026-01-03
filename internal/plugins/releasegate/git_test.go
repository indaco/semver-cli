package releasegate

import (
	"errors"
	"os/exec"
	"testing"
)

func TestIsWorktreeClean(t *testing.T) {
	tests := []struct {
		name        string
		gitOutput   string
		gitErr      error
		wantClean   bool
		wantErr     bool
		errContains string
	}{
		{
			name:      "clean worktree",
			gitOutput: "",
			gitErr:    nil,
			wantClean: true,
			wantErr:   false,
		},
		{
			name:      "dirty worktree with modified files",
			gitOutput: "M file.txt\n",
			gitErr:    nil,
			wantClean: false,
			wantErr:   false,
		},
		{
			name:      "dirty worktree with untracked files",
			gitOutput: "?? newfile.txt\n",
			gitErr:    nil,
			wantClean: false,
			wantErr:   false,
		},
		{
			name:        "git error",
			gitOutput:   "",
			gitErr:      errors.New("not a git repository"),
			wantClean:   false,
			wantErr:     true,
			errContains: "failed to check git status",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original execCommand
			origExec := execCommand
			defer func() { execCommand = origExec }()

			// Mock execCommand
			execCommand = func(name string, args ...string) *exec.Cmd {
				cmd := &mockCmd{
					stdout: tt.gitOutput,
					stderr: "",
					err:    tt.gitErr,
				}
				return cmd.toExecCmd()
			}

			clean, err := isWorktreeClean()

			if tt.wantErr {
				if err == nil {
					t.Error("isWorktreeClean() expected error, got nil")
				} else if tt.errContains != "" && !containsString(err.Error(), tt.errContains) {
					t.Errorf("isWorktreeClean() error = %q, want to contain %q", err.Error(), tt.errContains)
				}
			} else {
				if err != nil {
					t.Errorf("isWorktreeClean() unexpected error: %v", err)
				}
			}

			if clean != tt.wantClean {
				t.Errorf("isWorktreeClean() = %v, want %v", clean, tt.wantClean)
			}
		})
	}
}

func TestGetCurrentBranch(t *testing.T) {
	tests := []struct {
		name        string
		gitOutput   string
		gitErr      error
		want        string
		wantErr     bool
		errContains string
	}{
		{
			name:      "main branch",
			gitOutput: "main\n",
			gitErr:    nil,
			want:      "main",
			wantErr:   false,
		},
		{
			name:      "feature branch",
			gitOutput: "feature/test\n",
			gitErr:    nil,
			want:      "feature/test",
			wantErr:   false,
		},
		{
			name:        "git error",
			gitOutput:   "",
			gitErr:      errors.New("not a git repository"),
			want:        "",
			wantErr:     true,
			errContains: "failed to get current branch",
		},
		{
			name:        "empty output",
			gitOutput:   "",
			gitErr:      nil,
			want:        "",
			wantErr:     true,
			errContains: "failed to determine current branch",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original execCommand
			origExec := execCommand
			defer func() { execCommand = origExec }()

			// Mock execCommand
			execCommand = func(name string, args ...string) *exec.Cmd {
				cmd := &mockCmd{
					stdout: tt.gitOutput,
					stderr: "",
					err:    tt.gitErr,
				}
				return cmd.toExecCmd()
			}

			branch, err := getCurrentBranch()

			if tt.wantErr {
				if err == nil {
					t.Error("getCurrentBranch() expected error, got nil")
				} else if tt.errContains != "" && !containsString(err.Error(), tt.errContains) {
					t.Errorf("getCurrentBranch() error = %q, want to contain %q", err.Error(), tt.errContains)
				}
			} else {
				if err != nil {
					t.Errorf("getCurrentBranch() unexpected error: %v", err)
				}
			}

			if branch != tt.want {
				t.Errorf("getCurrentBranch() = %q, want %q", branch, tt.want)
			}
		})
	}
}

func TestGetRecentCommits(t *testing.T) {
	tests := []struct {
		name        string
		count       int
		gitOutput   string
		gitErr      error
		want        []string
		wantErr     bool
		errContains string
	}{
		{
			name:  "multiple commits",
			count: 5,
			gitOutput: `abc123 feat: add feature
def456 fix: resolve bug
ghi789 chore: update deps`,
			gitErr: nil,
			want: []string{
				"abc123 feat: add feature",
				"def456 fix: resolve bug",
				"ghi789 chore: update deps",
			},
			wantErr: false,
		},
		{
			name:      "single commit",
			count:     1,
			gitOutput: "abc123 Initial commit",
			gitErr:    nil,
			want:      []string{"abc123 Initial commit"},
			wantErr:   false,
		},
		{
			name:      "no commits",
			count:     5,
			gitOutput: "",
			gitErr:    nil,
			want:      []string{},
			wantErr:   false,
		},
		{
			name:        "git error",
			count:       5,
			gitOutput:   "",
			gitErr:      errors.New("not a git repository"),
			want:        nil,
			wantErr:     true,
			errContains: "failed to get commit history",
		},
		{
			name:  "zero count defaults to 10",
			count: 0,
			gitOutput: `abc123 feat: add feature
def456 fix: resolve bug`,
			gitErr: nil,
			want: []string{
				"abc123 feat: add feature",
				"def456 fix: resolve bug",
			},
			wantErr: false,
		},
		{
			name:  "negative count defaults to 10",
			count: -1,
			gitOutput: `abc123 feat: add feature
def456 fix: resolve bug`,
			gitErr: nil,
			want: []string{
				"abc123 feat: add feature",
				"def456 fix: resolve bug",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original execCommand
			origExec := execCommand
			defer func() { execCommand = origExec }()

			// Mock execCommand
			execCommand = func(name string, args ...string) *exec.Cmd {
				cmd := &mockCmd{
					stdout: tt.gitOutput,
					stderr: "",
					err:    tt.gitErr,
				}
				return cmd.toExecCmd()
			}

			commits, err := getRecentCommits(tt.count)

			if tt.wantErr {
				if err == nil {
					t.Error("getRecentCommits() expected error, got nil")
				} else if tt.errContains != "" && !containsString(err.Error(), tt.errContains) {
					t.Errorf("getRecentCommits() error = %q, want to contain %q", err.Error(), tt.errContains)
				}
			} else {
				if err != nil {
					t.Errorf("getRecentCommits() unexpected error: %v", err)
				}
			}

			if len(commits) != len(tt.want) {
				t.Errorf("getRecentCommits() returned %d commits, want %d", len(commits), len(tt.want))
				return
			}

			for i, commit := range commits {
				if commit != tt.want[i] {
					t.Errorf("getRecentCommits()[%d] = %q, want %q", i, commit, tt.want[i])
				}
			}
		})
	}
}

// mockCmd simulates exec.Cmd for testing.
type mockCmd struct {
	stdout string
	stderr string
	err    error
}

// toExecCmd creates a real exec.Cmd that simulates the mock behavior.
func (m *mockCmd) toExecCmd() *exec.Cmd {
	// Use a helper script that returns our mock data
	// For testing, we'll use sh -c to echo our outputs
	var cmd *exec.Cmd
	if m.err != nil {
		// Create a command that will fail
		cmd = exec.Command("sh", "-c", "echo '"+m.stderr+"' >&2; exit 1")
	} else {
		// Create a command that succeeds
		cmd = exec.Command("sh", "-c", "echo '"+m.stdout+"'")
	}

	return cmd
}
