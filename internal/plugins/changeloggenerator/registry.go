package changeloggenerator

import (
	"fmt"
	"os"

	"github.com/indaco/verso/internal/config"
)

var (
	defaultChangelogGenerator    ChangelogGenerator
	RegisterChangelogGeneratorFn = registerChangelogGenerator
	GetChangelogGeneratorFn      = getChangelogGenerator
)

func registerChangelogGenerator(cg ChangelogGenerator) {
	if defaultChangelogGenerator != nil {
		fmt.Fprintf(os.Stderr,
			"WARNING: Ignoring changelog generator %q: another generator (%q) is already registered.\n",
			cg.Name(), defaultChangelogGenerator.Name(),
		)
		return
	}
	defaultChangelogGenerator = cg
}

func getChangelogGenerator() ChangelogGenerator {
	return defaultChangelogGenerator
}

// ResetChangelogGenerator clears the registered changelog generator (for testing).
func ResetChangelogGenerator() {
	defaultChangelogGenerator = nil
}

// Register registers the changelog generator plugin with the given configuration.
func Register(cfg *config.ChangelogGeneratorConfig) {
	internalCfg := FromConfigStruct(cfg)
	RegisterChangelogGeneratorFn(NewChangelogGenerator(internalCfg))
}
