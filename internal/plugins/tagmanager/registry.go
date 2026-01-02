package tagmanager

import (
	"fmt"
	"os"
)

var (
	defaultTagManager    TagManager
	RegisterTagManagerFn = registerTagManager
	GetTagManagerFn      = getTagManager
)

func registerTagManager(tm TagManager) {
	if defaultTagManager != nil {
		fmt.Fprintf(os.Stderr,
			"WARNING: Ignoring tag manager %q: another manager (%q) is already registered.\n",
			tm.Name(), defaultTagManager.Name(),
		)
		return
	}
	defaultTagManager = tm
}

func getTagManager() TagManager {
	return defaultTagManager
}

// ResetTagManager clears the registered tag manager (for testing).
func ResetTagManager() {
	defaultTagManager = nil
}

// Register registers the tag manager plugin with the given configuration.
func Register(cfg *Config) {
	RegisterTagManagerFn(NewTagManager(cfg))
}
