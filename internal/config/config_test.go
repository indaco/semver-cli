package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/indaco/semver-cli/internal/testutils"
)

/* ------------------------------------------------------------------------- */
/* SUCCESS CASES                                                             */
/* ------------------------------------------------------------------------- */
func TestLoadConfig_FromEnv(t *testing.T) {
	os.Setenv("SEMVER_PATH", "env-defined/.version")
	defer os.Unsetenv("SEMVER_PATH")

	cfg, err := LoadConfigFn()
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if cfg.Path != "env-defined/.version" {
		t.Errorf("expected 'env-defined/.version', got %q", cfg.Path)
	}
}

func TestLoadConfig_ValidFile(t *testing.T) {
	content := "path: ./my-folder/.version\n"
	tmpPath := testutils.WriteTempConfig(t, content)
	tmpDir := filepath.Dir(tmpPath)

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory to %s: %v", tmpDir, err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(origDir); err != nil {
			t.Fatalf("failed to restore working directory: %v", err)
		}
	})

	cfg, err := LoadConfigFn()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cfg.Path != "./my-folder/.version" {
		t.Errorf("expected './my-folder/.version', got %q", cfg.Path)
	}
}

func TestLoadConfig_MissingFile(t *testing.T) {
	cfg, err := LoadConfigFn()
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if cfg != nil {
		t.Errorf("expected nil config, got %+v", cfg)
	}
}

func TestLoadConfig_DefaultPathFallback(t *testing.T) {
	content := "{}\n"
	tmpPath := testutils.WriteTempConfig(t, content)
	tmpDir := filepath.Dir(tmpPath)

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory to %s: %v", tmpDir, err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(origDir); err != nil {
			t.Fatalf("failed to restore working directory: %v", err)
		}
	})

	cfg, err := LoadConfigFn()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if cfg == nil {
		t.Fatal("Expected config, got nil")
	}

	if cfg.Path != ".version" {
		t.Errorf("Expected fallback path '.version', got '%s'", cfg.Path)
	}
}

func TestNormalizeVersionPath(t *testing.T) {
	// Case 1: path is a file
	got := NormalizeVersionPath("foo/.version")
	if got != "foo/.version" {
		t.Errorf("expected unchanged path, got %q", got)
	}

	// Case 2: path is a directory
	tmp := t.TempDir()
	got = NormalizeVersionPath(tmp)
	expected := filepath.Join(tmp, ".version")
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

/* ------------------------------------------------------------------------- */
/* ERROR CASES                                                               */
/* ------------------------------------------------------------------------- */
func TestLoadConfig_InvalidYAML(t *testing.T) {
	content := "not_yaml::: true" // this won't populate `path`
	tmpPath := testutils.WriteTempConfig(t, content)
	tmpDir := filepath.Dir(tmpPath)

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(origDir); err != nil {
			t.Fatalf("failed to restore working directory: %v", err)
		}
	})

	cfg, err := LoadConfigFn()
	if err == nil {
		t.Fatal("expected YAML parse error or missing key error, got nil")
	}
	if cfg != nil {
		t.Errorf("expected nil config, got %+v", cfg)
	}
}

func TestLoadConfig_UnmarshalError(t *testing.T) {
	content := ": this is invalid" // invalid YAML syntax
	tmpPath := testutils.WriteTempConfig(t, content)

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	if err := os.Chdir(filepath.Dir(tmpPath)); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(origDir); err != nil {
			t.Fatalf("failed to restore working directory: %v", err)
		}
	})

	cfg, err := LoadConfigFn()
	if err == nil {
		t.Fatal("expected unmarshal error, got nil")
	}
	if cfg != nil {
		t.Errorf("expected nil config, got %+v", cfg)
	}
}

func TestLoadConfig_ReadFileError(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(origDir); err != nil {
			t.Fatalf("failed to restore working directory: %v", err)
		}
	})

	// Create a directory instead of a file
	err = os.Mkdir(".semver.yaml", 0755)
	if err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadConfigFn()
	if err == nil {
		t.Fatal("expected error when reading non-file .semver.yaml, got nil")
	}
	if cfg != nil {
		t.Errorf("expected nil config, got %+v", cfg)
	}
}
