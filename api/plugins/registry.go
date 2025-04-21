package plugins

var metadataRegistry []Plugin

func RegisterPlugin(p Plugin) {
	metadataRegistry = append(metadataRegistry, p)
}

func AllPlugins() []Plugin {
	return metadataRegistry
}

func ResetPlugin() {
	metadataRegistry = nil
}
