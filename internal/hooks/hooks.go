package hooks

import (
	"fmt"

	"github.com/indaco/semver-cli/internal/console"
)

type PreReleaseHook interface {
	HookName() string
	Run() error
}

var preReleaseHooks []PreReleaseHook

func RegisterPreReleaseHook(h PreReleaseHook) {
	preReleaseHooks = append(preReleaseHooks, h)
}

func GetPreReleaseHooks() []PreReleaseHook {
	return preReleaseHooks
}

func ResetPreReleaseHooks() {
	preReleaseHooks = nil
}

func RunPreReleaseHooks(skip bool) error {
	if skip {
		return nil
	}

	for _, hook := range GetPreReleaseHooks() {
		fmt.Printf("🔧 Running pre-release hook: %s... ", hook.HookName())
		if err := hook.Run(); err != nil {
			console.PrintFailure("❌ FAIL")
			return fmt.Errorf("pre-release hook %q failed: %w", hook.HookName(), err)
		}
		console.PrintSuccess("✅ OK")
	}

	return nil
}
