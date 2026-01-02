package changelogparser

import (
	"fmt"
	"os"
)

var (
	defaultChangelogParser    ChangelogInferrer
	RegisterChangelogParserFn = registerChangelogParser
	GetChangelogParserFn      = getChangelogParser
)

func registerChangelogParser(p ChangelogInferrer) {
	if defaultChangelogParser != nil {
		fmt.Fprintf(os.Stderr,
			"WARNING: Ignoring changelog parser %q: another parser (%q) is already registered.\n",
			p.Name(), defaultChangelogParser.Name(),
		)
		return
	}
	defaultChangelogParser = p
}

func getChangelogParser() ChangelogInferrer {
	return defaultChangelogParser
}

// Register registers the changelog parser plugin with the semver plugin system.
func Register(cfg *Config) {
	RegisterChangelogParserFn(NewChangelogParser(cfg))
}

// Unregister removes the changelog parser plugin.
func Unregister() {
	defaultChangelogParser = nil
}

// ResetChangelogParser clears the registered changelog parser (for testing).
func ResetChangelogParser() {
	defaultChangelogParser = nil
}
