package pathutil

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/indaco/semver-cli/internal/apperrors"
)

func TestValidatePath(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name    string
		path    string
		baseDir string
		wantErr bool
		errType error
	}{
		{
			name:    "valid path without base",
			path:    "file.txt",
			baseDir: "",
			wantErr: false,
		},
		{
			name:    "valid absolute path without base",
			path:    "/tmp/file.txt",
			baseDir: "",
			wantErr: false,
		},
		{
			name:    "valid path within base",
			path:    filepath.Join(tmpDir, "subdir", "file.txt"),
			baseDir: tmpDir,
			wantErr: false,
		},
		{
			name:    "empty path",
			path:    "",
			baseDir: "",
			wantErr: true,
			errType: &apperrors.PathValidationError{},
		},
		{
			name:    "path traversal attempt",
			path:    filepath.Join(tmpDir, "..", "secret"),
			baseDir: tmpDir,
			wantErr: true,
			errType: &apperrors.PathValidationError{},
		},
		{
			name:    "path outside base dir",
			path:    "/etc/passwd",
			baseDir: tmpDir,
			wantErr: true,
			errType: &apperrors.PathValidationError{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ValidatePath(tt.path, tt.baseDir)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				if tt.errType != nil {
					var pathErr *apperrors.PathValidationError
					if !errors.As(err, &pathErr) {
						t.Errorf("expected PathValidationError, got %T", err)
					}
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if result == "" {
					t.Error("expected non-empty result")
				}
			}
		})
	}
}

func TestIsWithinDir(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name     string
		path     string
		dir      string
		expected bool
	}{
		{
			name:     "file within dir",
			path:     filepath.Join(tmpDir, "file.txt"),
			dir:      tmpDir,
			expected: true,
		},
		{
			name:     "subdir within dir",
			path:     filepath.Join(tmpDir, "subdir", "file.txt"),
			dir:      tmpDir,
			expected: true,
		},
		{
			name:     "dir itself",
			path:     tmpDir,
			dir:      tmpDir,
			expected: true,
		},
		{
			name:     "path outside dir",
			path:     "/etc/passwd",
			dir:      tmpDir,
			expected: false,
		},
		{
			name:     "parent of dir",
			path:     filepath.Dir(tmpDir),
			dir:      tmpDir,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsWithinDir(tt.path, tt.dir)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}
