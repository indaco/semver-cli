package bumpcmd

import (
	"context"

	"github.com/indaco/semver-cli/internal/config"
	"github.com/indaco/semver-cli/internal/extensionmgr"
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
