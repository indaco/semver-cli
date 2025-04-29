package config

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/goccy/go-yaml"
	"github.com/indaco/semver-cli/internal/testutils"
)

/* ------------------------------------------------------------------------- */
/* LOAD CONFIG                                                               */
/* ------------------------------------------------------------------------- */

func TestLoadConfig(t *testing.T) {
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}

	type setupFn func(t *testing.T) func()

	tests := []struct {
		name     string
		setup    setupFn
		wantPath string
		wantErr  bool
		wantNil  bool
	}{
		{
			name: "from env",
			setup: func(t *testing.T) func() {
				os.Setenv("SEMVER_PATH", "env-defined/.version")
				return func() {
					os.Unsetenv("SEMVER_PATH")
				}
			},
			wantPath: "env-defined/.version",
		},
		{
			name: "valid yaml file with path",
			setup: func(t *testing.T) func() {
				content := "path: ./my-folder/.version\n"
				tmpPath := testutils.WriteTempConfig(t, content)
				if err := os.Chdir(filepath.Dir(tmpPath)); err != nil {
					t.Fatalf("failed to chdir: %v", err)
				}
				return func() { _ = os.Chdir(origDir) }
			},
			wantPath: "./my-folder/.version",
		},
		{
			name: "missing file fallback",
			setup: func(t *testing.T) func() {
				tmpDir := t.TempDir()
				if err := os.Chdir(tmpDir); err != nil {
					t.Fatal(err)
				}
				return func() { _ = os.Chdir(origDir) }
			},
			wantNil: true,
		},
		{
			name: "empty config falls back to default path",
			setup: func(t *testing.T) func() {
				content := "{}\n"
				tmpPath := testutils.WriteTempConfig(t, content)
				if err := os.Chdir(filepath.Dir(tmpPath)); err != nil {
					t.Fatalf("failed to chdir: %v", err)
				}
				return func() { _ = os.Chdir(origDir) }
			},
			wantPath: ".version",
		},
		{
			name: "invalid yaml (bad format)",
			setup: func(t *testing.T) func() {
				content := "not_yaml::: true"
				tmpPath := testutils.WriteTempConfig(t, content)
				if err := os.Chdir(filepath.Dir(tmpPath)); err != nil {
					t.Fatalf("failed to chdir: %v", err)
				}
				return func() { _ = os.Chdir(origDir) }
			},
			wantErr: true,
			wantNil: true,
		},
		{
			name: "unmarshal error (syntax)",
			setup: func(t *testing.T) func() {
				content := ": this is invalid"
				tmpPath := testutils.WriteTempConfig(t, content)
				if err := os.Chdir(filepath.Dir(tmpPath)); err != nil {
					t.Fatalf("failed to chdir: %v", err)
				}
				return func() { _ = os.Chdir(origDir) }
			},
			wantErr: true,
			wantNil: true,
		},
		{
			name: "read file error (directory instead of file)",
			setup: func(t *testing.T) func() {
				tmpDir := t.TempDir()
				if err := os.Chdir(tmpDir); err != nil {
					t.Fatalf("failed to chdir: %v", err)
				}
				if err := os.Mkdir(".semver.yaml", 0755); err != nil {
					t.Fatal(err)
				}
				return func() { _ = os.Chdir(origDir) }
			},
			wantErr: true,
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			restore := tt.setup(t)
			defer restore()

			cfg, err := LoadConfigFn()
			if (err != nil) != tt.wantErr {
				t.Fatalf("expected err=%v, got err=%v", tt.wantErr, err)
			}
			if tt.wantNil && cfg != nil {
				t.Errorf("expected nil config, got %+v", cfg)
			}
			if !tt.wantNil && cfg == nil {
				t.Fatal("expected non-nil config, got nil")
			}
			if cfg != nil && cfg.Path != tt.wantPath {
				t.Errorf("expected path %q, got %q", tt.wantPath, cfg.Path)
			}
		})
	}
}

/* ------------------------------------------------------------------------- */
/* NORMALIZE VERSION PATH                                                    */
/* ------------------------------------------------------------------------- */

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
/* SAVE CONFIG                                                               */
/* ------------------------------------------------------------------------- */

func TestSaveConfigFn(t *testing.T) {
	t.Run("basic save scenarios", func(t *testing.T) {
		defer func() {
			marshalFn = yaml.Marshal
			openFileFn = os.OpenFile
		}()

		tests := []struct {
			name               string
			cfg                *Config
			wantErr            bool
			overwriteMarshalFn bool
			mockMarshalErr     error
			overwriteOpenFile  bool
		}{
			{
				name:    "save minimal config",
				cfg:     &Config{Path: "my.version"},
				wantErr: false,
			},
			{
				name: "save config with plugins",
				cfg: &Config{
					Path: "custom.version",
					Extensions: []ExtensionConfig{
						{Name: "example", Path: "/plugin/path", Enabled: true},
					},
				},
				wantErr: false,
			},
			{
				name:               "marshal failure",
				cfg:                &Config{Path: "fail.version"},
				wantErr:            true,
				overwriteMarshalFn: true,
				mockMarshalErr:     fmt.Errorf("mock marshal failure"),
			},
			{
				name:              "write fails due to file permission",
				cfg:               &Config{Path: "fail-write.version"},
				wantErr:           true,
				overwriteOpenFile: true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				tmp := t.TempDir()
				_ = os.Chdir(tmp)

				if tt.overwriteMarshalFn {
					marshalFn = func(any) ([]byte, error) {
						return nil, tt.mockMarshalErr
					}
				}

				if tt.overwriteOpenFile {
					openFileFn = func(name string, flag int, perm os.FileMode) (*os.File, error) {
						// Simulate permission denied by opening read-only
						path := filepath.Join(t.TempDir(), "readonly.yaml")
						f, err := os.Create(path)
						if err != nil {
							t.Fatal(err)
						}
						f.Close()
						_ = os.Chmod(path, 0400)
						return os.OpenFile(path, os.O_WRONLY, 0400)
					}
				}

				err := SaveConfigFn(tt.cfg)
				if (err != nil) != tt.wantErr {
					t.Fatalf("SaveConfigFn() error = %v, wantErr = %v", err, tt.wantErr)
				}

				if !tt.wantErr {
					if _, err := os.Stat(".semver.yaml"); err != nil {
						t.Errorf(".semver.yaml was not created: %v", err)
					}
				}
			})
		}
	})

	t.Run("write fails due to directory permission", func(t *testing.T) {
		tmp := t.TempDir()
		badDir := filepath.Join(tmp, "readonly")
		if err := os.Mkdir(badDir, 0500); err != nil {
			t.Fatal(err)
		}
		defer func() {
			if err := os.Chmod(badDir, 0755); err != nil {
				t.Logf("cleanup warning: failed to chmod %q: %v", badDir, err)
			}
		}()

		_ = os.Chdir(badDir)
		err := SaveConfigFn(&Config{Path: "blocked.version"})
		if err == nil {
			t.Error("expected error due to write permission, got nil")
		}
	})
}

func TestSaveConfigFn_WriteFileFn_Error(t *testing.T) {
	origWriteFn := writeFileFn
	defer func() {
		writeFileFn = origWriteFn
	}()

	tmp := t.TempDir()
	_ = os.Chdir(tmp)

	writeFileFn = func(f *os.File, data []byte) (int, error) {
		fmt.Println(">>> writeFileFn invoked")
		return 0, fmt.Errorf("simulated write failure")
	}

	cfg := &Config{Path: "whatever"}
	err := SaveConfigFn(cfg)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	want := "failed to write config data: simulated write failure"
	if err.Error() != want {
		t.Errorf("unexpected error. got: %q, want: %q", err.Error(), want)
	}
}
