package testutils

import (
	"context"
	"os"
	"testing"

	"github.com/urfave/cli/v3"
)

func BuildCLIForTests(path string, subCmds []*cli.Command) *cli.Command {
	return &cli.Command{
		Name: "semver",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "path",
				Aliases: []string{"p"},
				Usage:   "Path to .version file",
				Value:   path,
			},
			&cli.BoolFlag{
				Name:  "no-auto-init",
				Usage: "Disable auto-initialization of the .version file",
			},
		},
		Commands: subCmds,
	}
}

// RunCLITest runs a CLI command using the given args in the provided workdir,
// and fails the test if the command returns an error.
func RunCLITest(t *testing.T, appCli *cli.Command, args []string, workdir string) {
	t.Helper()

	err := withWorkingDir(t, workdir, func() error {
		return appCli.Run(context.Background(), args)
	})
	if err != nil {
		t.Fatalf("app.Run failed: %v", err)
	}
}

// RunCLITestAllowError runs a CLI command using the given args in the provided workdir,
// and returns any error instead of failing the test.
func RunCLITestAllowError(t *testing.T, appCli *cli.Command, args []string, workdir string) error {
	t.Helper()
	return withWorkingDir(t, workdir, func() error {
		return appCli.Run(context.Background(), args)
	})
}

// withWorkingDir temporarily changes the working directory, runs fn, and restores the original directory.
func withWorkingDir(t *testing.T, dir string, fn func() error) error {
	t.Helper()

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}

	if err := os.Chdir(dir); err != nil {
		t.Fatalf("failed to change to workdir: %v", err)
	}

	t.Cleanup(func() {
		if err := os.Chdir(origDir); err != nil {
			t.Fatalf("failed to restore working directory: %v", err)
		}
	})

	return fn()
}
