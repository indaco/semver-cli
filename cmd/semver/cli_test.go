package main

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/indaco/semver-cli/internal/config"
	"github.com/indaco/semver-cli/internal/semver"
)

/* ------------------------------------------------------------------------- */
/* SUCCESS CASES                                                             */
/* ------------------------------------------------------------------------- */
func TestNewCLI_BasicStructure(t *testing.T) {
	os.Unsetenv("SEMVER_PATH")

	tmpDir := t.TempDir()
	versionPath := filepath.Join(tmpDir, ".version")
	_ = os.WriteFile(versionPath, []byte("1.2.3\n"), semver.VersionFilePerm)

	cfg := &config.Config{Path: versionPath}
	app := newCLI(cfg)

	wantCommands := []string{"show", "set", "bump", "pre", "validate", "init"}
	for _, name := range wantCommands {
		found := false
		for _, cmd := range app.Commands {
			if cmd.Name == name {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected command %q to be registered", name)
		}
	}
}

func TestNewCLI_ShowCommand(t *testing.T) {
	tmpDir := t.TempDir()
	versionPath := filepath.Join(tmpDir, ".version")
	err := os.WriteFile(versionPath, []byte("2.3.4\n"), semver.VersionFilePerm)
	if err != nil {
		t.Fatal(err)
	}

	// Redirect stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cfg := &config.Config{Path: versionPath}
	app := newCLI(cfg)

	err = app.Run(context.Background(), []string{"semver", "show", "--path", versionPath})

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := strings.TrimSpace(buf.String())

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if output != "2.3.4" {
		t.Errorf("expected output '2.3.4', got %q", output)
	}
}

/* ------------------------------------------------------------------------- */
/* ERROR CASES                                                               */
/* ------------------------------------------------------------------------- */
func TestNewCLI_UsesConfigPath(t *testing.T) {
	tmp := t.TempDir()

	versionPath := filepath.Join(tmp, "custom.version")
	yamlPath := filepath.Join(tmp, ".semver.yaml")
	_ = os.WriteFile(yamlPath, []byte("path: ./custom.version\n"), 0644)
	_ = os.WriteFile(versionPath, []byte("1.0.0\n"), semver.VersionFilePerm)

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}

	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("failed to change to temp directory: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(origDir); err != nil {
			t.Fatalf("failed to restore original working directory: %v", err)
		}
	})

	cfg := &config.Config{Path: versionPath}
	app := newCLI(cfg)

	err = app.Run(context.Background(), []string{"semver", "bump", "patch"})
	if err != nil {
		t.Fatalf("app.Run failed: %v", err)
	}

	content, _ := os.ReadFile(versionPath)
	got := strings.TrimSpace(string(content))
	if got != "1.0.1" {
		t.Errorf("expected version to be 1.0.1, got %q", got)
	}
}

func TestRunCLI_InitializeVersionFileError(t *testing.T) {
	tmp := t.TempDir()

	noWrite := filepath.Join(tmp, "nonwritable")
	if err := os.Mkdir(noWrite, 0555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = os.Chmod(noWrite, 0755)
	})

	versionPath := filepath.Join("nonwritable", ".version")
	yamlPath := filepath.Join(tmp, ".semver.yaml")
	if err := os.WriteFile(yamlPath, []byte("path: "+versionPath+"\n"), 0644); err != nil {
		t.Fatal(err)
	}

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(origDir)
	})

	err = runCLI([]string{"semver", "bump", "patch"})
	if err == nil {
		t.Fatal("expected error from InitializeVersionFile, got nil")
	}
	if !strings.Contains(err.Error(), "permission denied") {
		t.Errorf("unexpected error: %v", err)
	}
}
