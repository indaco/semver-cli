// Package tui provides interactive terminal user interface components.
// It encapsulates TUI library specifics and provides clean interfaces for testing.
package tui

import "github.com/indaco/semver-cli/internal/workspace"

// Prompter abstracts interactive prompts for module selection.
// This interface enables dependency injection and testing with mock implementations.
type Prompter interface {
	// PromptModuleSelection shows interactive module selection UI.
	// Returns a Selection indicating user choices or an error.
	// Returns ErrCanceled if the user cancels the operation.
	PromptModuleSelection(modules []*workspace.Module) (Selection, error)

	// ConfirmOperation asks for yes/no confirmation.
	// Returns true if confirmed, false if declined.
	ConfirmOperation(message string) (bool, error)
}

// Selection represents the user's module selection from the TUI prompt.
type Selection struct {
	// All indicates if the user selected "apply to all modules".
	All bool

	// Modules contains the names of specific modules selected.
	// Empty if All is true or if operation was canceled.
	Modules []string

	// Canceled indicates if the user canceled the selection.
	Canceled bool
}

// Choice represents the initial selection choice.
type Choice int

const (
	// ChoiceAll indicates user wants to apply to all modules.
	ChoiceAll Choice = iota

	// ChoiceSelect indicates user wants to select specific modules.
	ChoiceSelect

	// ChoiceCancel indicates user wants to cancel the operation.
	ChoiceCancel
)

// String returns a string representation of the choice.
func (c Choice) String() string {
	switch c {
	case ChoiceAll:
		return "all"
	case ChoiceSelect:
		return "select"
	case ChoiceCancel:
		return "cancel"
	default:
		return "unknown"
	}
}

// ParseChoice converts a string to a Choice.
func ParseChoice(s string) Choice {
	switch s {
	case "all":
		return ChoiceAll
	case "select":
		return ChoiceSelect
	case "cancel":
		return ChoiceCancel
	default:
		return ChoiceCancel
	}
}

// AllModules returns a Selection with All set to true.
func AllModules() Selection {
	return Selection{All: true}
}

// SelectedModules returns a Selection with specific module names.
func SelectedModules(names []string) Selection {
	return Selection{Modules: names}
}

// CanceledSelection returns a Selection marked as canceled.
func CanceledSelection() Selection {
	return Selection{Canceled: true}
}
