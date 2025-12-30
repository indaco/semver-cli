//go:build integration

package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// Integration tests for the semver CLI.
// Run with: go test -tags=integration ./cmd/semver/...
//
// These tests build the actual binary and run it against real files,
// providing end-to-end verification of CLI behavior.

var binaryPath string

func TestMain(m *testing.M) {
	// Get the directory containing this test file
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		panic("failed to get current file path")
	}
	testDir := filepath.Dir(filename)

	// Build the binary once for all tests
	tmpDir, err := os.MkdirTemp("", "semver-integration-*")
	if err != nil {
		panic("failed to create temp dir: " + err.Error())
	}

	binaryPath = filepath.Join(tmpDir, "semver")
	cmd := exec.Command("go", "build", "-o", binaryPath, ".")
	cmd.Dir = testDir
	if output, err := cmd.CombinedOutput(); err != nil {
		panic("failed to build binary: " + string(output))
	}

	code := m.Run()

	// Cleanup
	os.RemoveAll(tmpDir)
	os.Exit(code)
}

// runSemver executes the semver binary with the given arguments.
func runSemver(t *testing.T, workdir string, args ...string) (string, error) {
	t.Helper()
	cmd := exec.Command(binaryPath, args...)
	cmd.Dir = workdir
	output, err := cmd.CombinedOutput()
	return strings.TrimSpace(string(output)), err
}

// writeVersionFile creates a .version file with the given content.
func writeVersionFile(t *testing.T, dir, version string) string {
	t.Helper()
	path := filepath.Join(dir, ".version")
	if err := os.WriteFile(path, []byte(version+"\n"), 0644); err != nil {
		t.Fatalf("failed to write version file: %v", err)
	}
	return path
}

// readVersionFile reads the content of the .version file.
func readVersionFile(t *testing.T, dir string) string {
	t.Helper()
	path := filepath.Join(dir, ".version")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read version file: %v", err)
	}
	return strings.TrimSpace(string(data))
}

// TestIntegration_Init tests the init command.
func TestIntegration_Init(t *testing.T) {
	t.Run("creates default version file", func(t *testing.T) {
		dir := t.TempDir()

		output, err := runSemver(t, dir, "init")
		if err != nil {
			t.Fatalf("init failed: %v, output: %s", err, output)
		}

		version := readVersionFile(t, dir)
		if version != "0.1.0" {
			t.Errorf("expected version 0.1.0, got %s", version)
		}
	})

	t.Run("creates version file at custom path", func(t *testing.T) {
		dir := t.TempDir()
		customPath := filepath.Join(dir, "custom", ".version")
		if err := os.MkdirAll(filepath.Dir(customPath), 0755); err != nil {
			t.Fatal(err)
		}

		output, err := runSemver(t, dir, "init", "--path", customPath)
		if err != nil {
			t.Fatalf("init failed: %v, output: %s", err, output)
		}

		data, err := os.ReadFile(customPath)
		if err != nil {
			t.Fatalf("failed to read custom version file: %v", err)
		}
		if strings.TrimSpace(string(data)) != "0.1.0" {
			t.Errorf("expected version 0.1.0, got %s", string(data))
		}
	})
}

// TestIntegration_Show tests the show command.
func TestIntegration_Show(t *testing.T) {
	t.Run("displays current version", func(t *testing.T) {
		dir := t.TempDir()
		writeVersionFile(t, dir, "1.2.3")

		output, err := runSemver(t, dir, "show")
		if err != nil {
			t.Fatalf("show failed: %v, output: %s", err, output)
		}

		if output != "1.2.3" {
			t.Errorf("expected output '1.2.3', got %q", output)
		}
	})

	t.Run("displays version with pre-release", func(t *testing.T) {
		dir := t.TempDir()
		writeVersionFile(t, dir, "2.0.0-beta.1")

		output, err := runSemver(t, dir, "show")
		if err != nil {
			t.Fatalf("show failed: %v, output: %s", err, output)
		}

		if output != "2.0.0-beta.1" {
			t.Errorf("expected output '2.0.0-beta.1', got %q", output)
		}
	})

	t.Run("strict mode fails when file missing", func(t *testing.T) {
		dir := t.TempDir()

		_, err := runSemver(t, dir, "show", "--strict")
		if err == nil {
			t.Error("expected error with --strict and missing file")
		}
	})

	t.Run("auto-init when file missing", func(t *testing.T) {
		dir := t.TempDir()

		output, err := runSemver(t, dir, "show")
		if err != nil {
			t.Fatalf("show failed: %v, output: %s", err, output)
		}

		// Should auto-initialize and show the version
		if !strings.Contains(output, "0.1.0") {
			t.Errorf("expected output to contain '0.1.0', got %q", output)
		}
	})
}

// TestIntegration_Set tests the set command.
func TestIntegration_Set(t *testing.T) {
	t.Run("sets version", func(t *testing.T) {
		dir := t.TempDir()
		writeVersionFile(t, dir, "1.0.0")

		output, err := runSemver(t, dir, "set", "2.0.0")
		if err != nil {
			t.Fatalf("set failed: %v, output: %s", err, output)
		}

		version := readVersionFile(t, dir)
		if version != "2.0.0" {
			t.Errorf("expected version 2.0.0, got %s", version)
		}
	})

	t.Run("sets version with pre-release", func(t *testing.T) {
		dir := t.TempDir()
		writeVersionFile(t, dir, "1.0.0")

		_, err := runSemver(t, dir, "set", "2.0.0", "--pre", "alpha.1")
		if err != nil {
			t.Fatalf("set failed: %v", err)
		}

		version := readVersionFile(t, dir)
		if version != "2.0.0-alpha.1" {
			t.Errorf("expected version 2.0.0-alpha.1, got %s", version)
		}
	})

	t.Run("sets version with build metadata", func(t *testing.T) {
		dir := t.TempDir()
		writeVersionFile(t, dir, "1.0.0")

		_, err := runSemver(t, dir, "set", "2.0.0", "--meta", "build.123")
		if err != nil {
			t.Fatalf("set failed: %v", err)
		}

		version := readVersionFile(t, dir)
		if version != "2.0.0+build.123" {
			t.Errorf("expected version 2.0.0+build.123, got %s", version)
		}
	})

	t.Run("sets version with pre-release and metadata", func(t *testing.T) {
		dir := t.TempDir()
		writeVersionFile(t, dir, "1.0.0")

		_, err := runSemver(t, dir, "set", "2.0.0", "--pre", "rc.1", "--meta", "ci.456")
		if err != nil {
			t.Fatalf("set failed: %v", err)
		}

		version := readVersionFile(t, dir)
		if version != "2.0.0-rc.1+ci.456" {
			t.Errorf("expected version 2.0.0-rc.1+ci.456, got %s", version)
		}
	})
}

// TestIntegration_BumpPatch tests bump patch command.
func TestIntegration_BumpPatch(t *testing.T) {
	tests := []struct {
		name     string
		initial  string
		expected string
	}{
		{"simple bump", "1.2.3", "1.2.4"},
		{"from zero", "0.0.0", "0.0.1"},
		{"large numbers", "10.20.30", "10.20.31"},
		{"clears pre-release", "1.2.3-alpha", "1.2.4"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			writeVersionFile(t, dir, tt.initial)

			_, err := runSemver(t, dir, "bump", "patch")
			if err != nil {
				t.Fatalf("bump patch failed: %v", err)
			}

			version := readVersionFile(t, dir)
			if version != tt.expected {
				t.Errorf("expected version %s, got %s", tt.expected, version)
			}
		})
	}
}

// TestIntegration_BumpMinor tests bump minor command.
func TestIntegration_BumpMinor(t *testing.T) {
	tests := []struct {
		name     string
		initial  string
		expected string
	}{
		{"simple bump", "1.2.3", "1.3.0"},
		{"resets patch", "1.2.9", "1.3.0"},
		{"from zero", "0.0.5", "0.1.0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			writeVersionFile(t, dir, tt.initial)

			_, err := runSemver(t, dir, "bump", "minor")
			if err != nil {
				t.Fatalf("bump minor failed: %v", err)
			}

			version := readVersionFile(t, dir)
			if version != tt.expected {
				t.Errorf("expected version %s, got %s", tt.expected, version)
			}
		})
	}
}

// TestIntegration_BumpMajor tests bump major command.
func TestIntegration_BumpMajor(t *testing.T) {
	tests := []struct {
		name     string
		initial  string
		expected string
	}{
		{"simple bump", "1.2.3", "2.0.0"},
		{"resets minor and patch", "1.9.9", "2.0.0"},
		{"from zero", "0.5.5", "1.0.0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			writeVersionFile(t, dir, tt.initial)

			_, err := runSemver(t, dir, "bump", "major")
			if err != nil {
				t.Fatalf("bump major failed: %v", err)
			}

			version := readVersionFile(t, dir)
			if version != tt.expected {
				t.Errorf("expected version %s, got %s", tt.expected, version)
			}
		})
	}
}

// TestIntegration_BumpRelease tests bump release command.
func TestIntegration_BumpRelease(t *testing.T) {
	tests := []struct {
		name     string
		initial  string
		expected string
	}{
		{"removes pre-release", "1.2.3-alpha.1", "1.2.3"},
		{"removes pre-release and metadata", "1.2.3-beta+build.123", "1.2.3"},
		{"no-op for release version", "1.2.3", "1.2.3"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			writeVersionFile(t, dir, tt.initial)

			_, err := runSemver(t, dir, "bump", "release")
			if err != nil {
				t.Fatalf("bump release failed: %v", err)
			}

			version := readVersionFile(t, dir)
			if version != tt.expected {
				t.Errorf("expected version %s, got %s", tt.expected, version)
			}
		})
	}
}

// TestIntegration_BumpAuto tests bump auto command.
func TestIntegration_BumpAuto(t *testing.T) {
	tests := []struct {
		name     string
		initial  string
		args     []string
		expected string
	}{
		{"promotes pre-release", "1.2.3-alpha.1", []string{"bump", "auto"}, "1.2.3"},
		{"bumps patch for release", "1.2.3", []string{"bump", "auto", "--no-infer"}, "1.2.4"},
		{"with label minor", "1.2.3", []string{"bump", "auto", "--label", "minor"}, "1.3.0"},
		{"with label major", "1.2.3", []string{"bump", "auto", "--label", "major"}, "2.0.0"},
		{"preserves metadata", "1.2.3-alpha+build.1", []string{"bump", "auto", "--preserve-meta"}, "1.2.3+build.1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			writeVersionFile(t, dir, tt.initial)

			_, err := runSemver(t, dir, tt.args...)
			if err != nil {
				t.Fatalf("bump auto failed: %v", err)
			}

			version := readVersionFile(t, dir)
			if version != tt.expected {
				t.Errorf("expected version %s, got %s", tt.expected, version)
			}
		})
	}

	t.Run("next alias works", func(t *testing.T) {
		dir := t.TempDir()
		writeVersionFile(t, dir, "1.2.3-alpha")

		_, err := runSemver(t, dir, "bump", "next")
		if err != nil {
			t.Fatalf("bump next (alias) failed: %v", err)
		}

		version := readVersionFile(t, dir)
		if version != "1.2.3" {
			t.Errorf("expected version 1.2.3, got %s", version)
		}
	})
}

// TestIntegration_Pre tests the pre command.
func TestIntegration_Pre(t *testing.T) {
	tests := []struct {
		name     string
		initial  string
		args     []string
		expected string
	}{
		{"sets alpha", "1.2.3", []string{"pre", "--label", "alpha"}, "1.2.4-alpha"},
		{"sets beta", "1.2.3", []string{"pre", "--label", "beta"}, "1.2.4-beta"},
		{"replaces pre-release", "1.2.3-alpha", []string{"pre", "--label", "beta"}, "1.2.3-beta"},
		{"increments alpha", "1.2.3", []string{"pre", "--label", "alpha", "--inc"}, "1.2.3-alpha.1"},
		{"increments existing", "1.2.3-alpha.1", []string{"pre", "--label", "alpha", "--inc"}, "1.2.3-alpha.2"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			writeVersionFile(t, dir, tt.initial)

			_, err := runSemver(t, dir, tt.args...)
			if err != nil {
				t.Fatalf("pre failed: %v", err)
			}

			version := readVersionFile(t, dir)
			if version != tt.expected {
				t.Errorf("expected version %s, got %s", tt.expected, version)
			}
		})
	}
}

// TestIntegration_Validate tests the validate command.
func TestIntegration_Validate(t *testing.T) {
	t.Run("valid version passes", func(t *testing.T) {
		dir := t.TempDir()
		writeVersionFile(t, dir, "1.2.3")

		output, err := runSemver(t, dir, "validate")
		if err != nil {
			t.Fatalf("validate failed: %v, output: %s", err, output)
		}

		if !strings.Contains(output, "Valid") {
			t.Errorf("expected output to contain 'Valid', got %q", output)
		}
	})

	t.Run("invalid version fails", func(t *testing.T) {
		dir := t.TempDir()
		writeVersionFile(t, dir, "not-a-version")

		_, err := runSemver(t, dir, "validate")
		if err == nil {
			t.Error("expected error for invalid version")
		}
	})

	t.Run("missing file fails", func(t *testing.T) {
		dir := t.TempDir()

		_, err := runSemver(t, dir, "validate", "--strict")
		if err == nil {
			t.Error("expected error for missing file with --strict")
		}
	})
}

// TestIntegration_BumpWithFlags tests bump commands with various flags.
func TestIntegration_BumpWithFlags(t *testing.T) {
	t.Run("bump patch with pre-release", func(t *testing.T) {
		dir := t.TempDir()
		writeVersionFile(t, dir, "1.2.3")

		_, err := runSemver(t, dir, "bump", "patch", "--pre", "beta.1")
		if err != nil {
			t.Fatalf("bump failed: %v", err)
		}

		version := readVersionFile(t, dir)
		if version != "1.2.4-beta.1" {
			t.Errorf("expected version 1.2.4-beta.1, got %s", version)
		}
	})

	t.Run("bump minor with metadata", func(t *testing.T) {
		dir := t.TempDir()
		writeVersionFile(t, dir, "1.2.3")

		_, err := runSemver(t, dir, "bump", "minor", "--meta", "ci.456")
		if err != nil {
			t.Fatalf("bump failed: %v", err)
		}

		version := readVersionFile(t, dir)
		if version != "1.3.0+ci.456" {
			t.Errorf("expected version 1.3.0+ci.456, got %s", version)
		}
	})

	t.Run("bump preserves metadata", func(t *testing.T) {
		dir := t.TempDir()
		writeVersionFile(t, dir, "1.2.3+build.789")

		_, err := runSemver(t, dir, "bump", "patch", "--preserve-meta")
		if err != nil {
			t.Fatalf("bump failed: %v", err)
		}

		version := readVersionFile(t, dir)
		if version != "1.2.4+build.789" {
			t.Errorf("expected version 1.2.4+build.789, got %s", version)
		}
	})
}

// TestIntegration_CustomPath tests --path flag across commands.
func TestIntegration_CustomPath(t *testing.T) {
	dir := t.TempDir()
	customPath := filepath.Join(dir, "versions", "app.version")
	if err := os.MkdirAll(filepath.Dir(customPath), 0755); err != nil {
		t.Fatal(err)
	}

	// Init at custom path
	_, err := runSemver(t, dir, "init", "--path", customPath)
	if err != nil {
		t.Fatalf("init failed: %v", err)
	}

	// Show from custom path
	output, err := runSemver(t, dir, "show", "--path", customPath)
	if err != nil {
		t.Fatalf("show failed: %v", err)
	}
	if output != "0.1.0" {
		t.Errorf("expected 0.1.0, got %s", output)
	}

	// Bump at custom path
	_, err = runSemver(t, dir, "bump", "minor", "--path", customPath)
	if err != nil {
		t.Fatalf("bump failed: %v", err)
	}

	// Verify bump
	output, err = runSemver(t, dir, "show", "--path", customPath)
	if err != nil {
		t.Fatalf("show failed: %v", err)
	}
	if output != "0.2.0" {
		t.Errorf("expected 0.2.0, got %s", output)
	}
}

// TestIntegration_VersionFlag tests --version flag.
func TestIntegration_VersionFlag(t *testing.T) {
	dir := t.TempDir()

	output, err := runSemver(t, dir, "--version")
	if err != nil {
		t.Fatalf("--version failed: %v", err)
	}

	if !strings.HasPrefix(output, "semver version v") {
		t.Errorf("expected version output, got %q", output)
	}
}

// TestIntegration_HelpFlag tests --help flag.
func TestIntegration_HelpFlag(t *testing.T) {
	dir := t.TempDir()

	output, err := runSemver(t, dir, "--help")
	if err != nil {
		t.Fatalf("--help failed: %v", err)
	}

	requiredStrings := []string{"semver", "show", "set", "bump", "pre", "validate", "init"}
	for _, s := range requiredStrings {
		if !strings.Contains(output, s) {
			t.Errorf("expected help to contain %q", s)
		}
	}
}
