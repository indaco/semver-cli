package testutils

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
}

func (m MockExtension) Name() string        { return m.NameValue }
func (m MockExtension) Version() string     { return m.VersionValue }
func (m MockExtension) Description() string { return m.DescriptionValue }

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
