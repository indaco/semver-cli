package operations

import (
	"context"
	"fmt"

	"github.com/indaco/semver-cli/internal/core"
	"github.com/indaco/semver-cli/internal/semver"
	"github.com/indaco/semver-cli/internal/workspace"
)

// ShowOperation reads and displays the current version of a module.
type ShowOperation struct {
	fs core.FileSystem
}

// NewShowOperation creates a new show operation.
func NewShowOperation(fs core.FileSystem) *ShowOperation {
	return &ShowOperation{
		fs: fs,
	}
}

// Execute reads the version from the module.
// The version is stored in the module's CurrentVersion field.
func (op *ShowOperation) Execute(ctx context.Context, mod *workspace.Module) error {
	// Check for context cancellation
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Create version manager
	vm := semver.NewVersionManager(op.fs, nil)

	// Read current version
	ver, err := vm.Read(mod.Path)
	if err != nil {
		return fmt.Errorf("failed to read version from %s: %w", mod.Path, err)
	}

	// Update module's current version
	mod.CurrentVersion = ver.String()

	return nil
}

// Name returns the name of this operation.
func (op *ShowOperation) Name() string {
	return "show"
}
