package plugins

import (
	"github.com/indaco/semver-cli/api"
)

var registry []api.Plugin

func Register(p api.Plugin) {
	registry = append(registry, p)
}

func All() []api.Plugin {
	return registry
}
