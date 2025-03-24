package semver

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

/* ------------------------------------------------------------------------- */
/* VERSION PARSING                                                           */
/* ------------------------------------------------------------------------- */

func TestParseAndString(t *testing.T) {
	tests := []struct {
		raw      string
		expected string
	}{
		{"1.2.3", "1.2.3"},
		{"1.2.3-alpha", "1.2.3-alpha"},
		{"  1.2.3-beta.1 ", "1.2.3-beta.1"},
	}

	for _, tt := range tests {
		v, err := parseVersion(tt.raw)
		if err != nil {
			t.Errorf("parseVersion(%q) failed: %v", tt.raw, err)
			continue
		}
		if v.String() != tt.expected {
			t.Errorf("expected %q, got %q", tt.expected, v.String())
		}
	}
}

func TestParseVersion_ErrorCases(t *testing.T) {
	t.Run("invalid format (missing patch)", func(t *testing.T) {
		_, err := parseVersion("1.2")
		if err == nil || !errors.Is(err, errInvalidVersion) {
			t.Errorf("expected ErrInvalidVersion, got %v", err)
		}
	})

	t.Run("non-numeric major", func(t *testing.T) {
		_, err := parseVersion("a.2.3")
		if err == nil || !errors.Is(err, errInvalidVersion) {
			t.Errorf("expected ErrInvalidVersion, got %v", err)
		}
	})

	t.Run("non-numeric minor", func(t *testing.T) {
		_, err := parseVersion("1.b.3")
		if err == nil || !errors.Is(err, errInvalidVersion) {
			t.Errorf("expected ErrInvalidVersion, got %v", err)
		}
	})

	t.Run("non-numeric patch", func(t *testing.T) {
		_, err := parseVersion("1.2.c")
		if err == nil || !errors.Is(err, errInvalidVersion) {
			t.Errorf("expected ErrInvalidVersion, got %v", err)
		}
	})
}

func TestParseVersion_InvalidFormat(t *testing.T) {
	invalidVersions := []string{
		"",
		"1",
		"1.2",
		"abc.def.ghi",
		"v1.2.3",
		"1.2.3.4",
	}

	for _, raw := range invalidVersions {
		_, err := parseVersion(raw)
		if err == nil {
			t.Errorf("expected error for invalid version %q, got nil", raw)
		}
	}
}

/* ------------------------------------------------------------------------- */
/* VERSION UPDATES                                                           */
/* ------------------------------------------------------------------------- */

func TestUpdateVersionWithPreRelease(t *testing.T) {
	path := writeTempVersion(t, "1.2.3-alpha")
	defer os.Remove(path)

	if err := UpdateVersion(path, "minor"); err != nil {
		t.Fatal(err)
	}
	got := strings.TrimSpace(readFile(t, path))
	expected := "1.3.0"
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestUpdateVersion_Patch(t *testing.T) {
	path := writeTempVersion(t, "1.2.3-beta")
	defer os.Remove(path)

	err := UpdateVersion(path, "patch")
	if err != nil {
		t.Fatal(err)
	}

	got := strings.TrimSpace(readFile(t, path))
	expected := "1.2.4"
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestUpdateVersion_Major(t *testing.T) {
	path := writeTempVersion(t, "1.2.3-beta.1")
	defer os.Remove(path)

	err := UpdateVersion(path, "major")
	if err != nil {
		t.Fatal(err)
	}

	got := strings.TrimSpace(readFile(t, path))
	expected := "2.0.0"
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestUpdateVersion_UnknownLevel(t *testing.T) {
	path := writeTempVersion(t, "1.2.3")
	defer os.Remove(path)

	err := UpdateVersion(path, "invalid")
	if err == nil {
		t.Fatal("expected error for unknown level, got nil")
	}

	if !strings.Contains(err.Error(), "unknown level") {
		t.Errorf("expected 'unknown level' error, got %v", err)
	}
}

func TestUpdateVersion_InvalidVersionFile(t *testing.T) {
	path := writeTempVersion(t, "not-a-version")
	defer os.Remove(path)

	err := UpdateVersion(path, "patch")
	if err == nil {
		t.Fatal("expected error due to invalid version, got nil")
	}

	if !strings.Contains(err.Error(), "invalid version format") {
		t.Errorf("unexpected error: %v", err)
	}
}

/* ------------------------------------------------------------------------- */
/* VERSION READ/WRITE                                                        */
/* ------------------------------------------------------------------------- */

func TestReadVersion_FileDoesNotExist(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nonexistent.version")

	_, err := ReadVersion(path)
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}

	if !os.IsNotExist(err) {
		t.Errorf("expected file-not-found error, got %v", err)
	}
}

func TestSetPreRelease(t *testing.T) {
	path := writeTempVersion(t, "1.2.3")
	defer os.Remove(path)

	version, _ := ReadVersion(path)
	version.PreRelease = "rc.1"
	if err := SaveVersion(path, version); err != nil {
		t.Fatal(err)
	}

	got := strings.TrimSpace(readFile(t, path))
	want := "1.2.3-rc.1"
	if got != want {
		t.Errorf("expected %q, got %q", want, got)
	}
}

/* ------------------------------------------------------------------------- */
/* PRE-RELEASE INCREMENTS                                                    */
/* ------------------------------------------------------------------------- */

func TestIncrementPreRelease(t *testing.T) {
	cases := []struct {
		current string
		base    string
		want    string
	}{
		{"alpha", "alpha", "alpha.1"},
		{"alpha.", "alpha", "alpha.1"},
		{"alpha.1", "alpha", "alpha.2"},
		{"alpha.9", "alpha", "alpha.10"},
		{"beta", "alpha", "alpha.1"},
		{"", "rc", "rc.1"},
	}

	for _, c := range cases {
		got := IncrementPreRelease(c.current, c.base)
		if got != c.want {
			t.Errorf("incrementPreRelease(%q, %q) = %q, want %q", c.current, c.base, got, c.want)
		}
	}
}

/* ------------------------------------------------------------------------- */
/* VERSION FILE INITIALIZATION                                               */
/* ------------------------------------------------------------------------- */

func TestInitializeVersionFile_NewFile_WithValidGitTag(t *testing.T) {
	tmpDir := t.TempDir()
	versionPath := filepath.Join(tmpDir, ".version")

	execCommand = fakeExecCommand("v1.2.3\n")
	defer func() { execCommand = originalExecCommand }()

	err := InitializeVersionFile(versionPath)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	data, _ := os.ReadFile(versionPath)
	got := strings.TrimSpace(string(data))
	want := "1.2.3"

	if got != want {
		t.Errorf("expected %q, got %q", want, got)
	}
}

func TestInitializeVersionFile_ExistingFile(t *testing.T) {
	tmpDir := t.TempDir()
	versionPath := filepath.Join(tmpDir, ".version")

	err := os.WriteFile(versionPath, []byte("1.2.3\n"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	execCommand = fakeExecCommand("v9.9.9\n")
	defer func() { execCommand = exec.Command }()

	err = InitializeVersionFile(versionPath)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	data, _ := os.ReadFile(versionPath)
	got := strings.TrimSpace(string(data))
	if got != "1.2.3" {
		t.Errorf("expected file content to remain '1.2.3', got %q", got)
	}
}

func TestInitializeVersionFile_InvalidGitTagFormat(t *testing.T) {
	tmpDir := t.TempDir()
	versionPath := filepath.Join(tmpDir, ".version")

	execCommand = fakeExecCommand("invalid-tag\n")
	defer func() { execCommand = exec.Command }()

	err := InitializeVersionFile(versionPath)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	data, _ := os.ReadFile(versionPath)
	got := strings.TrimSpace(string(data))
	want := "0.1.0"

	if got != want {
		t.Errorf("expected fallback version %q, got %q", want, got)
	}
}

/* ------------------------------------------------------------------------- */
/* HELPERS                                                                   */
/* ------------------------------------------------------------------------- */

func writeTempVersion(t *testing.T, content string) string {
	t.Helper()
	tmpFile, err := os.CreateTemp("", ".version")
	if err != nil {
		t.Fatal(err)
	}
	defer tmpFile.Close()

	if _, err := tmpFile.WriteString(content); err != nil {
		t.Fatal(err)
	}
	return tmpFile.Name()
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

/* ------------------------------------------------------------------------- */
/* GIT MOCKING                                                               */
/* ------------------------------------------------------------------------- */

var originalExecCommand = execCommand

func fakeExecCommand(output string) func(string, ...string) *exec.Cmd {
	return func(command string, args ...string) *exec.Cmd {
		cs := []string{"-test.run=TestHelperProcess", "--", command}
		cs = append(cs, args...)
		cmd := exec.Command(os.Args[0], cs...)
		cmd.Env = append(os.Environ(),
			"GO_WANT_HELPER_PROCESS=1",
			"MOCK_OUTPUT="+output,
		)
		return cmd
	}
}

// Fake subprocess entrypoint
func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	t.Log(">>> MOCK HELPER RUNNING <<<")
	_, _ = os.Stdout.WriteString(os.Getenv("MOCK_OUTPUT"))
	os.Exit(0)
}
