package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/indaco/semver-cli/internal/semver"
)

/* ------------------------------------------------------------------------- */
/* SUCCESS CASES                                                             */
/* ------------------------------------------------------------------------- */

func TestCLI_InitCommand_CreatesFile(t *testing.T) {
	tmp := t.TempDir()
	versionPath := filepath.Join(tmp, ".version")

	output := captureStdout(func() {
		runCLITest(t, []string{"semver", "init"}, tmp)
	})

	data, err := os.ReadFile(versionPath)
	if err != nil {
		t.Fatalf("expected .version file to be created, got error: %v", err)
	}

	got := strings.TrimSpace(string(data))
	if got != "0.1.0" {
		t.Errorf("expected version '0.1.0', got %q", got)
	}

	expectedOutput := fmt.Sprintf("Initialized %s with version 0.1.0", versionPath)
	if strings.TrimSpace(output) != expectedOutput {
		t.Errorf("unexpected output.\nExpected: %q\nGot:      %q", expectedOutput, output)
	}
}

func TestCLI_BumpPatchCommand(t *testing.T) {
	tmp := t.TempDir()
	writeVersionFile(t, tmp, "1.2.3")

	runCLITest(t, []string{"semver", "patch"}, tmp)

	content, _ := os.ReadFile(filepath.Join(tmp, ".version"))
	if got := strings.TrimSpace(string(content)); got != "1.2.4" {
		t.Errorf("expected 1.2.4, got %q", got)
	}
}

func TestCLI_BumpMinorCommand(t *testing.T) {
	tmp := t.TempDir()
	writeVersionFile(t, tmp, "1.2.3-alpha")

	runCLITest(t, []string{"semver", "minor"}, tmp)

	content, _ := os.ReadFile(filepath.Join(tmp, ".version"))
	if got := strings.TrimSpace(string(content)); got != "1.3.0" {
		t.Errorf("expected 1.3.0, got %q", got)
	}
}

func TestCLI_BumpMajorCommand(t *testing.T) {
	tmp := t.TempDir()
	writeVersionFile(t, tmp, "1.2.3")

	runCLITest(t, []string{"semver", "major"}, tmp)

	content, _ := os.ReadFile(filepath.Join(tmp, ".version"))
	if got := strings.TrimSpace(string(content)); got != "2.0.0" {
		t.Errorf("expected 2.0.0, got %q", got)
	}
}

func TestCLI_PreCommand_StaticLabel(t *testing.T) {
	tmp := t.TempDir()
	writeVersionFile(t, tmp, "1.2.3")

	runCLITest(t, []string{"semver", "pre", "--label", "beta.1"}, tmp)

	content, _ := os.ReadFile(filepath.Join(tmp, ".version"))
	if got := strings.TrimSpace(string(content)); got != "1.2.4-beta.1" {
		t.Errorf("expected 1.2.4-beta.1, got %q", got)
	}
}

func TestCLI_PreCommand_Increment(t *testing.T) {
	tmp := t.TempDir()
	writeVersionFile(t, tmp, "1.2.3-beta.3")

	runCLITest(t, []string{"semver", "pre", "--label", "beta", "--inc"}, tmp)

	content, _ := os.ReadFile(filepath.Join(tmp, ".version"))
	if got := strings.TrimSpace(string(content)); got != "1.2.3-beta.4" {
		t.Errorf("expected 1.2.3-beta.4, got %q", got)
	}
}

func TestCLI_ShowCommand(t *testing.T) {
	tmp := t.TempDir()
	writeVersionFile(t, tmp, "9.8.7")

	output := captureStdout(func() {
		runCLITest(t, []string{"semver", "show"}, tmp)
	})

	if output != "9.8.7" {
		t.Errorf("expected output '9.8.7', got %q", output)
	}
}

/* ------------------------------------------------------------------------- */
/* ERROR CASES                                                               */
/* ------------------------------------------------------------------------- */

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

	app := newCLI(versionPath)

	err := app.Run(context.Background(), []string{"semver", "init"})
	if err == nil {
		t.Fatal("expected initialization error, got nil")
	}
	if !strings.Contains(err.Error(), "permission denied") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestCLI_InitCommand_FileAlreadyExists(t *testing.T) {
	tmp := t.TempDir()
	versionPath := filepath.Join(tmp, ".version")
	if err := os.WriteFile(versionPath, []byte("1.2.3\n"), 0600); err != nil {
		t.Fatal(err)
	}

	output := captureStdout(func() {
		runCLITest(t, []string{"semver", "init"}, tmp)
	})

	expected := fmt.Sprintf("Version file already exists at %s", versionPath)
	if !strings.Contains(output, expected) {
		t.Errorf("expected output to contain %q, got %q", expected, output)
	}
}

func TestCLI_ShowCommand_FileNotFound(t *testing.T) {
	tmp := t.TempDir()
	defaultPath := filepath.Join(tmp, ".version")
	app := newCLI(defaultPath)

	err := app.Run(context.Background(), []string{"semver", "show", "--path", "./missing.version"})
	if err == nil {
		t.Fatal("expected error due to missing version file, got nil")
	}

	if !strings.Contains(err.Error(), "no such file") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestCLI_PreCommand_InvalidVersion(t *testing.T) {
	tmp := t.TempDir()
	customPath := filepath.Join(tmp, "bad.version")

	// Write invalid version string before CLI setup
	_ = os.WriteFile(customPath, []byte("not-a-version\n"), semver.VersionFilePerm)

	defaultPath := filepath.Join(tmp, ".version") // not used, but required by newCLI
	app := newCLI(defaultPath)

	err := app.Run(context.Background(), []string{
		"semver", "pre", "--label", "alpha", "--path", customPath,
	})
	if err == nil {
		t.Fatal("expected error due to invalid version, got nil")
	}
	if !strings.Contains(err.Error(), "invalid version format") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestCLI_BumpMinor_InitializeVersionFileError(t *testing.T) {
	tmp := t.TempDir()

	// Create a non-writable directory
	noWrite := filepath.Join(tmp, "protected")
	if err := os.Mkdir(noWrite, 0555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = os.Chmod(noWrite, 0755)
	})

	// Use a path inside the non-writable dir
	protectedPath := filepath.Join(noWrite, ".version")

	defaultPath := filepath.Join(tmp, ".version") // not used but needed for CLI setup
	app := newCLI(defaultPath)

	err := app.Run(context.Background(), []string{
		"semver", "minor", "--path", protectedPath,
	})
	if err == nil {
		t.Fatal("expected error from InitializeVersionFile, got nil")
	}
	if !strings.Contains(err.Error(), "permission denied") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestCLI_BumpMajor_InitializeVersionFileError(t *testing.T) {
	tmp := t.TempDir()

	noWrite := filepath.Join(tmp, "protected")
	if err := os.Mkdir(noWrite, 0555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = os.Chmod(noWrite, 0755)
	})

	protectedPath := filepath.Join(noWrite, ".version")

	defaultPath := filepath.Join(tmp, ".version")
	app := newCLI(defaultPath)

	err := app.Run(context.Background(), []string{
		"semver", "major", "--path", protectedPath,
	})
	if err == nil {
		t.Fatal("expected error from InitializeVersionFile, got nil")
	}
	if !strings.Contains(err.Error(), "permission denied") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestCLI_PreCommand_InitializeVersionFileError(t *testing.T) {
	tmp := t.TempDir()

	noWrite := filepath.Join(tmp, "protected")
	if err := os.Mkdir(noWrite, 0555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = os.Chmod(noWrite, 0755)
	})

	protectedPath := filepath.Join(noWrite, ".version")

	defaultPath := filepath.Join(tmp, ".version")
	app := newCLI(defaultPath)

	err := app.Run(context.Background(), []string{
		"semver", "pre", "--label", "alpha", "--path", protectedPath,
	})
	if err == nil {
		t.Fatal("expected error from InitializeVersionFile, got nil")
	}
	if !strings.Contains(err.Error(), "permission denied") {
		t.Errorf("unexpected error: %v", err)
	}
}

/* ------------------------------------------------------------------------- */
/* HELPERS                                                                   */
/* ------------------------------------------------------------------------- */
func writeVersionFile(t *testing.T, dir, version string) string {
	t.Helper()
	path := filepath.Join(dir, ".version")
	if err := os.WriteFile(path, []byte(version+"\n"), semver.VersionFilePerm); err != nil {
		t.Fatal(err)
	}
	return path
}

func captureStdout(f func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	return strings.TrimSpace(buf.String())
}

func runCLITest(t *testing.T, args []string, workdir string) {
	t.Helper()

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}

	if err := os.Chdir(workdir); err != nil {
		t.Fatalf("failed to change to workdir: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(origDir); err != nil {
			t.Fatalf("failed to restore working directory: %v", err)
		}
	})

	versionPath := filepath.Join(workdir, ".version")

	app := newCLI(versionPath)

	err = app.Run(context.Background(), args)
	if err != nil {
		t.Fatalf("app.Run failed: %v", err)
	}
}
