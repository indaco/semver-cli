package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/indaco/semver-cli/internal/config"
	"github.com/indaco/semver-cli/internal/semver"
	"github.com/indaco/semver-cli/internal/testutils"
)

/* ------------------------------------------------------------------------- */
/* INIT COMMAND                                                              */
/* ------------------------------------------------------------------------- */

func TestCLI_InitCommand_CreatesFile(t *testing.T) {
	tmp := t.TempDir()
	versionPath := filepath.Join(tmp, ".version")
	appCli := newCLI(versionPath)

	output, err := testutils.CaptureStdout(func() {
		testutils.RunCLITest(t, appCli, []string{"semver", "init"}, tmp)
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

	appCli := newCLI(versionPath)

	err := appCli.Run(context.Background(), []string{"semver", "init"})
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
	appCli := newCLI(versionPath)

	output, err := testutils.CaptureStdout(func() {
		testutils.RunCLITest(t, appCli, []string{"semver", "init"}, tmpDir)
	})
	if err != nil {
		t.Fatalf("Failed to capture stdout: %v", err)
	}

	expected := fmt.Sprintf("Version file already exists at %s", versionPath)
	if !strings.Contains(output, expected) {
		t.Errorf("expected output to contain %q, got %q", expected, output)
	}
}

func TestCLI_InitCommand_ReadVersionFails(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, ".version")

	// Override InitializeVersionFile to write invalid content
	original := semver.InitializeVersionFileFunc
	semver.InitializeVersionFileFunc = func(p string) error {
		return os.WriteFile(p, []byte("not-a-version\n"), 0600)
	}
	t.Cleanup(func() { semver.InitializeVersionFileFunc = original })

	appCli := newCLI(path)

	err := appCli.Run(context.Background(), []string{
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

func TestCLI_Command_InitializeVersionFilePermissionErrors(t *testing.T) {
	tests := []struct {
		name    string
		command []string
	}{
		{"bump minor", []string{"semver", "bump", "minor"}},
		{"bump major", []string{"semver", "bump", "major"}},
		{"pre label", []string{"semver", "pre", "--label", "alpha"}},
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
			appCli := newCLI(protectedPath)

			err := appCli.Run(context.Background(), append(tt.command, "--path", protectedPath))
			if err == nil || !strings.Contains(err.Error(), "permission denied") {
				t.Fatalf("expected permission denied error, got: %v", err)
			}
		})
	}
}

/* ------------------------------------------------------------------------- */
/* BUMP COMMAND                                                              */
/* ------------------------------------------------------------------------- */

func TestCLI_BumpCommand_Variants(t *testing.T) {
	tmpDir := t.TempDir()
	versionPath := filepath.Join(tmpDir, ".version")
	appCli := newCLI(versionPath)

	tests := []struct {
		name     string
		initial  string
		args     []string
		expected string
	}{
		{"patch bump", "1.2.3", []string{"semver", "bump", "patch"}, "1.2.4"},
		{"minor bump", "1.2.3", []string{"semver", "bump", "minor"}, "1.3.0"},
		{"major bump", "1.2.3", []string{"semver", "bump", "major"}, "2.0.0"},
		{"patch bump after pre-release", "1.2.3-alpha", []string{"semver", "bump", "patch"}, "1.2.4"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testutils.WriteTempVersionFile(t, tmpDir, tt.initial)
			testutils.RunCLITest(t, appCli, tt.args, tmpDir)

			got := testutils.ReadTempVersionFile(t, tmpDir)
			if got != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, got)
			}
		})
	}
}

func TestCLI_BumpCommand_AutoInitFeedback(t *testing.T) {
	tests := []struct {
		name    string
		initial string
		args    []string
	}{
		{"patch bump", "1.2.3", []string{"semver", "bump", "patch"}},
		{"minor bump", "1.2.3", []string{"semver", "bump", "minor"}},
		{"major bump", "1.2.3", []string{"semver", "bump", "major"}},
		{"patch bump after pre-release", "1.2.3-alpha", []string{"semver", "bump", "patch"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			versionPath := filepath.Join(tmpDir, ".version")
			appCli := newCLI(versionPath)

			output, err := testutils.CaptureStdout(func() {
				testutils.RunCLITest(t, appCli, tt.args, tmpDir)
			})
			if err != nil {
				t.Fatalf("Failed to capture stdout: %v", err)
			}

			expected := fmt.Sprintf("Auto-initialized %s with default version", versionPath)
			if !strings.Contains(output, expected) {
				t.Errorf("expected feedback %q, got %q", expected, output)
			}
		})
	}
}

func TestCLI_BumpReleaseCmd(t *testing.T) {
	tests := []struct {
		name           string
		initialVersion string
		args           []string
		expected       string
	}{
		{
			name:           "removes pre-release and metadata",
			initialVersion: "1.3.0-alpha.1+ci.123",
			args:           []string{"semver", "bump", "release"},
			expected:       "1.3.0",
		},
		{
			name:           "preserves metadata if flag is set",
			initialVersion: "1.3.0-alpha.2+build.99",
			args:           []string{"semver", "bump", "release", "--preserve-meta"},
			expected:       "1.3.0+build.99",
		},
		{
			name:           "no-op when already final version",
			initialVersion: "2.0.0",
			args:           []string{"semver", "bump", "release"},
			expected:       "2.0.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			versionPath := filepath.Join(tmpDir, ".version")
			appCli := newCLI(versionPath)

			testutils.WriteTempVersionFile(t, tmpDir, tt.initialVersion)
			testutils.RunCLITest(t, appCli, tt.args, tmpDir)

			got := testutils.ReadTempVersionFile(t, tmpDir)
			if got != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, got)
			}
		})
	}
}

func TestCLI_BumpNextCmd(t *testing.T) {
	tests := []struct {
		name     string
		initial  string
		args     []string
		expected string
	}{
		{
			name:     "promotes alpha to release",
			initial:  "1.2.3-alpha.1",
			args:     []string{"semver", "bump", "next"},
			expected: "1.2.3",
		},
		{
			name:     "promotes rc to release",
			initial:  "1.2.3-rc.1",
			args:     []string{"semver", "bump", "next"},
			expected: "1.2.3",
		},
		{
			name:     "default patch bump",
			initial:  "1.2.3",
			args:     []string{"semver", "bump", "next"},
			expected: "1.2.4",
		},
		{
			name:     "promotes pre-release in 0.x series",
			initial:  "0.9.0-alpha.1",
			args:     []string{"semver", "bump", "next"},
			expected: "0.9.0",
		},
		{
			name:     "bump minor from 0.9.0 as a special case",
			initial:  "0.9.0",
			args:     []string{"semver", "bump", "next"},
			expected: "0.10.0",
		},
		{
			name:     "preserve build metadata",
			initial:  "1.2.3-alpha.1+meta.123",
			args:     []string{"semver", "bump", "next", "--preserve-meta"},
			expected: "1.2.3+meta.123",
		},
		{
			name:     "strip build metadata by default",
			initial:  "1.2.3-alpha.1+meta.123",
			args:     []string{"semver", "bump", "next"},
			expected: "1.2.3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			versionPath := filepath.Join(tmpDir, ".version")
			appCli := newCLI(versionPath)

			testutils.WriteTempVersionFile(t, tmpDir, tt.initial)
			testutils.RunCLITest(t, appCli, tt.args, tmpDir)

			got := testutils.ReadTempVersionFile(t, tmpDir)
			if got != tt.expected {
				t.Errorf("expected version %q, got %q", tt.expected, got)
			}
		})
	}
}

func TestCLI_BumpNextCommand_WithLabelAndMeta(t *testing.T) {
	tests := []struct {
		name    string
		initial string
		args    []string
		want    string
	}{
		{
			name:    "label=patch",
			initial: "1.2.3",
			args:    []string{"semver", "bump", "next", "--label", "patch"},
			want:    "1.2.4",
		},
		{
			name:    "label=minor",
			initial: "1.2.3",
			args:    []string{"semver", "bump", "next", "--label", "minor"},
			want:    "1.3.0",
		},
		{
			name:    "label=major",
			initial: "1.2.3",
			args:    []string{"semver", "bump", "next", "--label", "major"},
			want:    "2.0.0",
		},
		{
			name:    "label=minor with metadata",
			initial: "1.2.3",
			args:    []string{"semver", "bump", "next", "--label", "minor", "--meta", "build.42"},
			want:    "1.3.0+build.42",
		},
		{
			name:    "preserve existing metadata",
			initial: "1.2.3+ci.88",
			args:    []string{"semver", "bump", "next", "--label", "patch", "--preserve-meta"},
			want:    "1.2.4+ci.88",
		},
		{
			name:    "override existing metadata",
			initial: "1.2.3+ci.88",
			args:    []string{"semver", "bump", "next", "--label", "patch", "--meta", "ci.99"},
			want:    "1.2.4+ci.99",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			versionPath := filepath.Join(tmpDir, ".version")
			appCli := newCLI(versionPath)
			testutils.WriteTempVersionFile(t, tmpDir, tt.initial)
			testutils.RunCLITest(t, appCli, tt.args, tmpDir)

			got := testutils.ReadTempVersionFile(t, tmpDir)
			if got != tt.want {
				t.Errorf("expected %q, got %q", tt.want, got)
			}
		})
	}
}

func TestBumpReleaseCmd_ErrorOnReadVersion(t *testing.T) {
	tmp := t.TempDir()
	versionPath := testutils.WriteTempVersionFile(t, tmp, "invalid-version")

	appCli := newCLI(versionPath)
	err := appCli.Run(context.Background(), []string{
		"semver", "bump", "release", "--path", versionPath,
	})

	if err == nil || !strings.Contains(err.Error(), "failed to read version") {
		t.Errorf("expected read version error, got: %v", err)
	}
}

func TestCLI_BumpReleaseCommand_SaveVersionFails(t *testing.T) {
	tmp := t.TempDir()
	versionPath := filepath.Join(tmp, ".version")

	// Write valid pre-release content
	if err := os.WriteFile(versionPath, []byte("1.2.3-alpha\n"), 0444); err != nil {
		t.Fatalf("failed to write read-only version file: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chmod(versionPath, 0644)
	})

	appCli := newCLI(versionPath)
	err := appCli.Run(context.Background(), []string{
		"semver", "bump", "release", "--path", versionPath, "--no-auto-init",
	})

	if err == nil {
		t.Fatal("expected error due to save failure, got nil")
	}

	if !strings.Contains(err.Error(), "failed to save version") {
		t.Errorf("expected error message to contain 'failed to save version', got: %v", err)
	}
}

func TestCLI_BumpNextCmd_Errors(t *testing.T) {
	tests := []struct {
		name          string
		setup         func(dir string)
		args          []string
		expectedErr   string
		skipOnWindows bool
	}{
		{
			name: "fails if version file is invalid",
			setup: func(dir string) {
				_ = os.WriteFile(filepath.Join(dir, ".version"), []byte("not-a-version\n"), 0600)
			},
			args:        []string{"semver", "bump", "next"},
			expectedErr: "failed to read version",
		},
		{
			name: "fails if version file is not writable",
			setup: func(dir string) {
				path := filepath.Join(dir, ".version")
				_ = os.WriteFile(path, []byte("1.2.3-alpha\n"), 0444)
				_ = os.Chmod(path, 0444)
			},
			args:          []string{"semver", "bump", "next"},
			expectedErr:   "failed to save version",
			skipOnWindows: true, // permission simulation less reliable on Windows
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.skipOnWindows && testutils.IsWindows() {
				t.Skip("skipping test on Windows")
			}

			tmp := t.TempDir()
			tt.setup(tmp)

			versionPath := filepath.Join(tmp, ".version")
			appCli := newCLI(versionPath)

			err := appCli.Run(context.Background(), tt.args)
			if err == nil || !strings.Contains(err.Error(), tt.expectedErr) {
				t.Fatalf("expected error to contain %q, got: %v", tt.expectedErr, err)
			}
		})
	}
}

func TestCLI_BumpNextCmd_InitVersionFileFails(t *testing.T) {
	tmp := t.TempDir()
	protected := filepath.Join(tmp, "protected")

	// Make directory not writable
	if err := os.Mkdir(protected, 0555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(protected, 0755) })

	versionPath := filepath.Join(protected, ".version")
	appCli := newCLI(versionPath)

	err := appCli.Run(context.Background(), []string{
		"semver", "bump", "next", "--path", versionPath,
	})
	if err == nil || !strings.Contains(err.Error(), "permission denied") {
		t.Fatalf("expected permission denied error, got: %v", err)
	}
}

func TestCLI_BumpNextCmd_BumpNextFails(t *testing.T) {
	tmp := t.TempDir()
	versionPath := testutils.WriteTempVersionFile(t, tmp, "1.2.3")

	original := semver.BumpNextFunc
	semver.BumpNextFunc = func(v semver.SemVersion) (semver.SemVersion, error) {
		return semver.SemVersion{}, fmt.Errorf("forced BumpNext failure")
	}
	t.Cleanup(func() {
		semver.BumpNextFunc = original
	})

	appCli := newCLI(versionPath)
	err := appCli.Run(context.Background(), []string{
		"semver", "bump", "next", "--path", versionPath,
	})

	if err == nil || !strings.Contains(err.Error(), "failed to determine next version") {
		t.Fatalf("expected BumpNext failure, got: %v", err)
	}
}

func TestCLI_BumpNextCmd_SaveVersionFails(t *testing.T) {
	tmp := t.TempDir()
	versionPath := filepath.Join(tmp, ".version")

	// Write valid content
	if err := os.WriteFile(versionPath, []byte("1.2.3-alpha\n"), 0644); err != nil {
		t.Fatalf("failed to write version: %v", err)
	}

	// Make file read-only
	if err := os.Chmod(versionPath, 0444); err != nil {
		t.Fatalf("failed to chmod version file: %v", err)
	}
	t.Cleanup(func() { _ = os.Chmod(versionPath, 0644) }) // ensure cleanup

	appCli := newCLI(versionPath)
	err := appCli.Run(context.Background(), []string{
		"semver", "bump", "next", "--path", versionPath, "--no-auto-init",
	})

	if err == nil || !strings.Contains(err.Error(), "failed to save version") {
		t.Fatalf("expected error containing 'failed to save version', got: %v", err)
	}
}

func TestCLI_BumpNextCommand_InvalidLabel(t *testing.T) {
	if os.Getenv("TEST_SEMVER_BUMP_NEXT_INVALID_LABEL") == "1" {
		tmp := t.TempDir()
		versionPath := testutils.WriteTempVersionFile(t, tmp, "1.2.3")

		appCli := newCLI(versionPath)
		err := appCli.Run(context.Background(), []string{
			"semver", "bump", "next", "--label", "banana", "--path", versionPath,
		})
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		os.Exit(0) // ❌ shouldn't happen
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestCLI_BumpNextCommand_InvalidLabel")
	cmd.Env = append(os.Environ(), "TEST_SEMVER_BUMP_NEXT_INVALID_LABEL=1")
	output, err := cmd.CombinedOutput()

	if err == nil {
		t.Fatal("expected non-zero exit status")
	}

	expected := "invalid --label: must be 'patch', 'minor', or 'major'"
	if !strings.Contains(string(output), expected) {
		t.Errorf("expected output to contain %q, got: %q", expected, string(output))
	}
}

func TestCLI_BumpNextCmd_BumpByLabelFails(t *testing.T) {
	tmp := t.TempDir()
	versionPath := testutils.WriteTempVersionFile(t, tmp, "1.2.3")

	original := semver.BumpByLabelFunc
	semver.BumpByLabelFunc = func(v semver.SemVersion, label string) (semver.SemVersion, error) {
		return semver.SemVersion{}, fmt.Errorf("boom")
	}
	t.Cleanup(func() {
		semver.BumpByLabelFunc = original
	})

	appCli := newCLI(versionPath)
	err := appCli.Run(context.Background(), []string{
		"semver", "bump", "next", "--label", "patch", "--path", versionPath,
	})

	if err == nil || !strings.Contains(err.Error(), "failed to bump version with label") {
		t.Fatalf("expected error due to label bump failure, got: %v", err)
	}
}

func TestBumpReleaseCmd_ErrorOnInitVersionFile(t *testing.T) {
	tmp := t.TempDir()
	protectedDir := filepath.Join(tmp, "protected")
	if err := os.Mkdir(protectedDir, 0555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(protectedDir, 0755) })

	versionPath := filepath.Join(protectedDir, ".version")
	appCli := newCLI(versionPath)
	err := appCli.Run(context.Background(), []string{
		"semver", "bump", "release", "--path", versionPath,
	})

	if err == nil || !strings.Contains(err.Error(), "permission denied") {
		t.Errorf("expected permission denied error, got: %v", err)
	}
}

/* ------------------------------------------------------------------------- */
/* PRE COMMAND                                                               */
/* ------------------------------------------------------------------------- */
func TestCLI_PreCommand_StaticLabel(t *testing.T) {
	tmpDir := t.TempDir()
	versionPath := filepath.Join(tmpDir, ".version")
	appCli := newCLI(versionPath)
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
	appCli := newCLI(versionPath)

	testutils.RunCLITest(t, appCli, []string{"semver", "pre", "--label", "beta", "--inc"}, tmpDir)
	content := testutils.ReadTempVersionFile(t, tmpDir)
	if got := content; got != "1.2.3-beta.4" {
		t.Errorf("expected 1.2.3-beta.4, got %q", got)
	}
}

func TestCLI_PreCommand_AutoInitFeedback(t *testing.T) {
	tmpDir := t.TempDir()
	versionPath := filepath.Join(tmpDir, ".version")
	appCli := newCLI(versionPath)

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

	defaultPath := filepath.Join(tmp, ".version") // not used, but required by newCLI
	appCli := newCLI(defaultPath)

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

		appCli := newCLI(versionPath)
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

/* ------------------------------------------------------------------------- */
/* SHOW COMMAND                                                              */
/* ------------------------------------------------------------------------- */

func TestCLI_ShowCommand(t *testing.T) {
	tmpDir := t.TempDir()
	testutils.WriteTempVersionFile(t, tmpDir, "9.8.7")
	versionPath := filepath.Join(tmpDir, ".version")
	appCli := newCLI(versionPath)

	output, err := testutils.CaptureStdout(func() {
		testutils.RunCLITest(t, appCli, []string{"semver", "show"}, tmpDir)
	})
	if err != nil {
		t.Fatalf("Failed to capture stdout: %v", err)
	}

	if output != "9.8.7" {
		t.Errorf("expected output '9.8.7', got %q", output)
	}
}

func TestCLI_ShowCommand_NoAutoInit_MissingFile(t *testing.T) {
	if os.Getenv("TEST_SEMVER_NO_AUTO_INIT") == "1" {
		tmp := t.TempDir()
		versionPath := filepath.Join(tmp, ".version")

		appCli := newCLI(versionPath)
		err := appCli.Run(context.Background(), []string{"semver", "show", "--no-auto-init"})
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
	tmpDir := t.TempDir()
	testutils.WriteTempVersionFile(t, tmpDir, "1.2.3")
	versionPath := filepath.Join(tmpDir, ".version")
	appCli := newCLI(versionPath)

	output, err := testutils.CaptureStdout(func() {
		testutils.RunCLITest(t, appCli, []string{"semver", "show", "--no-auto-init"}, tmpDir)
	})
	if err != nil {
		t.Fatalf("Failed to capture stdout: %v", err)
	}

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

	appCli := newCLI(versionPath)
	err := appCli.Run(context.Background(), []string{"semver", "show"})
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
/* SET COMMAND                                                               */
/* ------------------------------------------------------------------------- */

func TestCLI_SetVersionCommandVariants(t *testing.T) {
	tmpDir := t.TempDir()
	versionPath := filepath.Join(tmpDir, ".version")
	appCli := newCLI(versionPath)

	tests := []struct {
		name     string
		args     []string
		expected string
	}{
		{"set version", []string{"semver", "set", "2.5.0"}, "2.5.0"},
		{"set with pre-release", []string{"semver", "set", "3.0.0", "--pre", "beta.2"}, "3.0.0-beta.2"},
		{"set with metadata", []string{"semver", "set", "1.0.0", "--meta", "001"}, "1.0.0+001"},
		{"set with pre and meta", []string{"semver", "set", "1.0.0", "--pre", "alpha.1", "--meta", "exp.sha.5114f85"}, "1.0.0-alpha.1+exp.sha.5114f85"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testutils.RunCLITest(t, appCli, tt.args, tmpDir)
			got := testutils.ReadTempVersionFile(t, tmpDir)
			if got != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, got)
			}
		})
	}
}

func TestCLI_SetVersionCommand_InvalidFormat(t *testing.T) {
	tmp := t.TempDir()
	appCli := newCLI(filepath.Join(tmp, ".version"))

	err := appCli.Run(context.Background(), []string{"semver", "set", "invalid.version"})
	if err == nil {
		t.Fatal("expected error due to invalid version format, got nil")
	}
	if !strings.Contains(err.Error(), "invalid version format") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestCLI_SetVersionCommand_MissingArgument(t *testing.T) {
	if os.Getenv("TEST_SEMVER_SET_MISSING_ARG") == "1" {
		tmp := t.TempDir()
		versionPath := filepath.Join(tmp, ".version")
		appCli := newCLI(versionPath)
		err := appCli.Run(context.Background(), []string{"semver", "set", "--path", versionPath})
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1) // expected non-zero exit
		}
		os.Exit(0) // ❌ should not happen
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestCLI_SetVersionCommand_MissingArgument")
	cmd.Env = append(os.Environ(), "TEST_SEMVER_SET_MISSING_ARG=1")
	output, err := cmd.CombinedOutput()

	if err == nil {
		t.Fatal("expected non-zero exit status")
	}

	expected := "missing required version argument"
	if !strings.Contains(string(output), expected) {
		t.Errorf("expected output to contain %q, got %q", expected, string(output))
	}
}

func TestCLI_SetVersionCommand_SaveError(t *testing.T) {
	tmp := t.TempDir()

	protectedDir := filepath.Join(tmp, "protected")
	if err := os.Mkdir(protectedDir, 0555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = os.Chmod(protectedDir, 0755)
	})

	versionPath := filepath.Join(protectedDir, ".version")
	appCli := newCLI(versionPath)
	err := appCli.Run(context.Background(), []string{
		"semver", "set", "3.0.0", "--path", versionPath,
	})
	if err == nil {
		t.Fatal("expected error due to save failure, got nil")
	}

	if !strings.Contains(err.Error(), "failed to save version") {
		t.Errorf("unexpected error message: %v", err)
	}
}

/* ------------------------------------------------------------------------- */
/* VALIDATE COMMAND                                                          */
/* ------------------------------------------------------------------------- */

func TestCLI_ValidateCommand_ValidCases(t *testing.T) {
	tmpDir := t.TempDir()
	versionPath := filepath.Join(tmpDir, ".version")
	appCli := newCLI(versionPath)

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
				testutils.RunCLITest(t, appCli, []string{"semver", "validate"}, tmpDir)
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
			tmp := t.TempDir()
			testutils.WriteTempVersionFile(t, tmp, tt.version)
			appCli := newCLI(filepath.Join(tmp, ".version"))
			err := appCli.Run(context.Background(), []string{"semver", "validate"})
			if err == nil || !strings.Contains(err.Error(), tt.expectedError) {
				t.Fatalf("expected error containing %q, got: %v", tt.expectedError, err)
			}
		})
	}
}

/* ------------------------------------------------------------------------- */
/* PLUGIN ADD COMMAND                                                        */
/* ------------------------------------------------------------------------- */

func TestPluginRegisterCmd_Success(t *testing.T) {
	// Set up a temporary directory for the version file and config
	tmpDir := t.TempDir()
	versionPath := filepath.Join(tmpDir, ".version")
	configPath := filepath.Join(tmpDir, ".semver.yaml")

	// Create .semver.yaml with the required path field
	configContent := `path: .version`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to create .semver.yaml: %v", err)
	}

	// Create a subdirectory for the plugin to hold the plugin.yaml file
	pluginDir := filepath.Join(tmpDir, "mock-plugin")
	if err := os.Mkdir(pluginDir, 0755); err != nil {
		t.Fatalf("failed to create plugin directory: %v", err)
	}

	// Create a valid plugin.yaml file inside the plugin directory
	pluginPath := filepath.Join(pluginDir, "plugin.yaml")
	pluginContent := `name: mock-plugin
version: 1.0.0
description: Mock Plugin
author: Test Author
repository: https://github.com/test/repo
entry: mock-entry`

	if err := os.WriteFile(pluginPath, []byte(pluginContent), 0644); err != nil {
		t.Fatalf("failed to create plugin.yaml: %v", err)
	}

	// Prepare and run the CLI command
	appCli := newCLI(versionPath)

	// Ensure the plugin directory is passed correctly
	if _, err := os.Stat(pluginDir); os.IsNotExist(err) {
		t.Fatalf("plugin directory does not exist at %s", pluginDir)
	}

	// Run the command, ensuring we pass the correct plugin directory
	output, _ := testutils.CaptureStdout(func() {
		testutils.RunCLITest(t, appCli, []string{
			"semver", "plugin", "add", "--path", pluginDir,
			"--plugin-dir", tmpDir}, tmpDir)
	})

	// Check the output for success
	if !strings.Contains(output, "Plugin \"mock-plugin\" registered successfully.") {
		t.Fatalf("expected success message, got: %s", output)
	}
}

func TestPluginRegisterCmd_MissingPathArgument(t *testing.T) {
	if os.Getenv("TEST_SEMVER_PLUGIN_MISSING_PATH") == "1" {
		tmp := t.TempDir()
		versionPath := filepath.Join(tmp, ".version")
		appCli := newCLI(versionPath)

		// Run the CLI command with missing --path argument
		err := appCli.Run(context.Background(), []string{"semver", "plugin", "add"})
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1) // expected non-zero exit
		}
		os.Exit(0) // ❌ should not happen
	}

	// Run the test with the custom environment variable to trigger the error condition
	cmd := exec.Command(os.Args[0], "-test.run=TestPluginRegisterCmd_MissingPathArgument")
	cmd.Env = append(os.Environ(), "TEST_SEMVER_PLUGIN_MISSING_PATH=1")
	output, err := cmd.CombinedOutput()

	// Ensure that the test exits with an error
	if err == nil {
		t.Fatal("expected non-zero exit status")
	}

	// Define the expected error message
	expected := "missing --path (or --url) for plugin registration"

	// Check if the expected error message is in the captured output
	if !strings.Contains(string(output), expected) {
		t.Errorf("expected output to contain %q, got %q", expected, string(output))
	}
}

/* ------------------------------------------------------------------------- */
/* PLUGIN LIST COMMAND                                                   */
/* ------------------------------------------------------------------------- */

func TestPluginListCmd(t *testing.T) {
	// Set up a temporary directory for the config file and plugin directory
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".semver.yaml")

	// Test with no plugins registered
	// Create an empty .semver.yaml to simulate no plugins
	err := os.WriteFile(configPath, []byte("plugins: []\n"), 0644)
	if err != nil {
		t.Fatalf("failed to create .semver.yaml: %v", err)
	}

	// Set the config path and simulate calling the plugin list command
	appCli := newCLI(configPath)

	output, err := testutils.CaptureStdout(func() {
		testutils.RunCLITest(t, appCli, []string{"semver", "plugin", "list"}, tmpDir)
	})
	if err != nil {
		t.Fatalf("Failed to capture stdout: %v", err)
	}

	// Check if the expected message is printed
	expected := "No plugins registered."
	if !strings.Contains(output, expected) {
		t.Errorf("expected output to contain %q, got %q", expected, output)
	}

	// Test with some plugins registered
	// Create a .semver.yaml with sample plugins
	pluginsContent := `
plugins:
  - name: mock-plugin-1
    path: /path/to/mock-plugin-1
    enabled: true
  - name: mock-plugin-2
    path: /path/to/mock-plugin-2
    enabled: false
`
	err = os.WriteFile(configPath, []byte(pluginsContent), 0644)
	if err != nil {
		t.Fatalf("failed to write .semver.yaml: %v", err)
	}

	// Capture the output of the plugin list command again
	output, err = testutils.CaptureStdout(func() {
		testutils.RunCLITest(t, appCli, []string{"semver", "plugin", "list"}, tmpDir)
	})
	if err != nil {
		t.Fatalf("Failed to capture stdout: %v", err)
	}

	// Check if the output contains information about the plugins
	expectedPlugins := []string{
		"Name: mock-plugin-1, Path: /path/to/mock-plugin-1, Enabled: true",
		"Name: mock-plugin-2, Path: /path/to/mock-plugin-2, Enabled: false",
	}

	for _, exp := range expectedPlugins {
		if !strings.Contains(output, exp) {
			t.Errorf("expected output to contain %q, got %q", exp, output)
		}
	}
}

func TestPluginListCmd_LoadConfigError(t *testing.T) {
	// Create a mock of the LoadConfig function that returns an error
	originalLoadConfig := config.LoadConfigFn
	defer func() {
		// Restore the original LoadConfig function after the test
		config.LoadConfigFn = originalLoadConfig
	}()

	// Mock the LoadConfig function to simulate an error
	config.LoadConfigFn = func() (*config.Config, error) {
		return nil, fmt.Errorf("failed to load configuration")
	}

	// Set up a temporary directory for the config file (not used here, since LoadConfig will fail)
	tmpDir := t.TempDir()

	// Set the config path and simulate calling the plugin list command
	appCli := newCLI(tmpDir)

	// Capture the output of the plugin list command again
	output, err := testutils.CaptureStdout(func() {
		err := appCli.Run(context.Background(), []string{"semver", "plugin", "list"})
		// Capture the actual error during execution
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}
	})

	if err != nil {
		t.Fatalf("Failed to capture stdout: %v", err)
	}

	// Check if the error message was properly printed
	expectedErrorMessage := "failed to load configuration"
	if !strings.Contains(output, expectedErrorMessage) {
		t.Errorf("Expected error message to contain %q, but got: %q", expectedErrorMessage, output)
	}
}

/* ------------------------------------------------------------------------- */
/* PLUGIN REMOVE COMMAND                                                     */
/* ------------------------------------------------------------------------- */

func TestPluginRemoveCmd_DeleteFolderVariants(t *testing.T) {
	pluginName := "mock-plugin"

	tests := []struct {
		name          string
		deleteFolder  bool
		expectDeleted bool
	}{
		{
			name:          "delete-folder=false",
			deleteFolder:  false,
			expectDeleted: false,
		},
		{
			name:          "delete-folder=true",
			deleteFolder:  true,
			expectDeleted: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			// Prepare paths
			pluginsRoot := filepath.Join(tmpDir, ".semver-plugins")
			pluginDir := filepath.Join(pluginsRoot, pluginName)
			configPath := filepath.Join(tmpDir, ".semver.yaml")

			// Create dummy plugin dir
			if err := os.MkdirAll(pluginDir, 0755); err != nil {
				t.Fatalf("failed to create plugin directory: %v", err)
			}

			// Write .semver.yaml config
			content := `plugins:
  - name: mock-plugin
    path: /some/path
    enabled: true`
			if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
				t.Fatalf("failed to write config file: %v", err)
			}

			// Build CLI args
			args := []string{"semver", "plugin", "remove", "--name", pluginName}
			if tt.deleteFolder {
				args = append(args, "--delete-folder")
			}

			appCli := newCLI(configPath)
			output, err := testutils.CaptureStdout(func() {
				testutils.RunCLITest(t, appCli, args, tmpDir)
			})
			if err != nil {
				t.Fatalf("CLI run failed: %v", err)
			}

			// Check output
			if tt.deleteFolder {
				exp := fmt.Sprintf("✅ Plugin %q and its directory removed successfully.", pluginName)
				if !strings.Contains(output, exp) {
					t.Errorf("expected output to contain %q, got:\n%s", exp, output)
				}
			} else {
				exp := fmt.Sprintf("✅ Plugin %q removed, but its directory is preserved.", pluginName)
				if !strings.Contains(output, exp) {
					t.Errorf("expected output to contain %q, got:\n%s", exp, output)
				}
			}

			// Check if directory was deleted or not
			_, err = os.Stat(pluginDir)
			if tt.expectDeleted {
				if !os.IsNotExist(err) {
					t.Errorf("expected plugin directory to be deleted, but it still exists")
				}
			} else {
				if err != nil {
					t.Errorf("expected plugin directory to exist, got: %v", err)
				}
			}

			// Verify plugin is disabled in config
			cfg, err := config.LoadConfigFn()
			if err != nil {
				t.Fatalf("failed to reload config: %v", err)
			}
			found := false
			for _, plugin := range cfg.Plugins {
				if plugin.Name == pluginName {
					found = true
					if plugin.Enabled {
						t.Errorf("expected plugin %q to be disabled, but it's still enabled", pluginName)
					}
					break
				}
			}
			if !found {
				t.Errorf("plugin %q not found in config", pluginName)
			}
		})
	}
}

func TestPluginRemoveCmd_DeleteFolderFailure(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("permission-based RemoveAll failure is unreliable on Windows")
	}

	tmpDir := t.TempDir()
	pluginName := "mock-plugin"
	pluginDir := filepath.Join(tmpDir, ".semver-plugins", pluginName)
	pluginConfigPath := filepath.Join(tmpDir, ".semver.yaml")

	if err := os.MkdirAll(pluginDir, 0755); err != nil {
		t.Fatalf("failed to create plugin directory: %v", err)
	}

	protectedFile := filepath.Join(pluginDir, "protected.txt")
	if err := os.WriteFile(protectedFile, []byte("protected"), 0400); err != nil {
		t.Fatalf("failed to create protected file: %v", err)
	}
	if err := os.Chmod(pluginDir, 0500); err != nil {
		t.Fatalf("failed to chmod plugin dir: %v", err)
	}

	defer func() {
		_ = os.Chmod(protectedFile, 0600)
		_ = os.Chmod(pluginDir, 0700)
	}()

	content := `plugins:
  - name: mock-plugin
    path: /some/path
    enabled: true`
	if err := os.WriteFile(pluginConfigPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	appCli := newCLI(pluginConfigPath)

	var cliErr error
	_, captureErr := testutils.CaptureStdout(func() {
		cliErr = testutils.RunCLITestAllowError(t, appCli, []string{
			"semver", "plugin", "remove",
			"--name", pluginName,
			"--delete-folder",
		}, tmpDir)
	})
	if captureErr != nil {
		t.Fatalf("failed to capture output: %v", captureErr)
	}

	_ = os.Chmod(protectedFile, 0600)
	_ = os.Chmod(pluginDir, 0700)

	if cliErr == nil {
		t.Fatal("expected error but got nil")
	}

	expectedMsg := "failed to remove plugin directory"
	if !strings.Contains(cliErr.Error(), expectedMsg) {
		t.Errorf("expected error message to contain %q, got: %v", expectedMsg, cliErr)
	}

}

func TestCLI_PluginRemove_MissingName(t *testing.T) {
	if os.Getenv("TEST_PLUGIN_REMOVE_MISSING_NAME") == "1" {
		tmp := t.TempDir()

		// Write valid .semver.yaml with 1 plugin (won't be used, but still required)
		configPath := filepath.Join(tmp, ".semver.yaml")
		content := `plugins:
  - name: mock-plugin
    path: /some/path
    enabled: true`
		if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
			fmt.Fprintln(os.Stderr, "failed to write config:", err)
			os.Exit(1)
		}

		appCli := newCLI(configPath)

		// Run command WITHOUT --name (should trigger the validation)
		err := appCli.Run(context.Background(), []string{
			"semver", "plugin", "remove", "--path", configPath,
		})
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		// Shouldn't reach here
		os.Exit(0)
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestCLI_PluginRemove_MissingName")
	cmd.Env = append(os.Environ(), "TEST_PLUGIN_REMOVE_MISSING_NAME=1")
	output, err := cmd.CombinedOutput()

	if err == nil {
		t.Fatal("expected non-zero exit status")
	}

	expected := "please provide a plugin name to remove"
	if !strings.Contains(string(output), expected) {
		t.Errorf("expected output to contain %q, got:\n%s", expected, output)
	}
}

func TestPluginRemoveCmd_LoadConfigError(t *testing.T) {
	// Mock the LoadConfig function to simulate an error
	originalLoadConfig := config.LoadConfigFn
	defer func() {
		config.LoadConfigFn = originalLoadConfig
	}()

	config.LoadConfigFn = func() (*config.Config, error) {
		return nil, fmt.Errorf("failed to load configuration")
	}

	tmpDir := t.TempDir()
	pluginConfigPath := filepath.Join(tmpDir, ".semver.yaml")

	// Run the plugin remove command
	appCli := newCLI(pluginConfigPath)

	// Capture the output
	output, err := testutils.CaptureStdout(func() {
		testutils.RunCLITest(t, appCli, []string{"semver", "plugin", "remove", "--name", "mock-plugin"}, tmpDir)
	})

	if err != nil {
		t.Fatalf("Failed to capture stdout: %v", err)
	}

	// Expected error message
	expected := "failed to load configuration"
	if !strings.Contains(output, expected) {
		t.Errorf("expected output to contain %q, got %q", expected, output)
	}
}

func TestPluginRemoveCmd_PluginNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	pluginConfigPath := filepath.Join(tmpDir, ".semver.yaml")

	// Create a dummy .semver.yaml configuration file with no plugin
	content := `plugins: []`
	if err := os.WriteFile(pluginConfigPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create .semver.yaml: %v", err)
	}

	// Run the plugin remove command with a non-existent plugin
	appCli := newCLI(pluginConfigPath)

	// Capture the output
	output, err := testutils.CaptureStdout(func() {
		testutils.RunCLITest(t, appCli, []string{"semver", "plugin", "remove", "--name", "mock-plugin"}, tmpDir)
	})

	if err != nil {
		t.Fatalf("Failed to capture stdout: %v", err)
	}

	// Expected error message
	expected := "plugin \"mock-plugin\" not found"
	if !strings.Contains(output, expected) {
		t.Errorf("expected output to contain %q, got %q", expected, output)
	}
}

func TestPluginRemoveCmd_SaveConfigError(t *testing.T) {
	// Mock the SaveConfig function to simulate an error
	originalSaveConfig := config.SaveConfigFn
	defer func() {
		config.SaveConfigFn = originalSaveConfig
	}()

	config.SaveConfigFn = func(cfg *config.Config) error {
		return fmt.Errorf("failed to save updated configuration")
	}

	tmpDir := t.TempDir()
	pluginConfigPath := filepath.Join(tmpDir, ".semver.yaml")

	// Create a dummy .semver.yaml configuration file
	content := `plugins:
  - name: mock-plugin
    path: /path/to/plugin
    enabled: true`
	if err := os.WriteFile(pluginConfigPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create .semver.yaml: %v", err)
	}

	// Run the plugin remove command
	appCli := newCLI(pluginConfigPath)

	// Capture the output
	output, err := testutils.CaptureStdout(func() {
		testutils.RunCLITest(t, appCli, []string{"semver", "plugin", "remove", "--name", "mock-plugin"}, tmpDir)
	})

	if err != nil {
		t.Fatalf("Failed to capture stdout: %v", err)
	}

	// Expected error message
	expected := "failed to save updated configuration"
	if !strings.Contains(output, expected) {
		t.Errorf("expected output to contain %q, got %q", expected, output)
	}
}
