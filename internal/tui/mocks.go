package tui

import (
	"sync"

	"github.com/indaco/verso/internal/workspace"
)

// MockPrompter is a mock implementation of Prompter for testing.
// It allows tests to control the selection results without requiring user interaction.
type MockPrompter struct {
	mu sync.Mutex

	// SelectionResult is the pre-configured selection to return.
	SelectionResult Selection

	// SelectionError is the error to return from PromptModuleSelection.
	SelectionError error

	// ConfirmResult is the result to return from ConfirmOperation.
	ConfirmResult bool

	// ConfirmError is the error to return from ConfirmOperation.
	ConfirmError error

	// Calls records all method invocations for assertion.
	Calls []MockCall
}

// MockCall records a method invocation on MockPrompter.
type MockCall struct {
	Method  string
	Modules []*workspace.Module
	Message string
}

// NewMockPrompter creates a new MockPrompter.
func NewMockPrompter() *MockPrompter {
	return &MockPrompter{
		Calls: make([]MockCall, 0),
	}
}

// Ensure MockPrompter implements Prompter.
var _ Prompter = (*MockPrompter)(nil)

// PromptModuleSelection returns the pre-configured SelectionResult.
func (m *MockPrompter) PromptModuleSelection(modules []*workspace.Module) (Selection, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.Calls = append(m.Calls, MockCall{
		Method:  "PromptModuleSelection",
		Modules: modules,
	})

	if m.SelectionError != nil {
		return Selection{}, m.SelectionError
	}

	return m.SelectionResult, nil
}

// ConfirmOperation returns the pre-configured ConfirmResult.
func (m *MockPrompter) ConfirmOperation(message string) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.Calls = append(m.Calls, MockCall{
		Method:  "ConfirmOperation",
		Message: message,
	})

	if m.ConfirmError != nil {
		return false, m.ConfirmError
	}

	return m.ConfirmResult, nil
}

// Reset clears all recorded calls and resets state.
func (m *MockPrompter) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.Calls = make([]MockCall, 0)
	m.SelectionResult = Selection{}
	m.SelectionError = nil
	m.ConfirmResult = false
	m.ConfirmError = nil
}

// CallCount returns the number of times a method was called.
func (m *MockPrompter) CallCount(method string) int {
	m.mu.Lock()
	defer m.mu.Unlock()

	count := 0
	for _, call := range m.Calls {
		if call.Method == method {
			count++
		}
	}
	return count
}

// LastCall returns the most recent call to the specified method.
// Returns nil if the method was never called.
func (m *MockPrompter) LastCall(method string) *MockCall {
	m.mu.Lock()
	defer m.mu.Unlock()

	for i := len(m.Calls) - 1; i >= 0; i-- {
		if m.Calls[i].Method == method {
			return &m.Calls[i]
		}
	}
	return nil
}

// WithSelectionResult sets the selection result and returns the mock for chaining.
func (m *MockPrompter) WithSelectionResult(result Selection) *MockPrompter {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.SelectionResult = result
	return m
}

// WithSelectionError sets the selection error and returns the mock for chaining.
func (m *MockPrompter) WithSelectionError(err error) *MockPrompter {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.SelectionError = err
	return m
}

// WithConfirmResult sets the confirm result and returns the mock for chaining.
func (m *MockPrompter) WithConfirmResult(result bool) *MockPrompter {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ConfirmResult = result
	return m
}

// WithConfirmError sets the confirm error and returns the mock for chaining.
func (m *MockPrompter) WithConfirmError(err error) *MockPrompter {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ConfirmError = err
	return m
}
