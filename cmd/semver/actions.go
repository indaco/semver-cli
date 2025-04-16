package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/indaco/semver-cli/internal/config"
	"github.com/indaco/semver-cli/internal/pluginmanager"
	"github.com/indaco/semver-cli/internal/semver"
	"github.com/urfave/cli/v3"
)

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

func bumpNextCmd() func(ctx context.Context, cmd *cli.Command) error {
	return func(ctx context.Context, cmd *cli.Command) error {
		path := cmd.String("path")
		label := cmd.String("label")
		meta := cmd.String("meta")
		preserveMeta := cmd.Bool("preserve-meta")

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
		case preserveMeta:
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

func pluginAddCmd() func(ctx context.Context, cmd *cli.Command) error {
	return func(ctx context.Context, cmd *cli.Command) error {
		localPath := cmd.String("path")
		if localPath == "" {
			return cli.Exit("missing --path (or --url) for plugin registration", 1)
		}

		// Get the plugin directory (use the provided flag or default to current directory)
		pluginDirectory := cmd.String("plugin-dir")

		// Proceed with normal plugin registration
		return pluginmanager.RegisterLocalPlugin(localPath, ".semver.yaml", pluginDirectory)
	}
}

func pluginListCmd() func(ctx context.Context, cmd *cli.Command) error {
	return func(ctx context.Context, cmd *cli.Command) error {
		// Load the configuration file
		cfg, err := config.LoadConfigFn()
		if err != nil {
			// Print the error to stdout and return
			fmt.Println("failed to load configuration:", err)
			return nil
		}

		// If there are no plugins, notify the user
		if len(cfg.Plugins) == 0 {
			fmt.Println("No plugins registered.")
			return nil
		}

		// Display the list of plugins
		fmt.Println("List of Registered Plugins:")
		for _, plugin := range cfg.Plugins {
			fmt.Printf("Name: %s, Path: %s, Enabled: %v\n", plugin.Name, plugin.Path, plugin.Enabled)
		}

		return nil
	}
}

func pluginRemoveCmd() func(ctx context.Context, cmd *cli.Command) error {
	return func(ctx context.Context, cmd *cli.Command) error {
		// Get the plugin name from the flag
		pluginName := cmd.String("name")
		if pluginName == "" {
			return fmt.Errorf("please provide a plugin name to remove")
		}

		cfg, err := config.LoadConfigFn()
		if err != nil {
			fmt.Println("failed to load configuration:", err)
			return nil
		}

		var pluginToRemove *config.PluginConfig
		for i, plugin := range cfg.Plugins {
			if plugin.Name == pluginName {
				pluginToRemove = &cfg.Plugins[i]
				break
			}
		}

		if pluginToRemove == nil {
			fmt.Printf("plugin %q not found\n", pluginName)
			return nil
		}

		// Disable the plugin in the configuration (set Enabled to false)
		pluginToRemove.Enabled = false

		// Save the updated config back to the file
		if err := config.SaveConfigFn(cfg); err != nil {
			fmt.Println("failed to save updated configuration:", err)
			return nil
		}

		// Check if --delete-folder flag is set to remove the plugin folder
		if cmd.Bool("delete-folder") {
			// Remove the plugin directory from ".semver-plugins"
			pluginDirPath := filepath.Join(".semver-plugins", pluginName)
			if err := os.RemoveAll(pluginDirPath); err != nil {
				return fmt.Errorf("failed to remove plugin directory: %w", err)
			}
			fmt.Printf("✅ Plugin %q and its directory removed successfully.\n", pluginName)
		} else {
			fmt.Printf("✅ Plugin %q removed, but its directory is preserved.\n", pluginName)
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
