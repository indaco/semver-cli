package plugins

import (
	"fmt"
	"os"
)

type CommitParser interface {
	Name() string
	Parse(commits []string) (string, error)
}

var defaultCommitParser CommitParser

func RegisterCommitParser(p CommitParser) {
	if defaultCommitParser != nil {
		fmt.Fprintf(os.Stderr,
			"⚠️  Ignoring commit parser %q: another parser (%q) is already registered.\n",
			p.Name(), defaultCommitParser.Name(),
		)
		return
	}
	defaultCommitParser = p
}

func GetCommitParser() CommitParser {
	return defaultCommitParser
}

func ResetCommitParser() {
	defaultCommitParser = nil
}
