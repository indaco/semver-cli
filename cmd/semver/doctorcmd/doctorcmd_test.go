package doctorcmd

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/indaco/semver-cli/internal/config"
	"github.com/indaco/semver-cli/internal/testutils"
	"github.com/urfave/cli/v3"
)

func TestCLI_ValidateCommand_ValidCases(t *testing.T) {
	tmpDir := t.TempDir()
	versionPath := filepath.Join(tmpDir, ".version")

	// Prepare and run the CLI command
	cfg := &config.Config{Path: versionPath}
	appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run()})

	tests := []struct {
		name           string
		version        string
		expectedOutput string
	}{
		{
			name:    "valid semantic version",
			version: "1.2.3",
		},
		{
			name:    "valid version with build metadata",
			version: "1.2.3+exp.sha.5114f85",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testutils.WriteTempVersionFile(t, tmpDir, tt.version)

			output, err := testutils.CaptureStdout(func() {
				testutils.RunCLITest(t, appCli, []string{"semver", "doctor"}, tmpDir)
			})
			if err != nil {
				t.Fatalf("Failed to capture stdout: %v", err)
			}

			expected := fmt.Sprintf("Valid version file at %s/.version", tmpDir)
			if !strings.Contains(output, expected) {
				t.Errorf("expected output to contain %q, got %q", expected, output)
			}
		})
	}
}

func TestCLI_ValidateCommand_Errors(t *testing.T) {
	tests := []struct {
		name          string
		version       string
		expectedError string
	}{
		{"invalid version string", "not-a-version", "invalid version"},
		{"invalid build metadata", "1.0.0+inv@lid-meta", "invalid version"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			testutils.WriteTempVersionFile(t, tmpDir, tt.version)
			versionPath := filepath.Join(tmpDir, ".version")

			// Prepare and run the CLI command
			cfg := &config.Config{Path: versionPath}
			appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run()})

			err := appCli.Run(context.Background(), []string{"semver", "doctor"})
			if err == nil || !strings.Contains(err.Error(), tt.expectedError) {
				t.Fatalf("expected error containing %q, got: %v", tt.expectedError, err)
			}
		})
	}
}
