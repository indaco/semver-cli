package plugins

var registry []Plugin

func Register(p Plugin) {
	registry = append(registry, p)
}

func All() []Plugin {
	return registry
}

// ResetPlugins empties the internal registry between tests
func ResetPlugins() {
	registry = nil
}
