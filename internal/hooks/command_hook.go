package hooks

import (
	"os"
	"os/exec"
)

type CommandHook struct {
	Name    string
	Command string
}

func (h CommandHook) Run() error {
	cmd := exec.Command("sh", "-c", h.Command)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (h CommandHook) HookName() string {
	return h.Name
}
