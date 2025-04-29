package clix

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/indaco/semver-cli/internal/config"
	"github.com/indaco/semver-cli/internal/semver"
	"github.com/indaco/semver-cli/internal/testutils"
	"github.com/urfave/cli/v3"
)

func TestGetOrInitVersionFile(t *testing.T) {
	t.Run("noAutoInit=true and file exists", func(t *testing.T) {
		tmpDir := t.TempDir()
		tmpFile := testutils.WriteTempVersionFile(t, tmpDir, "0.1.0")

		created, err := getOrInitVersionFile(tmpFile, true)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if created {
			t.Errorf("expected created=false, got true")
		}
	})

	t.Run("noAutoInit=true and file missing", func(t *testing.T) {
		missingPath := filepath.Join(t.TempDir(), "missing.version")

		created, err := getOrInitVersionFile(missingPath, true)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if created {
			t.Errorf("expected created=false, got true")
		}
	})

	t.Run("noAutoInit=false and initialization succeeds", func(t *testing.T) {
		tmpDir := t.TempDir()
		targetPath := filepath.Join(tmpDir, ".version")

		created, err := getOrInitVersionFile(targetPath, false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !created {
			t.Errorf("expected created=true, got false")
		}

		// Check file actually created
		if _, err := os.Stat(targetPath); err != nil {
			t.Errorf("expected file to exist, got error: %v", err)
		}
	})
}

func TestGetOrInitVersionFile_InitError(t *testing.T) {
	// Backup the real function
	originalFunc := semver.InitializeVersionFileFunc
	defer func() { semver.InitializeVersionFileFunc = originalFunc }()

	// Override with a function that always fails
	semver.InitializeVersionFileFunc = func(path string) error {
		return errors.New("mock init failure")
	}

	tmpDir := t.TempDir()
	targetPath := filepath.Join(tmpDir, ".version")

	created, err := getOrInitVersionFile(targetPath, false)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if created {
		t.Errorf("expected created=false, got true")
	}
	if !strings.Contains(err.Error(), "mock init failure") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestFromCommand(t *testing.T) {
	t.Run("path exists, no-auto-init true", func(t *testing.T) {
		tmpDir := t.TempDir()
		tmpFile := testutils.WriteTempVersionFile(t, tmpDir, "0.1.0")

		cfg := &config.Config{Path: tmpFile}
		appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{})

		created, err := FromCommandFn(appCli)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if created {
			t.Errorf("expected created=false, got true")
		}
	})

	t.Run("path init success, no-auto-init false", func(t *testing.T) {
		tmpDir := t.TempDir()
		targetPath := filepath.Join(tmpDir, ".version")

		cfg := &config.Config{Path: targetPath}
		appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{})

		created, err := FromCommandFn(appCli)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !created {
			t.Errorf("expected created=true, got false")
		}
	})
}
