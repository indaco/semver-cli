package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/indaco/semver-cli/api/v0/extensions"
	"github.com/indaco/semver-cli/internal/config"
	extensionmanager "github.com/indaco/semver-cli/internal/extension-manager"
	commitparser "github.com/indaco/semver-cli/internal/plugins/commit-parser"
	"github.com/indaco/semver-cli/internal/plugins/commit-parser/gitlog"
	"github.com/indaco/semver-cli/internal/semver"
	"github.com/urfave/cli/v3"
)

// At package level
var tryInferBumpTypeFromCommitParserPluginFn = tryInferBumpTypeFromCommitParserPlugin

func initVersionCmd() func(ctx context.Context, cmd *cli.Command) error {
	return func(ctx context.Context, cmd *cli.Command) error {
		path := cmd.String("path")

		created, err := semver.InitializeVersionFileWithFeedback(path)
		if err != nil {
			return err
		}

		version, err := semver.ReadVersion(path)
		if err != nil {
			return fmt.Errorf("failed to read version file at %s: %w", path, err)
		}

		if created {
			fmt.Printf("Initialized %s with version %s\n", path, version.String())
		} else {
			fmt.Printf("Version file already exists at %s\n", path)
		}
		return nil
	}
}

// bumpPatch increments the patch version of the .version file.
func bumpPatchCmd() func(ctx context.Context, cmd *cli.Command) error {
	return func(ctx context.Context, cmd *cli.Command) error {
		path := cmd.String("path")
		pre := cmd.String("pre")
		meta := cmd.String("meta")
		preserve := cmd.Bool("preserve-meta")

		if _, err := getOrInitVersionFile(cmd); err != nil {
			return err
		}

		return semver.UpdateVersion(path, "patch", pre, meta, preserve)
	}
}

// bumpMinor increments the minor version and resets patch.
func bumpMinorCmd() func(ctx context.Context, cmd *cli.Command) error {
	return func(ctx context.Context, cmd *cli.Command) error {
		path := cmd.String("path")
		pre := cmd.String("pre")
		meta := cmd.String("meta")
		preserve := cmd.Bool("preserve-meta")

		if _, err := getOrInitVersionFile(cmd); err != nil {
			return err
		}

		return semver.UpdateVersion(path, "minor", pre, meta, preserve)
	}
}

// bumpMajor increments the major version and resets minor/patch.
func bumpMajorCmd() func(ctx context.Context, cmd *cli.Command) error {
	return func(ctx context.Context, cmd *cli.Command) error {
		path := cmd.String("path")
		pre := cmd.String("pre")
		meta := cmd.String("meta")
		preserve := cmd.Bool("preserve-meta")

		if _, err := getOrInitVersionFile(cmd); err != nil {
			return err
		}

		return semver.UpdateVersion(path, "major", pre, meta, preserve)
	}
}

func bumpReleaseCmd() func(ctx context.Context, cmd *cli.Command) error {
	return func(ctx context.Context, cmd *cli.Command) error {
		path := cmd.String("path")
		preserveMeta := cmd.Bool("preserve-meta")

		if _, err := getOrInitVersionFile(cmd); err != nil {
			return err
		}

		version, err := semver.ReadVersion(path)
		if err != nil {
			return fmt.Errorf("failed to read version: %w", err)
		}

		version.PreRelease = ""
		if !preserveMeta {
			version.Build = ""
		}

		if err := semver.SaveVersion(path, version); err != nil {
			return fmt.Errorf("failed to save version: %w", err)
		}

		fmt.Printf("Promoted to release version: %s\n", version.String())
		return nil
	}
}

func bumpNextCmd(cfg *config.Config) func(ctx context.Context, cmd *cli.Command) error {
	return func(ctx context.Context, cmd *cli.Command) error {
		path := cmd.String("path")
		label := cmd.String("label")
		meta := cmd.String("meta")
		since := cmd.String("since")
		until := cmd.String("until")
		isPreserveMeta := cmd.Bool("preserve-meta")
		isNoInferFlag := cmd.Bool("no-infer")

		disableInfer := isNoInferFlag || (cfg != nil && cfg.Plugins != nil && !cfg.Plugins.CommitParser)

		if _, err := getOrInitVersionFile(cmd); err != nil {
			return err
		}

		current, err := semver.ReadVersion(path)
		if err != nil {
			return fmt.Errorf("failed to read version: %w", err)
		}

		var next semver.SemVersion

		switch label {
		case "patch", "minor", "major":
			next, err = semver.BumpByLabelFunc(current, label)
			if err != nil {
				return fmt.Errorf("failed to bump version with label: %w", err)
			}
		case "":
			if !disableInfer {
				inferred := tryInferBumpTypeFromCommitParserPluginFn(since, until)
				if inferred != "" {
					fmt.Fprintf(os.Stderr, "🔍 Inferred bump type: %s\n", inferred)

					if current.PreRelease != "" {
						next = promotePreRelease(current, isPreserveMeta)
					} else {
						next, err = semver.BumpByLabelFunc(current, inferred)
						if err != nil {
							return fmt.Errorf("failed to bump inferred version: %w", err)
						}
					}
					break
				}
			}

			next, err = semver.BumpNextFunc(current)
			if err != nil {
				return fmt.Errorf("failed to determine next version: %w", err)
			}
		default:
			return cli.Exit("invalid --label: must be 'patch', 'minor', or 'major'", 1)
		}

		switch {
		case meta != "":
			next.Build = meta
		case isPreserveMeta:
			next.Build = current.Build
		default:
			next.Build = ""
		}

		if err := semver.SaveVersion(path, next); err != nil {
			return fmt.Errorf("failed to save version: %w", err)
		}

		fmt.Printf("Bumped version from %s to %s\n", current.String(), next.String())
		return nil
	}
}

// setPreRelease sets or increments the pre-release label.
func setPreReleaseCmd() func(ctx context.Context, cmd *cli.Command) error {
	return func(ctx context.Context, cmd *cli.Command) error {
		path := cmd.String("path")
		label := cmd.String("label")
		inc := cmd.Bool("inc")

		if _, err := getOrInitVersionFile(cmd); err != nil {
			return err
		}

		version, err := semver.ReadVersion(path)
		if err != nil {
			return fmt.Errorf("failed to read version: %w", err)
		}

		if inc {
			version.PreRelease = semver.IncrementPreRelease(version.PreRelease, label)
		} else {
			if version.PreRelease == "" {
				version.Patch++
			}
			version.PreRelease = label
		}

		if err := semver.SaveVersion(path, version); err != nil {
			return fmt.Errorf("failed to save version: %w", err)
		}

		return nil
	}
}

func showVersionCmd() func(ctx context.Context, cmd *cli.Command) error {
	return func(ctx context.Context, cmd *cli.Command) error {
		_, err := getOrInitVersionFile(cmd)
		if err != nil {
			return err
		}

		path := cmd.String("path")
		version, err := semver.ReadVersion(path)
		if err != nil {
			return fmt.Errorf("failed to read version file at %s: %w", path, err)
		}

		fmt.Println(version.String())
		return nil
	}
}

func setVersionCmd() func(ctx context.Context, cmd *cli.Command) error {
	return func(ctx context.Context, cmd *cli.Command) error {
		path := cmd.String("path")
		args := cmd.Args()

		if args.Len() < 1 {
			return cli.Exit("missing required version argument", 1)
		}

		raw := args.Get(0)
		pre := cmd.String("pre")
		meta := cmd.String("meta")

		version, err := semver.ParseVersion(raw)
		if err != nil {
			return fmt.Errorf("invalid version: %w", err)
		}
		version.PreRelease = pre
		version.Build = meta

		if err := semver.SaveVersion(path, version); err != nil {
			return fmt.Errorf("failed to save version: %w", err)
		}

		fmt.Printf("Set version to %s in %s\n", version.String(), path)
		return nil
	}
}

func validateVersionCmd() func(ctx context.Context, cmd *cli.Command) error {
	return func(ctx context.Context, cmd *cli.Command) error {
		path := cmd.String("path")
		_, err := semver.ReadVersion(path)
		if err != nil {
			return fmt.Errorf("invalid version file at %s: %w", path, err)
		}
		fmt.Printf("Valid version file at %s\n", path)
		return nil
	}
}

func extensionInstallCmd() func(ctx context.Context, cmd *cli.Command) error {
	return func(ctx context.Context, cmd *cli.Command) error {
		localPath := cmd.String("path")
		if localPath == "" {
			return cli.Exit("missing --path (or --url) for extension registration", 1)
		}

		// Get the extension directory (use the provided flag or default to current directory)
		extensionDirectory := cmd.String("extension-dir")

		// Proceed with normal extension registration
		return extensionmanager.RegisterLocalExtensionFn(localPath, ".semver.yaml", extensionDirectory)
	}
}

func extensionListCmd() func(ctx context.Context, cmd *cli.Command) error {
	return func(ctx context.Context, cmd *cli.Command) error {
		// Load the configuration file
		cfg, err := config.LoadConfigFn()
		if err != nil {
			// Print the error to stdout and return
			fmt.Println("failed to load configuration:", err)
			return nil
		}

		// If there are no plugins, notify the user
		if len(cfg.Extensions) == 0 {
			fmt.Println("No extensions registered.")
			return nil
		}

		// Create a lookup map of metadata
		metadataMap := map[string]extensions.Extension{}
		for _, meta := range extensions.AllExtensions() {
			metadataMap[meta.Name()] = meta
		}

		fmt.Println("List of Registered Extensions:")
		fmt.Println()
		fmt.Println("  NAME              VERSION     ENABLED   DESCRIPTION")
		fmt.Println("  ----------------------------------------------------------")

		for _, p := range cfg.Extensions {
			meta, ok := metadataMap[p.Name]
			version := "?"
			desc := "(no metadata)"
			if ok {
				version = meta.Version()
				desc = meta.Description()
			}

			fmt.Printf("  %-17s %-10s %-9v %s\n", p.Name, version, p.Enabled, desc)
		}

		fmt.Println()

		return nil
	}
}

func extensionRemoveCmd() func(ctx context.Context, cmd *cli.Command) error {
	return func(ctx context.Context, cmd *cli.Command) error {
		// Get the plugin name from the flag
		extensionName := cmd.String("name")
		if extensionName == "" {
			return fmt.Errorf("please provide an extension name to remove")
		}

		cfg, err := config.LoadConfigFn()
		if err != nil {
			fmt.Println("failed to load configuration:", err)
			return nil
		}

		var extensionToRemove *config.ExtensionConfig
		for i, extension := range cfg.Extensions {
			if extension.Name == extensionName {
				extensionToRemove = &cfg.Extensions[i]
				break
			}
		}

		if extensionToRemove == nil {
			fmt.Printf("extension %q not found\n", extensionName)
			return nil
		}

		// Disable the plugin in the configuration (set Enabled to false)
		extensionToRemove.Enabled = false

		// Save the updated config back to the file
		if err := config.SaveConfigFn(cfg); err != nil {
			fmt.Println("failed to save updated configuration:", err)
			return nil
		}

		// Check if --delete-folder flag is set to remove the extension folder
		if cmd.Bool("delete-folder") {
			// Remove the extension directory from ".semver-extensions"
			extensionDir := filepath.Join(".semver-extensions", extensionName)
			if err := os.RemoveAll(extensionDir); err != nil {
				return fmt.Errorf("failed to remove extension directory: %w", err)
			}
			fmt.Printf("✅ Extension %q and its directory removed successfully.\n", extensionName)
		} else {
			fmt.Printf("✅ Extension %q removed, but its directory is preserved.\n", extensionName)
		}

		return nil
	}
}

// getOrInitVersionFile handles .version file initialization or returns an error
// if auto-init is disabled and the file is missing.
func getOrInitVersionFile(cmd *cli.Command) (created bool, err error) {
	path := cmd.String("path")
	noAutoInit := cmd.Bool("no-auto-init")

	if noAutoInit {
		if _, err := os.Stat(path); err != nil {
			return false, cli.Exit(fmt.Sprintf("version file not found at %s", path), 1)
		}
		return false, nil
	}

	created, err = semver.InitializeVersionFileWithFeedback(path)
	if err != nil {
		return false, err
	}
	if created {
		fmt.Printf("Auto-initialized %s with default version\n", path)
	}
	return created, nil
}

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
