package bumpcmd

import (
	"context"
	"fmt"

	"github.com/indaco/semver-cli/internal/config"
	"github.com/indaco/semver-cli/internal/extensionmgr"
	"github.com/indaco/semver-cli/internal/plugins/tagmanager"
	"github.com/indaco/semver-cli/internal/plugins/versionvalidator"
	"github.com/indaco/semver-cli/internal/semver"
)

// runPreBumpExtensionHooks runs pre-bump extension hooks if not skipped.
func runPreBumpExtensionHooks(ctx context.Context, cfg *config.Config, newVersion, prevVersion, bumpType string, skipHooks bool) error {
	if skipHooks {
		return nil
	}
	return extensionmgr.RunPreBumpHooks(ctx, cfg, newVersion, prevVersion, bumpType)
}

// runPostBumpExtensionHooks runs post-bump extension hooks if not skipped.
func runPostBumpExtensionHooks(ctx context.Context, cfg *config.Config, path, prevVersion, bumpType string, skipHooks bool) error {
	if skipHooks {
		return nil
	}

	currentVersion, err := semver.ReadVersion(path)
	if err != nil {
		return err
	}

	prereleasePtr, metadataPtr := extractVersionPointers(currentVersion)
	return extensionmgr.RunPostBumpHooks(ctx, cfg, currentVersion.String(), prevVersion, bumpType, prereleasePtr, metadataPtr)
}

// extractVersionPointers extracts prerelease and metadata as pointers (nil if empty).
func extractVersionPointers(v semver.SemVersion) (*string, *string) {
	var prereleasePtr, metadataPtr *string
	if v.PreRelease != "" {
		prereleasePtr = &v.PreRelease
	}
	if v.Build != "" {
		metadataPtr = &v.Build
	}
	return prereleasePtr, metadataPtr
}

// calculateNewBuild determines the build metadata for a new version.
func calculateNewBuild(meta string, preserveMeta bool, currentBuild string) string {
	if meta != "" {
		return meta
	}
	if preserveMeta {
		return currentBuild
	}
	return ""
}

// validateTagAvailable checks if a tag can be created for the version.
// Returns nil if tag manager is not enabled or tag is available.
func validateTagAvailable(version semver.SemVersion) error {
	tm := tagmanager.GetTagManagerFn()
	if tm == nil {
		return nil
	}

	// Check if the plugin is enabled and auto-create is on
	if plugin, ok := tm.(*tagmanager.TagManagerPlugin); ok {
		if !plugin.IsEnabled() {
			return nil
		}
	}

	return tm.ValidateTagAvailable(version)
}

// createTagAfterBump creates a git tag for the version if tag manager is enabled.
func createTagAfterBump(version semver.SemVersion, bumpType string) error {
	tm := tagmanager.GetTagManagerFn()
	if tm == nil {
		return nil
	}

	// Check if the plugin is enabled and auto-create is on
	plugin, ok := tm.(*tagmanager.TagManagerPlugin)
	if !ok || !plugin.IsEnabled() {
		return nil
	}

	message := fmt.Sprintf("Release %s (%s bump)", version.String(), bumpType)
	if err := tm.CreateTag(version, message); err != nil {
		return fmt.Errorf("failed to create tag: %w", err)
	}

	tagName := tm.FormatTagName(version)
	fmt.Printf("Created tag: %s\n", tagName)

	if plugin.GetConfig().Push {
		fmt.Printf("Pushed tag: %s\n", tagName)
	}

	return nil
}

// validateVersionPolicy checks if the version bump is allowed by configured policies.
// Returns nil if version validator is not enabled or validation passes.
func validateVersionPolicy(newVersion, previousVersion semver.SemVersion, bumpType string) error {
	vv := versionvalidator.GetVersionValidatorFn()
	if vv == nil {
		return nil
	}

	// Check if the plugin is enabled
	if plugin, ok := vv.(*versionvalidator.VersionValidatorPlugin); ok {
		if !plugin.IsEnabled() {
			return nil
		}
	}

	return vv.Validate(newVersion, previousVersion, bumpType)
}
