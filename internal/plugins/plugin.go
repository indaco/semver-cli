package plugins

import "github.com/urfave/cli/v3"

type Plugin interface {
	Name() string
	Register(root *cli.Command)
}

var registry []Plugin

func Register(p Plugin) {
	registry = append(registry, p)
}

func All() []Plugin {
	return registry
}
