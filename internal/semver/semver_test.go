package semver

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

var originalExecCommand = execCommand

func TestSemVersion_String_WithBuildOnly(t *testing.T) {
	v := SemVersion{
		Major: 1,
		Minor: 0,
		Patch: 0,
		Build: "exp.sha.5114f85",
	}

	got := v.String()
	want := "1.0.0+exp.sha.5114f85"

	if got != want {
		t.Errorf("expected %q, got %q", want, got)
	}
}

/* ------------------------------------------------------------------------- */
/* VERSION FILE INITIALIZATION                                               */
/* ------------------------------------------------------------------------- */

func TestInitializeVersionFileWithFeedback(t *testing.T) {
	t.Run("file already exists and is valid", func(t *testing.T) {
		path := writeTempVersion(t, "2.3.4")

		created, err := InitializeVersionFileWithFeedback(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if created {
			t.Errorf("expected created=false, got true")
		}

	})

	t.Run("file already exists and is invalid", func(t *testing.T) {
		path := writeTempVersion(t, "not-a-version")

		created, err := InitializeVersionFileWithFeedback(path)
		if err != nil {
			t.Fatalf("unexpected error from feedback function: %v", err)
		}
		if created {
			t.Errorf("expected created=false for existing file, got true")
		}

		// Now test the actual parse failure
		_, err = ReadVersion(path)
		if err == nil {
			t.Fatal("expected error from ReadVersion, got nil")
		}
		if !strings.Contains(err.Error(), "invalid version format") {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("file does not exist, fallback to git tag", func(t *testing.T) {
		tmp := t.TempDir()
		path := filepath.Join(tmp, ".version")

		execCommand = fakeExecCommand("v1.2.3\n")
		defer func() { execCommand = originalExecCommand }()

		created, err := InitializeVersionFileWithFeedback(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !created {
			t.Errorf("expected created=true, got false")
		}

	})

	t.Run("file does not exist, fallback to default 0.1.0", func(t *testing.T) {
		tmp := t.TempDir()
		path := filepath.Join(tmp, ".version")

		execCommand = fakeExecCommand("invalid-tag\n")
		defer func() { execCommand = originalExecCommand }()

		created, err := InitializeVersionFileWithFeedback(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !created {
			t.Errorf("expected created=true, got false")
		}
	})
}

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
		v, err := ParseVersion(tt.raw)
		if err != nil {
			t.Errorf("ParseVersion(%q) failed: %v", tt.raw, err)
			continue
		}
		if v.String() != tt.expected {
			t.Errorf("expected %q, got %q", tt.expected, v.String())
		}
	}
}

func TestParseVersion_ValidWithVPrefix(t *testing.T) {
	v, err := ParseVersion("v1.2.3")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v.Major != 1 || v.Minor != 2 || v.Patch != 3 {
		t.Errorf("unexpected version parsed: %+v", v)
	}
}

func TestParseVersion_ErrorCases(t *testing.T) {
	t.Run("invalid format (missing patch)", func(t *testing.T) {
		_, err := ParseVersion("1.2")
		if err == nil || !errors.Is(err, errInvalidVersion) {
			t.Errorf("expected ErrInvalidVersion, got %v", err)
		}
	})

	t.Run("non-numeric major", func(t *testing.T) {
		_, err := ParseVersion("a.2.3")
		if err == nil || !errors.Is(err, errInvalidVersion) {
			t.Errorf("expected ErrInvalidVersion, got %v", err)
		}
	})

	t.Run("non-numeric minor", func(t *testing.T) {
		_, err := ParseVersion("1.b.3")
		if err == nil || !errors.Is(err, errInvalidVersion) {
			t.Errorf("expected ErrInvalidVersion, got %v", err)
		}
	})

	t.Run("non-numeric patch", func(t *testing.T) {
		_, err := ParseVersion("1.2.c")
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
		"1.2.3.4", // too many parts
	}

	for _, raw := range invalidVersions {
		_, err := ParseVersion(raw)
		if err == nil {
			t.Errorf("expected error for invalid version %q, got nil", raw)
		}
	}
}

func TestParseVersion_NumberConversionErrors(t *testing.T) {
	tests := []struct {
		input         string
		expectedError string
	}{
		{"a.2.3", "invalid major version"},
		{"1.b.3", "invalid minor version"},
		{"1.2.c", "invalid patch version"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			_, err := ParseVersion(tt.input)
			if err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !strings.Contains(err.Error(), tt.expectedError) {
				t.Errorf("expected error to contain %q, got %v", tt.expectedError, err)
			}
		})
	}
}

/* ------------------------------------------------------------------------- */
/* VERSION UPDATES                                                           */
/* ------------------------------------------------------------------------- */

func TestUpdateVersion_Scenarios(t *testing.T) {
	tests := []struct {
		name        string
		initial     string
		level       string
		pre         string
		meta        string
		preserve    bool
		expected    string
		expectErr   bool
		expectedErr string
	}{
		{"patch bump", "1.2.3", "patch", "", "", false, "1.2.4", false, ""},
		{"minor bump", "1.2.3", "minor", "", "", false, "1.3.0", false, ""},
		{"major bump", "1.2.3", "major", "", "", false, "2.0.0", false, ""},
		{"with pre-release", "1.2.3", "patch", "alpha.1", "", false, "1.2.4-alpha.1", false, ""},
		{"with metadata", "1.2.3", "patch", "", "ci.123", false, "1.2.4+ci.123", false, ""},
		{"with pre + metadata", "1.2.3", "patch", "rc.1", "ci.456", false, "1.2.4-rc.1+ci.456", false, ""},
		{"preserve metadata", "1.2.3+build.789", "patch", "", "", true, "1.2.4+build.789", false, ""},
		{"clear metadata", "1.2.3+build.789", "patch", "", "", false, "1.2.4", false, ""},
		{"preserve metadata but override", "1.2.3+build.789", "patch", "", "custom.1", true, "1.2.4+custom.1", false, ""},
		{"invalid bump level", "1.2.3", "banana", "", "", false, "", true, "invalid bump type"},
		{"invalid initial version", "not-a-version", "patch", "", "", false, "", true, "invalid version format"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := writeTempVersion(t, tt.initial)
			defer os.Remove(path)

			err := UpdateVersion(path, tt.level, tt.pre, tt.meta, tt.preserve)

			if tt.expectErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if !strings.Contains(err.Error(), tt.expectedErr) {
					t.Errorf("expected error to contain %q, got %v", tt.expectedErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			got := strings.TrimSpace(readFile(t, path))
			if got != tt.expected {
				t.Errorf("expected version %q, got %q", tt.expected, got)
			}
		})
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

	err := InitializeVersionFileFunc(versionPath)
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

	err = InitializeVersionFileFunc(versionPath)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	data, _ := os.ReadFile(versionPath)
	got := strings.TrimSpace(string(data))
	if got != "1.2.3" {
		t.Errorf("expected file content to remain '1.2.3', got %q", got)
	}
}

func TestSaveVersion_MkdirAllFails(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a file where the directory is expected
	conflictPath := filepath.Join(tmpDir, "conflict")
	if err := os.WriteFile(conflictPath, []byte("not a dir"), 0644); err != nil {
		t.Fatal(err)
	}

	versionFile := filepath.Join(conflictPath, ".version") // invalid: parent is a file

	err := SaveVersion(versionFile, SemVersion{1, 2, 3, "", ""})
	if err == nil {
		t.Fatal("expected error due to mkdir on a file, got nil")
	}

	if !strings.Contains(err.Error(), "not a directory") && !strings.Contains(err.Error(), "is a file") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestInitializeVersionFile_InvalidGitTagFormat(t *testing.T) {
	tmpDir := t.TempDir()
	versionPath := filepath.Join(tmpDir, ".version")

	execCommand = fakeExecCommand("invalid-tag\n")
	defer func() { execCommand = exec.Command }()

	err := InitializeVersionFileFunc(versionPath)
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

func TestInitializeVersionFileWithFeedback_InitializationFails(t *testing.T) {
	tmp := t.TempDir()
	noWrite := filepath.Join(tmp, "nowrite")
	if err := os.Mkdir(noWrite, 0555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = os.Chmod(noWrite, 0755)
	})

	versionPath := filepath.Join(noWrite, ".version")

	created, err := InitializeVersionFileWithFeedback(versionPath)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if created {
		t.Errorf("expected created to be false, got true")
	}
}

func TestInitializeVersionFileWithFeedback_FileCreatedButInvalidContent(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, ".version")

	original := InitializeVersionFileFunc
	InitializeVersionFileFunc = func(path string) error {
		return os.WriteFile(path, []byte("not-a-version\n"), 0600)
	}
	defer func() { InitializeVersionFileFunc = original }()

	created, err := InitializeVersionFileWithFeedback(path)
	if err != nil {
		t.Fatalf("unexpected error from init: %v", err)
	}
	if !created {
		t.Errorf("expected created=true, got false")
	}

	// Must manually read and check failure
	_, err = ReadVersion(path)
	if err == nil {
		t.Fatal("expected error from ReadVersion, got nil")
	}
	if !strings.Contains(err.Error(), "invalid version format") {
		t.Errorf("unexpected error: %v", err)
	}
}

/* ------------------------------------------------------------------------- */
/* BUMP NEXT                                                                 */
/* ------------------------------------------------------------------------- */

func TestBumpNext(t *testing.T) {
	tests := []struct {
		name     string
		current  SemVersion
		expected SemVersion
	}{
		{
			name: "promote alpha pre-release",
			current: SemVersion{
				Major: 1, Minor: 2, Patch: 3, PreRelease: "alpha.1",
			},
			expected: SemVersion{
				Major: 1, Minor: 2, Patch: 3,
			},
		},
		{
			name: "promote rc pre-release",
			current: SemVersion{
				Major: 1, Minor: 2, Patch: 3, PreRelease: "rc.1",
			},
			expected: SemVersion{
				Major: 1, Minor: 2, Patch: 3,
			},
		},
		{
			name: "default patch bump",
			current: SemVersion{
				Major: 1, Minor: 2, Patch: 3,
			},
			expected: SemVersion{
				Major: 1, Minor: 2, Patch: 4,
			},
		},
		{
			name: "promote 0.x alpha to final",
			current: SemVersion{
				Major: 0, Minor: 9, Patch: 0, PreRelease: "alpha.1",
			},
			expected: SemVersion{
				Major: 0, Minor: 9, Patch: 0,
			},
		},
		{
			name: "optional heuristic bump from 0.9.0 to 0.10.0",
			current: SemVersion{
				Major: 0, Minor: 9, Patch: 0,
			},
			expected: SemVersion{
				Major: 0, Minor: 10, Patch: 0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := BumpNext(tt.current)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected.String(), got.String())
			}
		})
	}
}

/* ------------------------------------------------------------------------- */
/* BUMP BY LABEL                                                             */
/* ------------------------------------------------------------------------- */

func TestBumpByLabel(t *testing.T) {
	tests := []struct {
		name     string
		current  SemVersion
		label    string
		expected string
		wantErr  bool
	}{
		{"patch bump", SemVersion{1, 2, 3, "", ""}, "patch", "1.2.4", false},
		{"minor bump", SemVersion{1, 2, 3, "", ""}, "minor", "1.3.0", false},
		{"major bump", SemVersion{1, 2, 3, "", ""}, "major", "2.0.0", false},
		{"invalid label", SemVersion{1, 2, 3, "", ""}, "foobar", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := BumpByLabel(tt.current, tt.label)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.String() != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, got.String())
			}
		})
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
