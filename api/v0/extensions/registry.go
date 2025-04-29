package extensions

var metadataRegistry []Extension

func RegisterExtension(p Extension) {
	metadataRegistry = append(metadataRegistry, p)
}

func AllExtensions() []Extension {
	return metadataRegistry
}

func ResetExtension() {
	metadataRegistry = nil
}
