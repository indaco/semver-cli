package plugins

import "github.com/urfave/cli/v3"

type Plugin interface {
	Name() string
	Register(root *cli.Command)
}
