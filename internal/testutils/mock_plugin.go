package testutils

// mockPlugin implements only Plugin
type MockPlugin struct {
	NameValue        string
	VersionValue     string
	DescriptionValue string
}

func (m MockPlugin) Name() string        { return m.NameValue }
func (m MockPlugin) Version() string     { return m.VersionValue }
func (m MockPlugin) Description() string { return m.DescriptionValue }
