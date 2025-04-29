package precmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/indaco/semver-cli/internal/clix"
	"github.com/indaco/semver-cli/internal/config"
	"github.com/indaco/semver-cli/internal/semver"
	"github.com/indaco/semver-cli/internal/testutils"
	"github.com/urfave/cli/v3"
)

func TestCLI_PreCommand_StaticLabel(t *testing.T) {
	tmpDir := t.TempDir()
	versionPath := filepath.Join(tmpDir, ".version")

	// Prepare and run the CLI command
	cfg := &config.Config{Path: versionPath}
	appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run()})

	testutils.WriteTempVersionFile(t, tmpDir, "1.2.3")
	testutils.RunCLITest(t, appCli, []string{"semver", "pre", "--label", "beta.1"}, tmpDir)
	content := testutils.ReadTempVersionFile(t, tmpDir)
	if got := content; got != "1.2.4-beta.1" {
		t.Errorf("expected 1.2.4-beta.1, got %q", got)
	}
}

func TestCLI_PreCommand_Increment(t *testing.T) {
	tmpDir := t.TempDir()
	testutils.WriteTempVersionFile(t, tmpDir, "1.2.3-beta.3")
	versionPath := filepath.Join(tmpDir, ".version")

	// Prepare and run the CLI command
	cfg := &config.Config{Path: versionPath}
	appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run()})

	testutils.RunCLITest(t, appCli, []string{"semver", "pre", "--label", "beta", "--inc"}, tmpDir)
	content := testutils.ReadTempVersionFile(t, tmpDir)
	if got := content; got != "1.2.3-beta.4" {
		t.Errorf("expected 1.2.3-beta.4, got %q", got)
	}
}

func TestCLI_PreCommand_AutoInitFeedback(t *testing.T) {
	tmpDir := t.TempDir()
	versionPath := filepath.Join(tmpDir, ".version")

	// Prepare and run the CLI command
	cfg := &config.Config{Path: versionPath}
	appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run()})

	output, err := testutils.CaptureStdout(func() {
		testutils.RunCLITest(t, appCli, []string{"semver", "pre", "--label", "alpha"}, tmpDir)
	})
	if err != nil {
		t.Fatalf("Failed to capture stdout: %v", err)
	}

	expected := fmt.Sprintf("Auto-initialized %s with default version", versionPath)
	if !strings.Contains(output, expected) {
		t.Errorf("expected feedback %q, got %q", expected, output)
	}
}

func TestCLI_PreCommand_InvalidVersion(t *testing.T) {
	tmp := t.TempDir()
	customPath := filepath.Join(tmp, "bad.version")

	// Write invalid version string before CLI setup
	_ = os.WriteFile(customPath, []byte("not-a-version\n"), semver.VersionFilePerm)

	defaultPath := filepath.Join(tmp, ".version")

	// Prepare and run the CLI command
	cfg := &config.Config{Path: defaultPath}
	appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run()})

	err := appCli.Run(context.Background(), []string{
		"semver", "pre", "--label", "alpha", "--path", customPath,
	})
	if err == nil {
		t.Fatal("expected error due to invalid version, got nil")
	}
	if !strings.Contains(err.Error(), "invalid version format") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestCLI_PreCommand_SaveVersionFails(t *testing.T) {
	if os.Getenv("TEST_SEMVER_PRE_SAVE_FAIL") == "1" {
		tmp := t.TempDir()
		versionPath := filepath.Join(tmp, ".version")

		if err := os.WriteFile(versionPath, []byte("1.2.3\n"), 0444); err != nil {
			fmt.Fprintln(os.Stderr, "failed to write .version file:", err)
			os.Exit(1)
		}

		if err := os.Chmod(versionPath, 0444); err != nil {
			fmt.Fprintln(os.Stderr, "failed to chmod .version file:", err)
			os.Exit(1)
		}

		// Prepare and run the CLI command
		cfg := &config.Config{Path: versionPath}
		appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run()})

		err := appCli.Run(context.Background(), []string{
			"semver", "pre", "--label", "rc", "--path", versionPath,
		})

		_ = os.Chmod(versionPath, 0644)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		os.Exit(0) // Unexpected success
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestCLI_PreCommand_SaveVersionFails")
	cmd.Env = append(os.Environ(), "TEST_SEMVER_PRE_SAVE_FAIL=1")
	output, err := cmd.CombinedOutput()

	if err == nil {
		t.Fatal("expected error due to save failure, got nil")
	}

	if !strings.Contains(string(output), "failed to save version") {
		t.Errorf("expected wrapped error message, got: %q", string(output))
	}
}

func TestCLI_PreCommand_FromCommandFails(t *testing.T) {
	tmpDir := t.TempDir()
	versionPath := filepath.Join(tmpDir, ".version")

	// Backup and override clix.FromCommand
	originalFromCommand := clix.FromCommandFn
	clix.FromCommandFn = func(cmd *cli.Command) (bool, error) {
		return false, fmt.Errorf("mock FromCommand error")
	}
	t.Cleanup(func() { clix.FromCommandFn = originalFromCommand })

	cfg := &config.Config{Path: versionPath}
	appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run()})

	err := appCli.Run(context.Background(), []string{"semver", "pre", "--label", "beta.1"})
	if err == nil || !strings.Contains(err.Error(), "mock FromCommand error") {
		t.Fatalf("expected FromCommand error, got: %v", err)
	}
}
