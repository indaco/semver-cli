package operations

import (
	"context"
	"fmt"

	"github.com/indaco/semver-cli/internal/core"
	"github.com/indaco/semver-cli/internal/semver"
	"github.com/indaco/semver-cli/internal/workspace"
)

// SetOperation sets the version of a module to a specific value.
type SetOperation struct {
	fs      core.FileSystem
	version string
}

// NewSetOperation creates a new set operation.
func NewSetOperation(fs core.FileSystem, version string) *SetOperation {
	return &SetOperation{
		fs:      fs,
		version: version,
	}
}

// Execute sets the version on the module.
func (op *SetOperation) Execute(ctx context.Context, mod *workspace.Module) error {
	// Check for context cancellation
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Parse the version string
	newVer, err := semver.ParseVersion(op.version)
	if err != nil {
		return fmt.Errorf("invalid version %q: %w", op.version, err)
	}

	// Create version manager
	vm := semver.NewVersionManager(op.fs, nil)

	// Write the new version
	if err := vm.Save(mod.Path, newVer); err != nil {
		return fmt.Errorf("failed to write version to %s: %w", mod.Path, err)
	}

	// Update module's current version for display
	mod.CurrentVersion = newVer.String()

	return nil
}

// Name returns the name of this operation.
func (op *SetOperation) Name() string {
	return fmt.Sprintf("set %s", op.version)
}
