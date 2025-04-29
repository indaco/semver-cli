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

	"github.com/indaco/semver-cli/api/v0/extensions"
	"github.com/indaco/semver-cli/internal/config"
	commitparser "github.com/indaco/semver-cli/internal/plugins/commit-parser"
	"github.com/indaco/semver-cli/internal/plugins/commit-parser/gitlog"
	"github.com/indaco/semver-cli/internal/semver"
	"github.com/indaco/semver-cli/internal/testutils"
)

/* ------------------------------------------------------------------------- */
/* INIT COMMAND                                                              */
/* ------------------------------------------------------------------------- */

func TestCLI_InitCommand_CreatesFile(t *testing.T) {
	tmp := t.TempDir()
	versionPath := filepath.Join(tmp, ".version")

	// Prepare and run the CLI command
	cfg := &config.Config{Path: versionPath}
	appCli := newCLI(cfg)

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

	// Prepare and run the CLI command
	cfg := &config.Config{Path: versionPath}
	appCli := newCLI(cfg)

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

	// Prepare and run the CLI command
	cfg := &config.Config{Path: versionPath}
	appCli := newCLI(cfg)

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

	// Prepare and run the CLI command
	cfg := &config.Config{Path: path}
	appCli := newCLI(cfg)

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

			// Prepare and run the CLI command
			cfg := &config.Config{Path: protectedPath}
			appCli := newCLI(cfg)

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
	cfg := &config.Config{Path: versionPath}
	appCli := newCLI(cfg)

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

			// Prepare and run the CLI command
			cfg := &config.Config{Path: versionPath}
			appCli := newCLI(cfg)

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

			// Prepare and run the CLI command
			cfg := &config.Config{Path: versionPath}
			appCli := newCLI(cfg)

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

			// Prepare and run the CLI command
			cfg := &config.Config{Path: versionPath}
			appCli := newCLI(cfg)

			testutils.WriteTempVersionFile(t, tmpDir, tt.initial)
			testutils.RunCLITest(t, appCli, tt.args, tmpDir)

			got := testutils.ReadTempVersionFile(t, tmpDir)
			if got != tt.expected {
				t.Errorf("expected version %q, got %q", tt.expected, got)
			}
		})
	}
}

func TestCLI_BumpNextCmd_InferredBump(t *testing.T) {
	tmp := t.TempDir()
	versionPath := testutils.WriteTempVersionFile(t, tmp, "1.2.3")

	// Save original function and restore later
	originalInfer := tryInferBumpTypeFromCommitParserPluginFn
	defer func() { tryInferBumpTypeFromCommitParserPluginFn = originalInfer }()

	// Mock the inference to simulate an inferred "minor" bump
	tryInferBumpTypeFromCommitParserPluginFn = func(since, until string) string {
		return "minor"
	}

	cfg := &config.Config{Path: versionPath, Plugins: &config.PluginConfig{CommitParser: true}}
	appCli := newCLI(cfg)

	err := appCli.Run(context.Background(), []string{
		"semver", "bump", "next", "--path", versionPath,
	})

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	got := testutils.ReadTempVersionFile(t, tmp)
	want := "1.3.0"
	if got != want {
		t.Errorf("expected bumped version %q, got %q", want, got)
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

			// Prepare and run the CLI command
			cfg := &config.Config{Path: versionPath}
			appCli := newCLI(cfg)

			testutils.WriteTempVersionFile(t, tmpDir, tt.initial)
			testutils.RunCLITest(t, appCli, tt.args, tmpDir)

			got := testutils.ReadTempVersionFile(t, tmpDir)
			if got != tt.want {
				t.Errorf("expected %q, got %q", tt.want, got)
			}
		})
	}
}

func TestCLI_BumpNextCmd_InferredPromotion(t *testing.T) {
	tmp := t.TempDir()
	versionPath := testutils.WriteTempVersionFile(t, tmp, "1.2.3-beta.1")

	originalInfer := tryInferBumpTypeFromCommitParserPluginFn
	defer func() { tryInferBumpTypeFromCommitParserPluginFn = originalInfer }()

	tryInferBumpTypeFromCommitParserPluginFn = func(since, until string) string {
		return "minor"
	}

	cfg := &config.Config{Path: versionPath, Plugins: &config.PluginConfig{CommitParser: true}}
	appCli := newCLI(cfg)

	err := appCli.Run(context.Background(), []string{
		"semver", "bump", "next", "--path", versionPath,
	})

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	got := testutils.ReadTempVersionFile(t, tmp)
	want := "1.2.3" // Promotion, not minor bump
	if got != want {
		t.Errorf("expected promoted version %q, got %q", want, got)
	}
}

func TestCLI_BumpNextCmd_PromotePreReleaseWithPreserveMeta(t *testing.T) {
	tmp := t.TempDir()
	versionPath := testutils.WriteTempVersionFile(t, tmp, "1.2.3-beta.2+ci.99")

	// Override tryInferBumpTypeFromCommitParserPlugin
	originalInfer := tryInferBumpTypeFromCommitParserPluginFn
	tryInferBumpTypeFromCommitParserPluginFn = func(since, until string) string {
		return "minor" // Force a non-empty inference so that promotePreRelease is called
	}
	t.Cleanup(func() { tryInferBumpTypeFromCommitParserPluginFn = originalInfer })

	cfg := &config.Config{Path: versionPath}
	appCli := newCLI(cfg)

	err := appCli.Run(context.Background(), []string{
		"semver", "bump", "next", "--path", versionPath, "--preserve-meta",
	})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	got := testutils.ReadTempVersionFile(t, tmp)
	want := "1.2.3+ci.99"
	if got != want {
		t.Errorf("expected promoted version with metadata %q, got %q", want, got)
	}
}

func TestCLI_BumpNextCmd_InferredBumpFails(t *testing.T) {
	tmp := t.TempDir()
	versionPath := testutils.WriteTempVersionFile(t, tmp, "1.2.3")

	originalBumpByLabel := semver.BumpByLabelFunc
	originalInferFunc := tryInferBumpTypeFromCommitParserPluginFn

	// Force BumpByLabelFunc to fail
	semver.BumpByLabelFunc = func(v semver.SemVersion, label string) (semver.SemVersion, error) {
		return semver.SemVersion{}, fmt.Errorf("forced inferred bump failure")
	}

	// Force inference to return something
	tryInferBumpTypeFromCommitParserPluginFn = func(since, until string) string {
		return "minor"
	}

	t.Cleanup(func() {
		semver.BumpByLabelFunc = originalBumpByLabel
		tryInferBumpTypeFromCommitParserPluginFn = originalInferFunc
	})

	// Prepare and run CLI
	cfg := &config.Config{Path: versionPath}
	appCli := newCLI(cfg)

	err := appCli.Run(context.Background(), []string{
		"semver", "bump", "next", "--path", versionPath,
	})

	if err == nil || !strings.Contains(err.Error(), "failed to bump inferred version") {
		t.Fatalf("expected error about inferred bump failure, got: %v", err)
	}
}

func TestTryInferBumpTypeFromCommitParserPlugin_GetCommitsError(t *testing.T) {
	testutils.WithMock(func() {
		// Mock GetCommits to fail
		originalGetCommits := gitlog.GetCommitsFn
		originalParser := commitparser.GetCommitParserFn

		gitlog.GetCommitsFn = func(since, until string) ([]string, error) {
			return nil, fmt.Errorf("simulated gitlog error")
		}
		commitparser.GetCommitParserFn = func() commitparser.CommitParser {
			return testutils.MockCommitParser{} // Return any parser
		}

		t.Cleanup(func() {
			gitlog.GetCommitsFn = originalGetCommits
			commitparser.GetCommitParserFn = originalParser
		})
	}, func() {
		label := tryInferBumpTypeFromCommitParserPlugin("", "")
		if label != "" {
			t.Errorf("expected empty label on gitlog error, got %q", label)
		}
	})
}

func TestTryInferBumpTypeFromCommitParserPlugin_ParserError(t *testing.T) {
	testutils.WithMock(
		func() {
			// Setup mocks
			gitlog.GetCommitsFn = func(since, until string) ([]string, error) {
				return []string{"fix: something"}, nil
			}
			commitparser.GetCommitParserFn = func() commitparser.CommitParser {
				return testutils.MockCommitParser{Err: fmt.Errorf("parser error")}
			}
		},
		func() {
			label := tryInferBumpTypeFromCommitParserPlugin("", "")
			if label != "" {
				t.Errorf("expected empty label on parser error, got %q", label)
			}
		},
	)
}

func TestTryInferBumpTypeFromCommitParserPlugin_Success(t *testing.T) {
	testutils.WithMock(
		func() {
			// Setup mocks
			gitlog.GetCommitsFn = func(since, until string) ([]string, error) {
				return []string{"feat: add feature"}, nil
			}
			commitparser.GetCommitParserFn = func() commitparser.CommitParser {
				return testutils.MockCommitParser{Label: "minor"}
			}
		},
		func() {
			label := tryInferBumpTypeFromCommitParserPlugin("", "")
			if label != "minor" {
				t.Errorf("expected label 'minor', got %q", label)
			}
		},
	)
}

func TestBumpReleaseCmd_ErrorOnReadVersion(t *testing.T) {
	tmp := t.TempDir()
	versionPath := testutils.WriteTempVersionFile(t, tmp, "invalid-version")

	// Prepare and run the CLI command
	cfg := &config.Config{Path: versionPath}
	appCli := newCLI(cfg)
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

	// Prepare and run the CLI command
	cfg := &config.Config{Path: versionPath}
	appCli := newCLI(cfg)
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

			// Prepare and run the CLI command
			cfg := &config.Config{Path: versionPath}
			appCli := newCLI(cfg)

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

	// Prepare and run the CLI command
	cfg := &config.Config{Path: versionPath}
	appCli := newCLI(cfg)

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

	// Prepare and run the CLI command
	cfg := &config.Config{Path: versionPath}
	appCli := newCLI(cfg)

	err := appCli.Run(context.Background(), []string{
		"semver", "bump", "next", "--path", versionPath, "--no-infer",
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

	// Prepare and run the CLI command
	cfg := &config.Config{Path: versionPath}
	appCli := newCLI(cfg)

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

		// Prepare and run the CLI command
		cfg := &config.Config{Path: versionPath}
		appCli := newCLI(cfg)

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

	// Prepare and run the CLI command
	cfg := &config.Config{Path: versionPath}
	appCli := newCLI(cfg)

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

	// Prepare and run the CLI command
	cfg := &config.Config{Path: versionPath}
	appCli := newCLI(cfg)

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

	// Prepare and run the CLI command
	cfg := &config.Config{Path: versionPath}
	appCli := newCLI(cfg)
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
	appCli := newCLI(cfg)

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
	appCli := newCLI(cfg)

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
	appCli := newCLI(cfg)

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
		appCli := newCLI(cfg)

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

	// Prepare and run the CLI command
	cfg := &config.Config{Path: versionPath}
	appCli := newCLI(cfg)

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

		// Prepare and run the CLI command
		cfg := &config.Config{Path: versionPath}
		appCli := newCLI(cfg)

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

	// Prepare and run the CLI command
	cfg := &config.Config{Path: versionPath}
	appCli := newCLI(cfg)

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

	// Prepare and run the CLI command
	cfg := &config.Config{Path: versionPath}
	appCli := newCLI(cfg)

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

	// Prepare and run the CLI command
	cfg := &config.Config{Path: versionPath}
	appCli := newCLI(cfg)

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
	tmpDir := t.TempDir()
	versionPath := filepath.Join(tmpDir, ".version")

	// Prepare and run the CLI command
	cfg := &config.Config{Path: versionPath}
	appCli := newCLI(cfg)

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

		// Prepare and run the CLI command
		cfg := &config.Config{Path: versionPath}
		appCli := newCLI(cfg)

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

	// Prepare and run the CLI command
	cfg := &config.Config{Path: versionPath}
	appCli := newCLI(cfg)

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

	// Prepare and run the CLI command
	cfg := &config.Config{Path: versionPath}
	appCli := newCLI(cfg)

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
			tmpDir := t.TempDir()
			testutils.WriteTempVersionFile(t, tmpDir, tt.version)
			versionPath := filepath.Join(tmpDir, ".version")

			// Prepare and run the CLI command
			cfg := &config.Config{Path: versionPath}
			appCli := newCLI(cfg)

			err := appCli.Run(context.Background(), []string{"semver", "validate"})
			if err == nil || !strings.Contains(err.Error(), tt.expectedError) {
				t.Fatalf("expected error containing %q, got: %v", tt.expectedError, err)
			}
		})
	}
}

/* ------------------------------------------------------------------------- */
/* EXTENSION INSTALL COMMAND                                                 */
/* ------------------------------------------------------------------------- */

func TestExtensionIstallCmd_Success(t *testing.T) {
	// Set up a temporary directory for the version file and config
	tmpDir := t.TempDir()
	versionPath := filepath.Join(tmpDir, ".version")
	configPath := filepath.Join(tmpDir, ".semver.yaml")

	// Create .semver.yaml with the required path field
	configContent := `path: .version`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to create .semver.yaml: %v", err)
	}

	// Create a subdirectory for the extension to hold the extension.yaml file
	extensionDir := filepath.Join(tmpDir, "mock-extension")
	if err := os.Mkdir(extensionDir, 0755); err != nil {
		t.Fatalf("failed to create extension directory: %v", err)
	}

	// Create a valid extension.yaml file inside the extension directory
	extensionPath := filepath.Join(extensionDir, "extension.yaml")
	extensionContent := `name: mock-extension
version: 1.0.0
description: Mock Extension
author: Test Author
repository: https://github.com/test/repo
entry: mock-entry`

	if err := os.WriteFile(extensionPath, []byte(extensionContent), 0644); err != nil {
		t.Fatalf("failed to create extension.yaml: %v", err)
	}

	// Prepare and run the CLI command
	cfg := &config.Config{Path: versionPath}
	appCli := newCLI(cfg)

	// Ensure the extension directory is passed correctly
	if _, err := os.Stat(extensionDir); os.IsNotExist(err) {
		t.Fatalf("extension directory does not exist at %s", extensionDir)
	}

	// Run the command, ensuring we pass the correct extension directory
	output, _ := testutils.CaptureStdout(func() {
		testutils.RunCLITest(t, appCli, []string{
			"semver", "extension", "install", "--path", extensionDir,
			"--extension-dir", tmpDir}, tmpDir)
	})

	// Check the output for success
	if !strings.Contains(output, "Extension \"mock-extension\" registered successfully.") {
		t.Fatalf("expected success message, got: %s", output)
	}
}

func TestExtensionRegisterCmd_MissingPathArgument(t *testing.T) {
	if os.Getenv("TEST_SEMVER_EXTENSION_MISSING_PATH") == "1" {
		tmp := t.TempDir()
		versionPath := filepath.Join(tmp, ".version")

		// Prepare and run the CLI command
		cfg := &config.Config{Path: versionPath}
		appCli := newCLI(cfg)

		// Run the CLI command with missing --path argument
		err := appCli.Run(context.Background(), []string{"semver", "extension", "install"})
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1) // expected non-zero exit
		}
		os.Exit(0) // ❌ should not happen
	}

	// Run the test with the custom environment variable to trigger the error condition
	cmd := exec.Command(os.Args[0], "-test.run=TestExtensionRegisterCmd_MissingPathArgument")
	cmd.Env = append(os.Environ(), "TEST_SEMVER_EXTENSION_MISSING_PATH=1")
	output, err := cmd.CombinedOutput()

	// Ensure that the test exits with an error
	if err == nil {
		t.Fatal("expected non-zero exit status")
	}

	// Define the expected error message
	expected := "missing --path (or --url) for extension registration"

	// Check if the expected error message is in the captured output
	if !strings.Contains(string(output), expected) {
		t.Errorf("expected output to contain %q, got %q", expected, string(output))
	}
}

/* ------------------------------------------------------------------------- */
/* EXTENSION LIST COMMAND                                                    */
/* ------------------------------------------------------------------------- */

func TestExtensionListCmd(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".semver.yaml")

	// Test with no plugins
	err := os.WriteFile(configPath, []byte("extensions: []\n"), 0644)
	if err != nil {
		t.Fatalf("failed to create .semver.yaml: %v", err)
	}

	// Prepare and run the CLI command
	cfg := &config.Config{Path: configPath}
	appCli := newCLI(cfg)

	output, err := testutils.CaptureStdout(func() {
		testutils.RunCLITest(t, appCli, []string{"semver", "extension", "list"}, tmpDir)
	})
	if err != nil {
		t.Fatalf("Failed to capture stdout: %v", err)
	}

	if !strings.Contains(output, "No extensions registered.") {
		t.Errorf("expected output to contain 'No extensions registered.', got:\n%s", output)
	}

	// Add plugin entries
	extensionsContent := `
extensions:
  - name: mock-extension-1
    path: /path/to/mock-extension-1
    enabled: true
  - name: mock-extension-2
    path: /path/to/mock-extension-2
    enabled: false
`
	err = os.WriteFile(configPath, []byte(extensionsContent), 0644)
	if err != nil {
		t.Fatalf("failed to write .semver.yaml: %v", err)
	}

	output, err = testutils.CaptureStdout(func() {
		testutils.RunCLITest(t, appCli, []string{"semver", "extension", "list"}, tmpDir)
	})
	if err != nil {
		t.Fatalf("Failed to capture stdout: %v", err)
	}

	expectedRows := []string{
		"mock-extension-1",
		"true",
		"mock-extension-2",
		"false",
		"(no metadata)",
	}

	for _, expected := range expectedRows {
		if !strings.Contains(output, expected) {
			t.Errorf("expected output to contain %q, got:\n%s", expected, output)
		}
	}
}

func TestExtensionListCmd_LoadConfigError(t *testing.T) {
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

	// Prepare and run the CLI command
	cfg := &config.Config{Path: tmpDir}
	appCli := newCLI(cfg)

	// Capture the output of the plugin list command again
	output, err := testutils.CaptureStdout(func() {
		err := appCli.Run(context.Background(), []string{"semver", "extension", "list"})
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

func TestExtensionListCmd_WithRegisteredMetadata(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".semver.yaml")

	extensionName := "test-extension"

	// Write .semver.yaml with extension that matches registered metadata
	content := fmt.Sprintf(`extensions:
  - name: %s
    path: /some/path
    enabled: true
`, extensionName)
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	// Register mock metadata plugin
	extensions.ResetExtension()
	t.Cleanup(extensions.ResetExtension)

	extensions.RegisterExtension(testutils.MockExtension{
		NameValue:        extensionName,
		VersionValue:     "9.9.9",
		DescriptionValue: "Registered test extension",
	})

	// Prepare and run the CLI command
	cfg := &config.Config{Path: configPath}
	appCli := newCLI(cfg)

	output, err := testutils.CaptureStdout(func() {
		testutils.RunCLITest(t, appCli, []string{"semver", "extension", "list"}, tmpDir)
	})
	if err != nil {
		t.Fatalf("Failed to capture stdout: %v", err)
	}

	// Ensure metadata was printed
	expectedValues := []string{
		extensionName,
		"9.9.9",
		"true",
		"Registered test extension",
	}
	for _, expected := range expectedValues {
		if !strings.Contains(output, expected) {
			t.Errorf("expected output to contain %q, got:\n%s", expected, output)
		}
	}
}

/* ------------------------------------------------------------------------- */
/* EXTENSION REMOVE COMMAND                                                  */
/* ------------------------------------------------------------------------- */

func TestExtensionRemoveCmd_DeleteFolderVariants(t *testing.T) {
	extensionName := "mock-extension"

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
			extensionsRoot := filepath.Join(tmpDir, ".semver-extensions")
			extensionDir := filepath.Join(extensionsRoot, extensionName)
			configPath := filepath.Join(tmpDir, ".semver.yaml")

			// Create dummy extension dir
			if err := os.MkdirAll(extensionDir, 0755); err != nil {
				t.Fatalf("failed to create extension directory: %v", err)
			}

			// Write .semver.yaml config
			content := `extensions:
  - name: mock-extension
    path: /some/path
    enabled: true`
			if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
				t.Fatalf("failed to write config file: %v", err)
			}

			// Build CLI args
			args := []string{"semver", "extension", "remove", "--name", extensionName}
			if tt.deleteFolder {
				args = append(args, "--delete-folder")
			}

			// Prepare and run the CLI command
			cfg := &config.Config{Path: configPath}
			appCli := newCLI(cfg)

			output, err := testutils.CaptureStdout(func() {
				testutils.RunCLITest(t, appCli, args, tmpDir)
			})
			if err != nil {
				t.Fatalf("CLI run failed: %v", err)
			}

			// Check output
			if tt.deleteFolder {
				exp := fmt.Sprintf("✅ Extension %q and its directory removed successfully.", extensionName)
				if !strings.Contains(output, exp) {
					t.Errorf("expected output to contain %q, got:\n%s", exp, output)
				}
			} else {
				exp := fmt.Sprintf("✅ Extension %q removed, but its directory is preserved.", extensionName)
				if !strings.Contains(output, exp) {
					t.Errorf("expected output to contain %q, got:\n%s", exp, output)
				}
			}

			// Check if directory was deleted or not
			_, err = os.Stat(extensionDir)
			if tt.expectDeleted {
				if !os.IsNotExist(err) {
					t.Errorf("expected extension directory to be deleted, but it still exists")
				}
			} else {
				if err != nil {
					t.Errorf("expected extension directory to exist, got: %v", err)
				}
			}

			// Verify extension is disabled in config
			cfg, err = config.LoadConfigFn()
			if err != nil {
				t.Fatalf("failed to reload config: %v", err)
			}
			found := false
			for _, ext := range cfg.Extensions {
				if ext.Name == extensionName {
					found = true
					if ext.Enabled {
						t.Errorf("expected extension %q to be disabled, but it's still enabled", extensionName)
					}
					break
				}
			}
			if !found {
				t.Errorf("extension %q not found in config", extensionName)
			}
		})
	}
}

func TestExtensionRemoveCmd_DeleteFolderFailure(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("permission-based RemoveAll failure is unreliable on Windows")
	}

	tmpDir := t.TempDir()
	extensionName := "mock-extension"
	extensionDir := filepath.Join(tmpDir, ".semver-extensions", extensionName)
	extensionConfigPath := filepath.Join(tmpDir, ".semver.yaml")

	if err := os.MkdirAll(extensionDir, 0755); err != nil {
		t.Fatalf("failed to create extension directory: %v", err)
	}

	protectedFile := filepath.Join(extensionDir, "protected.txt")
	if err := os.WriteFile(protectedFile, []byte("protected"), 0400); err != nil {
		t.Fatalf("failed to create protected file: %v", err)
	}
	if err := os.Chmod(extensionDir, 0500); err != nil {
		t.Fatalf("failed to chmod extension dir: %v", err)
	}

	defer func() {
		_ = os.Chmod(protectedFile, 0600)
		_ = os.Chmod(extensionDir, 0700)
	}()

	content := `extensions:
  - name: mock-extension
    path: /some/path
    enabled: true`
	if err := os.WriteFile(extensionConfigPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	// Prepare and run the CLI command
	cfg := &config.Config{Path: extensionConfigPath}
	appCli := newCLI(cfg)

	var cliErr error
	_, captureErr := testutils.CaptureStdout(func() {
		cliErr = testutils.RunCLITestAllowError(t, appCli, []string{
			"semver", "extension", "remove",
			"--name", extensionName,
			"--delete-folder",
		}, tmpDir)
	})
	if captureErr != nil {
		t.Fatalf("failed to capture output: %v", captureErr)
	}

	_ = os.Chmod(protectedFile, 0600)
	_ = os.Chmod(extensionDir, 0700)

	if cliErr == nil {
		t.Fatal("expected error but got nil")
	}

	expectedMsg := "failed to remove extension directory"
	if !strings.Contains(cliErr.Error(), expectedMsg) {
		t.Errorf("expected error message to contain %q, got: %v", expectedMsg, cliErr)
	}

}

func TestCLI_ExtensionRemove_MissingName(t *testing.T) {
	if os.Getenv("TEST_EXTENSION_REMOVE_MISSING_NAME") == "1" {
		tmp := t.TempDir()

		// Write valid .semver.yaml with 1 extension (won't be used, but still required)
		configPath := filepath.Join(tmp, ".semver.yaml")
		content := `extensions:
  - name: mock-extension
    path: /some/path
    enabled: true`
		if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
			fmt.Fprintln(os.Stderr, "failed to write config:", err)
			os.Exit(1)
		}

		// Prepare and run the CLI command
		cfg := &config.Config{Path: configPath}
		appCli := newCLI(cfg)

		// Run command WITHOUT --name (should trigger the validation)
		err := appCli.Run(context.Background(), []string{
			"semver", "extension", "remove", "--path", configPath,
		})
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		// Shouldn't reach here
		os.Exit(0)
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestCLI_ExtensionRemove_MissingName")
	cmd.Env = append(os.Environ(), "TEST_EXTENSION_REMOVE_MISSING_NAME=1")
	output, err := cmd.CombinedOutput()

	if err == nil {
		t.Fatal("expected non-zero exit status")
	}

	expected := "please provide an extension name to remove"
	if !strings.Contains(string(output), expected) {
		t.Errorf("expected output to contain %q, got:\n%s", expected, output)
	}
}

func TestExtensionRemoveCmd_LoadConfigError(t *testing.T) {
	// Mock the LoadConfig function to simulate an error
	originalLoadConfig := config.LoadConfigFn
	defer func() {
		config.LoadConfigFn = originalLoadConfig
	}()

	config.LoadConfigFn = func() (*config.Config, error) {
		return nil, fmt.Errorf("failed to load configuration")
	}

	tmpDir := t.TempDir()
	extensionConfigPath := filepath.Join(tmpDir, ".semver.yaml")

	// Prepare and run the CLI command
	cfg := &config.Config{Path: extensionConfigPath}
	appCli := newCLI(cfg)

	// Capture the output
	output, err := testutils.CaptureStdout(func() {
		testutils.RunCLITest(t, appCli, []string{"semver", "extension", "remove", "--name", "mock-plugin"}, tmpDir)
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

func TestExtensionRemoveCmd_PluginNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	extensionConfigPath := filepath.Join(tmpDir, ".semver.yaml")

	// Create a dummy .semver.yaml configuration file with no extension
	content := `extensions: []`
	if err := os.WriteFile(extensionConfigPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create .semver.yaml: %v", err)
	}

	// Prepare and run the CLI command
	cfg := &config.Config{Path: extensionConfigPath}
	appCli := newCLI(cfg)

	// Capture the output
	output, err := testutils.CaptureStdout(func() {
		testutils.RunCLITest(t, appCli, []string{"semver", "extension", "remove", "--name", "mock-extension"}, tmpDir)
	})

	if err != nil {
		t.Fatalf("Failed to capture stdout: %v", err)
	}

	// Expected error message
	expected := "extension \"mock-extension\" not found"
	if !strings.Contains(output, expected) {
		t.Errorf("expected output to contain %q, got %q", expected, output)
	}
}

func TestExtensionRemoveCmd_SaveConfigError(t *testing.T) {
	// Mock the SaveConfig function to simulate an error
	originalSaveConfig := config.SaveConfigFn
	defer func() {
		config.SaveConfigFn = originalSaveConfig
	}()

	config.SaveConfigFn = func(cfg *config.Config) error {
		return fmt.Errorf("failed to save updated configuration")
	}

	tmpDir := t.TempDir()
	extensionConfigPath := filepath.Join(tmpDir, ".semver.yaml")

	// Create a dummy .semver.yaml configuration file
	content := `extensions:
  - name: mock-extension
    path: /path/to/extension
    enabled: true`
	if err := os.WriteFile(extensionConfigPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create .semver.yaml: %v", err)
	}

	// Prepare and run the CLI command
	cfg := &config.Config{Path: extensionConfigPath}
	appCli := newCLI(cfg)

	// Capture the output
	output, err := testutils.CaptureStdout(func() {
		testutils.RunCLITest(t, appCli, []string{"semver", "extension", "remove", "--name", "mock-extension"}, tmpDir)
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
