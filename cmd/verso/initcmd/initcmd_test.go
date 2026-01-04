package initcmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/indaco/verso/cmd/verso/bumpcmd"
	"github.com/indaco/verso/cmd/verso/precmd"
	"github.com/indaco/verso/internal/config"
	"github.com/indaco/verso/internal/testutils"
	"github.com/urfave/cli/v3"
)

func TestCLI_InitCommand_CreatesFile(t *testing.T) {
	tmp := t.TempDir()
	versionPath := filepath.Join(tmp, ".version")

	// Prepare and run the CLI command
	cfg := &config.Config{Path: versionPath}
	appCli := testutils.BuildCLIForTests(
		cfg.Path,
		[]*cli.Command{Run(), bumpcmd.Run(cfg)},
	)

	output, err := testutils.CaptureStdout(func() {
		testutils.RunCLITest(t, appCli, []string{"verso", "init"}, tmp)
	})
	if err != nil {
		t.Fatalf("Failed to capture stdout: %v", err)
	}

	got := testutils.ReadTempVersionFile(t, tmp)
	if got != "0.1.0" {
		t.Errorf("expected version '0.1.0', got %q", got)
	}

	expectedOutput := fmt.Sprintf("Initialized %s with version 0.1.0", versionPath)
	if strings.TrimSpace(output) != expectedOutput {
		t.Errorf("unexpected output.\nExpected: %q\nGot:      %q", expectedOutput, output)
	}
}

func TestCLI_InitCommand_InitializationError(t *testing.T) {
	tmp := t.TempDir()
	noWrite := filepath.Join(tmp, "nowrite")
	if err := os.Mkdir(noWrite, 0555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = os.Chmod(noWrite, 0755)
	})

	versionPath := filepath.Join(noWrite, ".version")

	// Prepare and run the CLI command
	cfg := &config.Config{Path: versionPath}
	appCli := testutils.BuildCLIForTests(
		cfg.Path,
		[]*cli.Command{Run(), bumpcmd.Run(cfg)},
	)

	err := appCli.Run(context.Background(), []string{"verso", "init"})
	if err == nil {
		t.Fatal("expected initialization error, got nil")
	}
	if !strings.Contains(err.Error(), "permission denied") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestCLI_InitCommand_FileAlreadyExists(t *testing.T) {
	tmpDir := t.TempDir()
	testutils.WriteTempVersionFile(t, tmpDir, "1.2.3")
	versionPath := filepath.Join(tmpDir, ".version")

	// Prepare and run the CLI command
	cfg := &config.Config{Path: versionPath}
	appCli := testutils.BuildCLIForTests(
		cfg.Path,
		[]*cli.Command{Run(), bumpcmd.Run(cfg)},
	)

	output, err := testutils.CaptureStdout(func() {
		testutils.RunCLITest(t, appCli, []string{"verso", "init"}, tmpDir)
	})
	if err != nil {
		t.Fatalf("Failed to capture stdout: %v", err)
	}

	expected := fmt.Sprintf("Version file already exists at %s", versionPath)
	if !strings.Contains(output, expected) {
		t.Errorf("expected output to contain %q, got %q", expected, output)
	}
}

func TestCLI_InitCommand_ExistingInvalidContent(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, ".version")

	// Create a file with invalid version content (simulating manual corruption)
	if err := os.WriteFile(path, []byte("not-a-version\n"), 0600); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	// Prepare and run the CLI command
	cfg := &config.Config{Path: path}
	appCli := testutils.BuildCLIForTests(
		cfg.Path,
		[]*cli.Command{Run(), bumpcmd.Run(cfg)},
	)

	err := appCli.Run(context.Background(), []string{
		"verso", "init", "--path", path,
	})
	if err == nil {
		t.Fatal("expected error due to invalid version content, got nil")
	}
	if !strings.Contains(err.Error(), "failed to read version file") {
		t.Errorf("unexpected error: %v", err)
	}
	if !strings.Contains(err.Error(), "invalid version format") {
		t.Errorf("expected invalid version format message, got %v", err)
	}
}

func TestCLI_Command_InitializeVersionFilePermissionErrors(t *testing.T) {
	tests := []struct {
		name    string
		command []string
	}{
		{"bump minor", []string{"verso", "bump", "minor"}},
		{"bump major", []string{"verso", "bump", "major"}},
		{"pre label", []string{"verso", "pre", "--label", "alpha"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmp := t.TempDir()
			noWrite := filepath.Join(tmp, "protected")
			if err := os.Mkdir(noWrite, 0555); err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				_ = os.Chmod(noWrite, 0755)
			})
			protectedPath := filepath.Join(noWrite, ".version")

			// Prepare and run the CLI command
			cfg := &config.Config{Path: protectedPath}
			appCli := testutils.BuildCLIForTests(
				cfg.Path,
				[]*cli.Command{Run(), bumpcmd.Run(cfg), precmd.Run()},
			)

			err := appCli.Run(context.Background(), append(tt.command, "--path", protectedPath))
			if err == nil || !strings.Contains(err.Error(), "permission denied") {
				t.Fatalf("expected permission denied error, got: %v", err)
			}
		})
	}
}
