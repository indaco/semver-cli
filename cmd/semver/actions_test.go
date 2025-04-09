package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
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

	runCLITest(t, []string{"semver", "bump", "patch"}, tmp)

	content, _ := os.ReadFile(filepath.Join(tmp, ".version"))
	if got := strings.TrimSpace(string(content)); got != "1.2.4" {
		t.Errorf("expected 1.2.4, got %q", got)
	}
}

func TestCLI_BumpPatchCommand_AutoInitFeedback(t *testing.T) {
	tmp := t.TempDir()

	output := captureStdout(func() {
		runCLITest(t, []string{"semver", "bump", "patch"}, tmp)
	})

	expected := fmt.Sprintf("Auto-initialized %s with default version", filepath.Join(tmp, ".version"))
	if !strings.Contains(output, expected) {
		t.Errorf("expected feedback %q, got %q", expected, output)
	}
}

func TestCLI_BumpMinorCommand(t *testing.T) {
	tmp := t.TempDir()
	writeVersionFile(t, tmp, "1.2.3-alpha")

	runCLITest(t, []string{"semver", "bump", "minor"}, tmp)

	content, _ := os.ReadFile(filepath.Join(tmp, ".version"))
	if got := strings.TrimSpace(string(content)); got != "1.3.0" {
		t.Errorf("expected 1.3.0, got %q", got)
	}
}

func TestCLI_BumpMinorCommand_AutoInitFeedback(t *testing.T) {
	tmp := t.TempDir()

	output := captureStdout(func() {
		runCLITest(t, []string{"semver", "bump", "minor"}, tmp)
	})

	expected := fmt.Sprintf("Auto-initialized %s with default version", filepath.Join(tmp, ".version"))
	if !strings.Contains(output, expected) {
		t.Errorf("expected feedback %q, got %q", expected, output)
	}
}

func TestCLI_BumpMajorCommand(t *testing.T) {
	tmp := t.TempDir()
	writeVersionFile(t, tmp, "1.2.3")

	runCLITest(t, []string{"semver", "bump", "major"}, tmp)

	content, _ := os.ReadFile(filepath.Join(tmp, ".version"))
	if got := strings.TrimSpace(string(content)); got != "2.0.0" {
		t.Errorf("expected 2.0.0, got %q", got)
	}
}

func TestCLI_BumpMajorCommand_AutoInitFeedback(t *testing.T) {
	tmp := t.TempDir()

	output := captureStdout(func() {
		runCLITest(t, []string{"semver", "bump", "major"}, tmp)
	})

	expected := fmt.Sprintf("Auto-initialized %s with default version", filepath.Join(tmp, ".version"))
	if !strings.Contains(output, expected) {
		t.Errorf("expected feedback %q, got %q", expected, output)
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

func TestCLI_PreCommand_AutoInitFeedback(t *testing.T) {
	tmp := t.TempDir()

	output := captureStdout(func() {
		runCLITest(t, []string{"semver", "pre", "--label", "alpha"}, tmp)
	})

	expected := fmt.Sprintf("Auto-initialized %s with default version", filepath.Join(tmp, ".version"))
	if !strings.Contains(output, expected) {
		t.Errorf("expected feedback %q, got %q", expected, output)
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

func TestCLI_SetVersion_Valid(t *testing.T) {
	tmp := t.TempDir()

	runCLITest(t, []string{"semver", "set", "2.5.0"}, tmp)

	content, _ := os.ReadFile(filepath.Join(tmp, ".version"))
	if got := strings.TrimSpace(string(content)); got != "2.5.0" {
		t.Errorf("expected 2.5.0, got %q", got)
	}
}

func TestCLI_SetVersion_WithPreRelease(t *testing.T) {
	tmp := t.TempDir()

	runCLITest(t, []string{"semver", "set", "3.0.0", "--pre", "beta.2"}, tmp)

	content, _ := os.ReadFile(filepath.Join(tmp, ".version"))
	if got := strings.TrimSpace(string(content)); got != "3.0.0-beta.2" {
		t.Errorf("expected 3.0.0-beta.2, got %q", got)
	}
}

func TestCLI_SetVersion_WithBuildMetadata(t *testing.T) {
	tmp := t.TempDir()

	runCLITest(t, []string{"semver", "set", "1.0.0", "--meta", "001"}, tmp)

	content, _ := os.ReadFile(filepath.Join(tmp, ".version"))
	if got := strings.TrimSpace(string(content)); got != "1.0.0+001" {
		t.Errorf("expected 1.0.0+001, got %q", got)
	}
}

func TestCLI_SetVersion_WithPreReleaseAndBuildMetadata(t *testing.T) {
	tmp := t.TempDir()

	runCLITest(t, []string{"semver", "set", "1.0.0", "--pre", "alpha.1", "--meta", "exp.sha.5114f85"}, tmp)

	content, _ := os.ReadFile(filepath.Join(tmp, ".version"))
	if got := strings.TrimSpace(string(content)); got != "1.0.0-alpha.1+exp.sha.5114f85" {
		t.Errorf("expected 1.0.0-alpha.1+exp.sha.5114f85, got %q", got)
	}
}

func TestCLI_ValidateCommand_ValidVersion(t *testing.T) {
	tmp := t.TempDir()
	writeVersionFile(t, tmp, "1.2.3")

	output := captureStdout(func() {
		runCLITest(t, []string{"semver", "validate"}, tmp)
	})

	expected := fmt.Sprintf("Valid version file at %s/.version", tmp)
	if !strings.Contains(output, expected) {
		t.Errorf("expected output to contain %q, got %q", expected, output)
	}
}

func TestCLI_ValidateCommand_WithMetadata(t *testing.T) {
	tmp := t.TempDir()
	writeVersionFile(t, tmp, "1.2.3+exp.sha.5114f85")

	output := captureStdout(func() {
		runCLITest(t, []string{"semver", "validate"}, tmp)
	})

	expected := fmt.Sprintf("Valid version file at %s/.version", tmp)
	if !strings.Contains(output, expected) {
		t.Errorf("expected output to contain %q, got %q", expected, output)
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

func TestCLI_InitCommand_ReadVersionFails(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, ".version")

	// Override InitializeVersionFile to write invalid content
	original := semver.InitializeVersionFile
	semver.InitializeVersionFile = func(p string) error {
		return os.WriteFile(p, []byte("not-a-version\n"), 0600)
	}
	t.Cleanup(func() { semver.InitializeVersionFile = original })

	app := newCLI(path)

	err := app.Run(context.Background(), []string{
		"semver", "init", "--path", path,
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
		"semver", "bump", "minor", "--path", protectedPath,
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
		"semver", "bump", "major", "--path", protectedPath,
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

func TestCLI_PreCommand_SaveVersionFails(t *testing.T) {
	if os.Getenv("TEST_SEMVER_PRE_SAVE_FAIL") == "1" {
		tmp := t.TempDir()
		versionPath := filepath.Join(tmp, ".version")

		// Write a valid version
		if err := os.WriteFile(versionPath, []byte("1.2.3\n"), 0444); err != nil {
			fmt.Fprintln(os.Stderr, "failed to write .version file:", err)
			os.Exit(1)
		}

		// Ensure the file itself is read-only
		if err := os.Chmod(versionPath, 0444); err != nil {
			fmt.Fprintln(os.Stderr, "failed to chmod .version file:", err)
			os.Exit(1)
		}
		defer func() {
			_ = os.Chmod(versionPath, 0644) // cleanup
		}()

		app := newCLI(versionPath)
		err := app.Run(context.Background(), []string{
			"semver", "pre", "--label", "rc", "--path", versionPath,
		})
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

func TestCLI_SetVersion_InvalidFormat(t *testing.T) {
	tmp := t.TempDir()
	app := newCLI(filepath.Join(tmp, ".version"))

	err := app.Run(context.Background(), []string{"semver", "set", "invalid.version"})
	if err == nil {
		t.Fatal("expected error due to invalid version format, got nil")
	}
	if !strings.Contains(err.Error(), "invalid version format") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestCLI_SetVersion_MissingArgument(t *testing.T) {
	if os.Getenv("TEST_SEMVER_MISSING_ARG") == "1" {
		tmp := t.TempDir()
		versionPath := filepath.Join(tmp, ".version")
		app := newCLI(versionPath)
		err := app.Run(context.Background(), []string{"semver", "set"})
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestCLI_SetVersion_MissingArgument")
	cmd.Env = append(os.Environ(), "TEST_SEMVER_MISSING_ARG=1")
	output, err := cmd.CombinedOutput()

	if err == nil {
		t.Fatal("expected non-zero exit status")
	}

	expected := "missing required version argument"
	if !strings.Contains(string(output), expected) {
		t.Errorf("expected output to contain %q, got %q", expected, string(output))
	}
}

func TestCLI_SetVersion_SaveError(t *testing.T) {
	tmp := t.TempDir()

	protectedDir := filepath.Join(tmp, "protected")
	if err := os.Mkdir(protectedDir, 0555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = os.Chmod(protectedDir, 0755)
	})

	versionPath := filepath.Join(protectedDir, ".version")
	app := newCLI(versionPath)

	err := app.Run(context.Background(), []string{
		"semver", "set", "3.0.0", "--path", versionPath,
	})
	if err == nil {
		t.Fatal("expected error due to save failure, got nil")
	}

	if !strings.Contains(err.Error(), "failed to save version") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestCLI_ValidateCommand_InvalidVersion(t *testing.T) {
	tmp := t.TempDir()
	path := writeVersionFile(t, tmp, "not-a-version")

	app := newCLI(path)
	err := app.Run(context.Background(), []string{"semver", "validate"})
	if err == nil {
		t.Fatal("expected error due to invalid version, got nil")
	}
	if !strings.Contains(err.Error(), "invalid version") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestCLI_ValidateCommand_MissingFile(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "missing.version")

	app := newCLI(path)
	err := app.Run(context.Background(), []string{"semver", "validate", "--path", path})
	if err == nil {
		t.Fatal("expected error due to missing version file, got nil")
	}
	if !strings.Contains(err.Error(), "no such file") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestCLI_ValidateCommand_InvalidMetadata(t *testing.T) {
	tmp := t.TempDir()
	writeVersionFile(t, tmp, "1.0.0+inv@lid-meta")

	app := newCLI(filepath.Join(tmp, ".version"))
	err := app.Run(context.Background(), []string{"semver", "validate"})

	if err == nil {
		t.Fatal("expected error due to invalid metadata, got nil")
	}
	if !strings.Contains(err.Error(), "invalid version") {
		t.Errorf("expected error about invalid version, got %v", err)
	}
}

func TestCLI_ShowCommand_NoAutoInit_MissingFile(t *testing.T) {
	if os.Getenv("TEST_SEMVER_NO_AUTO_INIT") == "1" {
		tmp := t.TempDir()
		versionPath := filepath.Join(tmp, ".version")

		app := newCLI(versionPath)
		err := app.Run(context.Background(), []string{"semver", "show", "--no-auto-init"})
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestCLI_ShowCommand_NoAutoInit_MissingFile")
	cmd.Env = append(os.Environ(), "TEST_SEMVER_NO_AUTO_INIT=1")
	output, err := cmd.CombinedOutput()

	if err == nil {
		t.Fatal("expected non-zero exit status")
	}

	expected := "version file not found"
	if !strings.Contains(string(output), expected) {
		t.Errorf("expected output to contain %q, got %q", expected, string(output))
	}
}
func TestCLI_ShowCommand_NoAutoInit_FileExists(t *testing.T) {
	tmp := t.TempDir()
	writeVersionFile(t, tmp, "1.2.3")

	output := captureStdout(func() {
		runCLITest(t, []string{"semver", "show", "--no-auto-init"}, tmp)
	})

	if output != "1.2.3" {
		t.Errorf("expected output '1.2.3', got %q", output)
	}
}

func TestCLI_ShowCommand_InvalidVersionContent(t *testing.T) {
	tmp := t.TempDir()
	versionPath := filepath.Join(tmp, ".version")

	// Write an invalid version string
	if err := os.WriteFile(versionPath, []byte("not-a-semver\n"), 0644); err != nil {
		t.Fatalf("failed to write invalid version: %v", err)
	}

	app := newCLI(versionPath)

	err := app.Run(context.Background(), []string{"semver", "show"})
	if err == nil {
		t.Fatal("expected error due to invalid version, got nil")
	}

	if !strings.Contains(err.Error(), "failed to read version file at") {
		t.Errorf("unexpected error message: %v", err)
	}
	if !strings.Contains(err.Error(), "invalid version format") {
		t.Errorf("error does not mention 'invalid version format': %v", err)
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
