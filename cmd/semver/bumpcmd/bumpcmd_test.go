package bumpcmd

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
	"github.com/indaco/semver-cli/internal/hooks"
	"github.com/indaco/semver-cli/internal/plugins/commitparser"
	"github.com/indaco/semver-cli/internal/plugins/commitparser/gitlog"
	"github.com/indaco/semver-cli/internal/semver"
	"github.com/indaco/semver-cli/internal/testutils"
	"github.com/urfave/cli/v3"
)

func TestCLI_BumpCommand_Variants(t *testing.T) {
	tmpDir := t.TempDir()
	versionPath := filepath.Join(tmpDir, ".version")

	// Prepare the CLI command
	cfg := &config.Config{Path: versionPath}
	appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run(cfg)})

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
			appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run(cfg)})
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

func TestCLI_BumpSubcommands_EarlyFailures(t *testing.T) {

	tests := []struct {
		name        string
		args        []string
		override    func() func() // returns restore function
		expectedErr string
	}{
		{
			name: "patch - FromCommand fails",
			args: []string{"semver", "bump", "patch"},
			override: func() func() {
				original := clix.FromCommandFn
				clix.FromCommandFn = func(cmd *cli.Command) (bool, error) {
					return false, fmt.Errorf("mock FromCommand error")
				}
				return func() { clix.FromCommandFn = original }
			},
			expectedErr: "mock FromCommand error",
		},
		{
			name: "patch - RunPreReleaseHooks fails",
			args: []string{"semver", "bump", "patch"},
			override: func() func() {
				original := hooks.RunPreReleaseHooksFn
				hooks.RunPreReleaseHooksFn = func(skip bool) error {
					return fmt.Errorf("mock pre-release hooks error")
				}
				return func() { hooks.RunPreReleaseHooksFn = original }
			},
			expectedErr: "mock pre-release hooks error",
		},
		{
			name: "minor - FromCommand fails",
			args: []string{"semver", "bump", "minor"},
			override: func() func() {
				original := clix.FromCommandFn
				clix.FromCommandFn = func(cmd *cli.Command) (bool, error) {
					return false, fmt.Errorf("mock FromCommand error")
				}
				return func() { clix.FromCommandFn = original }
			},
			expectedErr: "mock FromCommand error",
		},
		{
			name: "minor - RunPreReleaseHooks fails",
			args: []string{"semver", "bump", "minor"},
			override: func() func() {
				original := hooks.RunPreReleaseHooksFn
				hooks.RunPreReleaseHooksFn = func(skip bool) error {
					return fmt.Errorf("mock pre-release hooks error")
				}
				return func() { hooks.RunPreReleaseHooksFn = original }
			},
			expectedErr: "mock pre-release hooks error",
		},
		{
			name: "major - FromCommand fails",
			args: []string{"semver", "bump", "major"},
			override: func() func() {
				original := clix.FromCommandFn
				clix.FromCommandFn = func(cmd *cli.Command) (bool, error) {
					return false, fmt.Errorf("mock FromCommand error")
				}
				return func() { clix.FromCommandFn = original }
			},
			expectedErr: "mock FromCommand error",
		},
		{
			name: "major - RunPreReleaseHooks fails",
			args: []string{"semver", "bump", "major"},
			override: func() func() {
				original := hooks.RunPreReleaseHooksFn
				hooks.RunPreReleaseHooksFn = func(skip bool) error {
					return fmt.Errorf("mock pre-release hooks error")
				}
				return func() { hooks.RunPreReleaseHooksFn = original }
			},
			expectedErr: "mock pre-release hooks error",
		},
		{
			name: "auto - FromCommand fails",
			args: []string{"semver", "bump", "auto"},
			override: func() func() {
				original := clix.FromCommandFn
				clix.FromCommandFn = func(cmd *cli.Command) (bool, error) {
					return false, fmt.Errorf("mock FromCommand error")
				}
				return func() { clix.FromCommandFn = original }
			},
			expectedErr: "mock FromCommand error",
		},
		{
			name: "auto - RunPreReleaseHooks fails",
			args: []string{"semver", "bump", "auto"},
			override: func() func() {
				original := hooks.RunPreReleaseHooksFn
				hooks.RunPreReleaseHooksFn = func(skip bool) error {
					return fmt.Errorf("mock pre-release hooks error")
				}
				return func() { hooks.RunPreReleaseHooksFn = original }
			},
			expectedErr: "mock pre-release hooks error",
		},
		{
			name: "release - FromCommand fails",
			args: []string{"semver", "bump", "release"},
			override: func() func() {
				original := clix.FromCommandFn
				clix.FromCommandFn = func(cmd *cli.Command) (bool, error) {
					return false, fmt.Errorf("mock FromCommand error")
				}
				return func() { clix.FromCommandFn = original }
			},
			expectedErr: "mock FromCommand error",
		},
		{
			name: "release - RunPreReleaseHooks fails",
			args: []string{"semver", "bump", "release"},
			override: func() func() {
				original := hooks.RunPreReleaseHooksFn
				hooks.RunPreReleaseHooksFn = func(skip bool) error {
					return fmt.Errorf("mock pre-release hooks error")
				}
				return func() { hooks.RunPreReleaseHooksFn = original }
			},
			expectedErr: "mock pre-release hooks error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			versionPath := filepath.Join(tmpDir, ".version")
			testutils.WriteTempVersionFile(t, tmpDir, "1.2.3")

			restore := tt.override()
			defer restore()

			cfg := &config.Config{Path: versionPath}
			appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run(cfg)})

			err := appCli.Run(context.Background(), tt.args)
			if err == nil || !strings.Contains(err.Error(), tt.expectedErr) {
				t.Fatalf("expected error containing %q, got: %v", tt.expectedErr, err)
			}
		})
	}
}

func TestCLI_BumpReleaseCmd(t *testing.T) {
	tmpDir := t.TempDir()
	versionPath := filepath.Join(tmpDir, ".version")

	// Prepare the CLI command
	cfg := &config.Config{Path: versionPath}
	appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run(cfg)})

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

			testutils.WriteTempVersionFile(t, tmpDir, tt.initialVersion)
			testutils.RunCLITest(t, appCli, tt.args, tmpDir)

			got := testutils.ReadTempVersionFile(t, tmpDir)
			if got != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, got)
			}
		})
	}
}

func TestCLI_BumpAutoCmd(t *testing.T) {
	tmpDir := t.TempDir()
	versionPath := filepath.Join(tmpDir, ".version")

	// Prepare and run the CLI command
	cfg := &config.Config{Path: versionPath}
	appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run(cfg)})

	tests := []struct {
		name     string
		initial  string
		args     []string
		expected string
	}{
		{
			name:     "promotes alpha to release",
			initial:  "1.2.3-alpha.1",
			args:     []string{"semver", "bump", "auto"},
			expected: "1.2.3",
		},
		{
			name:     "promotes rc to release",
			initial:  "1.2.3-rc.1",
			args:     []string{"semver", "bump", "auto"},
			expected: "1.2.3",
		},
		{
			name:     "default patch bump",
			initial:  "1.2.3",
			args:     []string{"semver", "bump", "auto"},
			expected: "1.2.4",
		},
		{
			name:     "promotes pre-release in 0.x series",
			initial:  "0.9.0-alpha.1",
			args:     []string{"semver", "bump", "auto"},
			expected: "0.9.0",
		},
		{
			name:     "bump minor from 0.9.0 as a special case",
			initial:  "0.9.0",
			args:     []string{"semver", "bump", "auto"},
			expected: "0.10.0",
		},
		{
			name:     "preserve build metadata",
			initial:  "1.2.3-alpha.1+meta.123",
			args:     []string{"semver", "bump", "auto", "--preserve-meta"},
			expected: "1.2.3+meta.123",
		},
		{
			name:     "strip build metadata by default",
			initial:  "1.2.3-alpha.1+meta.123",
			args:     []string{"semver", "bump", "auto"},
			expected: "1.2.3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testutils.WriteTempVersionFile(t, tmpDir, tt.initial)
			testutils.RunCLITest(t, appCli, tt.args, tmpDir)

			got := testutils.ReadTempVersionFile(t, tmpDir)
			if got != tt.expected {
				t.Errorf("expected version %q, got %q", tt.expected, got)
			}
		})
	}
}

func TestCLI_BumpAutoCmd_InferredBump(t *testing.T) {
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
	appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run(cfg)})

	err := appCli.Run(context.Background(), []string{
		"semver", "bump", "auto", "--path", versionPath,
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

func TestCLI_BumpAutoCommand_WithLabelAndMeta(t *testing.T) {
	tmpDir := t.TempDir()
	versionPath := filepath.Join(tmpDir, ".version")

	// Prepare the CLI command
	cfg := &config.Config{Path: versionPath}
	appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run(cfg)})

	tests := []struct {
		name    string
		initial string
		args    []string
		want    string
	}{
		{
			name:    "label=patch",
			initial: "1.2.3",
			args:    []string{"semver", "bump", "auto", "--label", "patch"},
			want:    "1.2.4",
		},
		{
			name:    "label=minor",
			initial: "1.2.3",
			args:    []string{"semver", "bump", "auto", "--label", "minor"},
			want:    "1.3.0",
		},
		{
			name:    "label=major",
			initial: "1.2.3",
			args:    []string{"semver", "bump", "auto", "--label", "major"},
			want:    "2.0.0",
		},
		{
			name:    "label=minor with metadata",
			initial: "1.2.3",
			args:    []string{"semver", "bump", "auto", "--label", "minor", "--meta", "build.42"},
			want:    "1.3.0+build.42",
		},
		{
			name:    "preserve existing metadata",
			initial: "1.2.3+ci.88",
			args:    []string{"semver", "bump", "auto", "--label", "patch", "--preserve-meta"},
			want:    "1.2.4+ci.88",
		},
		{
			name:    "override existing metadata",
			initial: "1.2.3+ci.88",
			args:    []string{"semver", "bump", "auto", "--label", "patch", "--meta", "ci.99"},
			want:    "1.2.4+ci.99",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testutils.WriteTempVersionFile(t, tmpDir, tt.initial)
			testutils.RunCLITest(t, appCli, tt.args, tmpDir)

			got := testutils.ReadTempVersionFile(t, tmpDir)
			if got != tt.want {
				t.Errorf("expected %q, got %q", tt.want, got)
			}
		})
	}
}

func TestCLI_BumpAutoCmd_InferredPromotion(t *testing.T) {
	tmp := t.TempDir()
	versionPath := testutils.WriteTempVersionFile(t, tmp, "1.2.3-beta.1")

	originalInfer := tryInferBumpTypeFromCommitParserPluginFn
	defer func() { tryInferBumpTypeFromCommitParserPluginFn = originalInfer }()

	tryInferBumpTypeFromCommitParserPluginFn = func(since, until string) string {
		return "minor"
	}

	cfg := &config.Config{Path: versionPath, Plugins: &config.PluginConfig{CommitParser: true}}
	appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run(cfg)})

	err := appCli.Run(context.Background(), []string{
		"semver", "bump", "auto", "--path", versionPath,
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

func TestCLI_BumpAutoCmd_PromotePreReleaseWithPreserveMeta(t *testing.T) {
	tmp := t.TempDir()
	versionPath := testutils.WriteTempVersionFile(t, tmp, "1.2.3-beta.2+ci.99")

	// Override tryInferBumpTypeFromCommitParserPlugin
	originalInfer := tryInferBumpTypeFromCommitParserPluginFn
	tryInferBumpTypeFromCommitParserPluginFn = func(since, until string) string {
		return "minor" // Force a non-empty inference so that promotePreRelease is called
	}
	t.Cleanup(func() { tryInferBumpTypeFromCommitParserPluginFn = originalInfer })

	cfg := &config.Config{Path: versionPath}
	appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run(cfg)})

	err := appCli.Run(context.Background(), []string{
		"semver", "bump", "auto", "--path", versionPath, "--preserve-meta",
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

func TestCLI_BumpAutoCmd_InferredBumpFails(t *testing.T) {
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
	appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run(cfg)})

	err := appCli.Run(context.Background(), []string{
		"semver", "bump", "auto", "--path", versionPath,
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
	appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run(cfg)})
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
	appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run(cfg)})
	err := appCli.Run(context.Background(), []string{
		"semver", "bump", "release", "--path", versionPath, "--strict",
	})

	if err == nil {
		t.Fatal("expected error due to save failure, got nil")
	}

	if !strings.Contains(err.Error(), "failed to save version") {
		t.Errorf("expected error message to contain 'failed to save version', got: %v", err)
	}
}

func TestCLI_BumpAutoCmd_Errors(t *testing.T) {
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
			args:        []string{"semver", "bump", "auto"},
			expectedErr: "failed to read version",
		},
		{
			name: "fails if version file is not writable",
			setup: func(dir string) {
				path := filepath.Join(dir, ".version")
				_ = os.WriteFile(path, []byte("1.2.3-alpha\n"), 0444)
				_ = os.Chmod(path, 0444)
			},
			args:          []string{"semver", "bump", "auto"},
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
			appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run(cfg)})

			err := appCli.Run(context.Background(), tt.args)
			if err == nil || !strings.Contains(err.Error(), tt.expectedErr) {
				t.Fatalf("expected error to contain %q, got: %v", tt.expectedErr, err)
			}
		})
	}
}

func TestCLI_BumpAutoCmd_InitVersionFileFails(t *testing.T) {
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
	appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run(cfg)})

	err := appCli.Run(context.Background(), []string{
		"semver", "bump", "auto", "--path", versionPath,
	})
	if err == nil || !strings.Contains(err.Error(), "permission denied") {
		t.Fatalf("expected permission denied error, got: %v", err)
	}
}

func TestCLI_BumpAutoCmd_BumpNextFails(t *testing.T) {
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
	appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run(cfg)})

	err := appCli.Run(context.Background(), []string{
		"semver", "bump", "auto", "--path", versionPath, "--no-infer",
	})

	if err == nil || !strings.Contains(err.Error(), "failed to determine next version") {
		t.Fatalf("expected BumpNext failure, got: %v", err)
	}
}

func TestCLI_BumpAutoCmd_SaveVersionFails(t *testing.T) {
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
	appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run(cfg)})

	err := appCli.Run(context.Background(), []string{
		"semver", "bump", "auto", "--path", versionPath, "--strict",
	})

	if err == nil || !strings.Contains(err.Error(), "failed to save version") {
		t.Fatalf("expected error containing 'failed to save version', got: %v", err)
	}
}

func TestCLI_BumpAutoCommand_InvalidLabel(t *testing.T) {
	if os.Getenv("TEST_SEMVER_BUMP_AUTO_INVALID_LABEL") == "1" {
		tmp := t.TempDir()
		versionPath := testutils.WriteTempVersionFile(t, tmp, "1.2.3")

		// Prepare and run the CLI command
		cfg := &config.Config{Path: versionPath}
		appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run(cfg)})

		err := appCli.Run(context.Background(), []string{
			"semver", "bump", "auto", "--label", "banana", "--path", versionPath,
		})
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		os.Exit(0) // shouldn't happen
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestCLI_BumpAutoCommand_InvalidLabel")
	cmd.Env = append(os.Environ(), "TEST_SEMVER_BUMP_AUTO_INVALID_LABEL=1")
	output, err := cmd.CombinedOutput()

	if err == nil {
		t.Fatal("expected non-zero exit status")
	}

	expected := "invalid --label: must be 'patch', 'minor', or 'major'"
	if !strings.Contains(string(output), expected) {
		t.Errorf("expected output to contain %q, got: %q", expected, string(output))
	}
}

func TestCLI_BumpAutoCmd_BumpByLabelFails(t *testing.T) {
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
	appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run(cfg)})

	err := appCli.Run(context.Background(), []string{
		"semver", "bump", "auto", "--label", "patch", "--path", versionPath,
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
	appCli := testutils.BuildCLIForTests(cfg.Path, []*cli.Command{Run(cfg)})

	err := appCli.Run(context.Background(), []string{
		"semver", "bump", "release", "--path", versionPath,
	})

	if err == nil || !strings.Contains(err.Error(), "permission denied") {
		t.Errorf("expected permission denied error, got: %v", err)
	}
}
