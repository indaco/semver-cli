package testutils

import (
	"context"
	"errors"
	"testing"

	"github.com/urfave/cli/v3"
)

var ErrTestFail = errors.New("test failure")

func TestBuildCLIForTests(t *testing.T) {
	path := "/test/.version"
	subCmds := []*cli.Command{
		{
			Name: "test-cmd",
			Action: func(ctx context.Context, cmd *cli.Command) error {
				return nil
			},
		},
	}

	appCli := BuildCLIForTests(path, subCmds)

	if appCli.Name != "semver" {
		t.Errorf("BuildCLIForTests() name = %q, want %q", appCli.Name, "semver")
	}

	if len(appCli.Flags) != 2 {
		t.Errorf("BuildCLIForTests() should have 2 flags, got %d", len(appCli.Flags))
	}

	if len(appCli.Commands) != 1 {
		t.Errorf("BuildCLIForTests() should have 1 command, got %d", len(appCli.Commands))
	}
}

func TestRunCLITest(t *testing.T) {
	tmpDir := t.TempDir()

	appCli := BuildCLIForTests("", []*cli.Command{
		{
			Name: "noop",
			Action: func(ctx context.Context, cmd *cli.Command) error {
				return nil
			},
		},
	})

	// This should not fail
	RunCLITest(t, appCli, []string{"semver", "noop"}, tmpDir)
}

func TestRunCLITestAllowError(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("success", func(t *testing.T) {
		appCli := BuildCLIForTests("", []*cli.Command{
			{
				Name: "success",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					return nil
				},
			},
		})

		err := RunCLITestAllowError(t, appCli, []string{"semver", "success"}, tmpDir)
		if err != nil {
			t.Errorf("RunCLITestAllowError() unexpected error = %v", err)
		}
	})

	t.Run("error", func(t *testing.T) {
		appCli := BuildCLIForTests("", []*cli.Command{
			{
				Name: "fail",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					return ErrTestFail
				},
			},
		})

		err := RunCLITestAllowError(t, appCli, []string{"semver", "fail"}, tmpDir)
		if err == nil {
			t.Error("RunCLITestAllowError() expected error, got nil")
		}
	})
}

func TestWithWorkingDir(t *testing.T) {
	tmpDir := t.TempDir()

	called := false
	err := withWorkingDir(t, tmpDir, func() error {
		called = true
		return nil
	})

	if err != nil {
		t.Errorf("withWorkingDir() error = %v", err)
	}

	if !called {
		t.Error("withWorkingDir() did not call the function")
	}
}
