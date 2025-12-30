package bumpcmd

import (
	"context"
	"fmt"
	"os"

	"github.com/indaco/semver-cli/internal/clix"
	"github.com/indaco/semver-cli/internal/config"
	"github.com/indaco/semver-cli/internal/hooks"
	"github.com/indaco/semver-cli/internal/plugins/commitparser"
	"github.com/indaco/semver-cli/internal/plugins/commitparser/gitlog"
	"github.com/indaco/semver-cli/internal/semver"
	"github.com/urfave/cli/v3"
)

var tryInferBumpTypeFromCommitParserPluginFn = tryInferBumpTypeFromCommitParserPlugin

// autoCmd returns the "auto" subcommand.
func autoCmd(cfg *config.Config) *cli.Command {
	return &cli.Command{
		Name:    "auto",
		Aliases: []string{"next"},
		Usage:   "Smart bump logic (e.g. promote pre-release or bump patch)",
		UsageText: `semver bump auto [--label patch|minor|major] [--meta data] [--preserve-meta] [--since ref] [--until ref] [--no-infer]

By default, semver tries to infer the bump type from recent commit messages using the built-in commit-parser plugin.
You can override this behavior with the --label flag, disable it explicitly with --no-infer, or disable the plugin via the config file (.semver.yaml).`,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "label",
				Usage: "Optional bump label override (patch, minor, major)",
			},
			&cli.StringFlag{
				Name:  "meta",
				Usage: "Set build metadata (e.g. 'ci.123')",
			},
			&cli.BoolFlag{
				Name:  "preserve-meta",
				Usage: "Preserve existing build metadata instead of clearing it",
			},
			&cli.StringFlag{
				Name:  "since",
				Usage: "Start commit/tag for bump inference (default: last tag or HEAD~10)",
			},
			&cli.StringFlag{
				Name:  "until",
				Usage: "End commit/tag for bump inference (default: HEAD)",
			},
			&cli.BoolFlag{
				Name:  "no-infer",
				Usage: "Disable bump inference from commit messages (overrides config)",
			},
			&cli.BoolFlag{
				Name:  "hook-only",
				Usage: "Only run pre-release hooks, do not modify the version",
			},
			&cli.BoolFlag{
				Name:  "skip-hooks",
				Usage: "Skip pre-release hooks",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			return runBumpAuto(cfg, cmd)
		},
	}
}

// runBumpAuto performs smart bumping (e.g. promote, patch, infer).
func runBumpAuto(cfg *config.Config, cmd *cli.Command) error {
	path := cmd.String("path")
	label := cmd.String("label")
	meta := cmd.String("meta")
	since := cmd.String("since")
	until := cmd.String("until")
	isPreserveMeta := cmd.Bool("preserve-meta")
	isNoInferFlag := cmd.Bool("no-infer")
	isSkipHooks := cmd.Bool("skip-hooks")

	disableInfer := isNoInferFlag || (cfg != nil && cfg.Plugins != nil && !cfg.Plugins.CommitParser)

	if _, err := clix.FromCommandFn(cmd); err != nil {
		return err
	}

	current, err := semver.ReadVersion(path)
	if err != nil {
		return fmt.Errorf("failed to read version: %w", err)
	}

	if err := hooks.RunPreReleaseHooksFn(isSkipHooks); err != nil {
		return err
	}

	next, err := getNextVersion(current, label, disableInfer, since, until, isPreserveMeta)
	if err != nil {
		return err
	}

	next = setBuildMetadata(current, next, meta, isPreserveMeta)

	if err := semver.SaveVersion(path, next); err != nil {
		return fmt.Errorf("failed to save version: %w", err)
	}

	fmt.Printf("Bumped version from %s to %s\n", current.String(), next.String())
	return nil
}

// getNextVersion determines the next semantic version based on the provided label,
// commit inference, or default bump logic. It returns an error if bumping fails
// or if an invalid label is specified.
func getNextVersion(
	current semver.SemVersion,
	label string,
	disableInfer bool,
	since, until string,
	preserveMeta bool,
) (semver.SemVersion, error) {
	var next semver.SemVersion
	var err error

	switch label {
	case "patch", "minor", "major":
		next, err = semver.BumpByLabelFunc(current, label)
		if err != nil {
			return semver.SemVersion{}, fmt.Errorf("failed to bump version with label: %w", err)
		}
	case "":
		if !disableInfer {
			inferred := tryInferBumpTypeFromCommitParserPluginFn(since, until)
			if inferred != "" {
				fmt.Fprintf(os.Stderr, "🔍 Inferred bump type: %s\n", inferred)

				if current.PreRelease != "" {
					return promotePreRelease(current, preserveMeta), nil
				}
				next, err = semver.BumpByLabelFunc(current, inferred)
				if err != nil {
					return semver.SemVersion{}, fmt.Errorf("failed to bump inferred version: %w", err)
				}
				return next, nil
			}
		}

		next, err = semver.BumpNextFunc(current)
		if err != nil {
			return semver.SemVersion{}, fmt.Errorf("failed to determine next version: %w", err)
		}
	default:
		return semver.SemVersion{}, cli.Exit("invalid --label: must be 'patch', 'minor', or 'major'", 1)
	}

	return next, nil
}

// setBuildMetadata updates the build metadata of the next version based on
// the provided meta string and the preserve flag.
func setBuildMetadata(current, next semver.SemVersion, meta string, preserve bool) semver.SemVersion {
	switch {
	case meta != "":
		next.Build = meta
	case preserve:
		next.Build = current.Build
	default:
		next.Build = ""
	}
	return next
}

// promotePreRelease strips pre-release and optionally preserves metadata.
func promotePreRelease(current semver.SemVersion, preserveMeta bool) semver.SemVersion {
	next := current
	next.PreRelease = ""
	if preserveMeta {
		next.Build = current.Build
	} else {
		next.Build = ""
	}
	return next
}

// tryInferBumpTypeFromCommitParserPlugin tries to infer bump type from commit messages.
func tryInferBumpTypeFromCommitParserPlugin(since, until string) string {
	parser := commitparser.GetCommitParserFn()
	if parser == nil {
		return ""
	}

	commits, err := gitlog.GetCommitsFn(since, until)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to read commits: %v\n", err)
		return ""
	}

	label, err := parser.Parse(commits)
	if err != nil {
		fmt.Fprintf(os.Stderr, "commit parser failed: %v\n", err)
		return ""
	}

	return label
}
