package commitparser

import (
	"fmt"
	"os"
)

var (
	defaultCommitParser    CommitParser
	RegisterCommitParserFn = registerCommitParser
	GetCommitParserFn      = getCommitParser
)

func registerCommitParser(p CommitParser) {
	if defaultCommitParser != nil {
		fmt.Fprintf(os.Stderr,
			"WARNING: Ignoring commit parser %q: another parser (%q) is already registered.\n",
			p.Name(), defaultCommitParser.Name(),
		)
		return
	}
	defaultCommitParser = p
}

func getCommitParser() CommitParser {
	return defaultCommitParser
}

func ResetCommitParser() {
	defaultCommitParser = nil
}
