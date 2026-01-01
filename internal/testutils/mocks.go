package testutils

import "fmt"

// MockPlugin implements only Plugin
type MockPlugin struct {
	NameValue        string
	VersionValue     string
	DescriptionValue string
}

func (m MockPlugin) Name() string        { return m.NameValue }
func (m MockPlugin) Version() string     { return m.VersionValue }
func (m MockPlugin) Description() string { return m.DescriptionValue }

// MockExtension implements only Extension
type MockExtension struct {
	NameValue        string
	VersionValue     string
	DescriptionValue string
	HooksValue       []string
	EntryValue       string
}

func (m MockExtension) Name() string        { return m.NameValue }
func (m MockExtension) Version() string     { return m.VersionValue }
func (m MockExtension) Description() string { return m.DescriptionValue }
func (m MockExtension) Hooks() []string     { return m.HooksValue }
func (m MockExtension) Entry() string       { return m.EntryValue }

// MockCommitParser implements the plugins.CommitParser interface for testing.
type MockCommitParser struct {
	Label string
	Err   error
}

func (m MockCommitParser) Name() string {
	return "mock"
}
func (m MockCommitParser) Parse(_ []string) (string, error) {
	return m.Label, m.Err
}

func WithMock(setup func(), testFunc func()) {
	setup()
	testFunc()
}

// MockHook implements PreReleaseHook for testing
type MockHook struct {
	Name      string
	ShouldErr bool
}

func (m MockHook) HookName() string {
	return m.Name
}

func (m MockHook) Run() error {
	if m.ShouldErr {
		return fmt.Errorf("%s failed", m.Name)
	}
	return nil
}
