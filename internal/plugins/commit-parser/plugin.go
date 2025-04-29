package commitparser

import (
	"errors"
	"strings"
)

/* ------------------------------------------------------------------------- */
/* INTERFACES                                                                */
/* ------------------------------------------------------------------------- */

// CommitParser defines the interface for parsing a list of commit messages
// and determining the corresponding semver bump type.

type CommitParser interface {
	Name() string
	Parse(commits []string) (string, error)
}

type CommitParserPlugin struct{}

func (CommitParserPlugin) Name() string { return "commit-parser" }
func (CommitParserPlugin) Description() string {
	return "Parses conventional commits to infer bump type"
}
func (CommitParserPlugin) Version() string { return "v0.1.0" }

/* ------------------------------------------------------------------------- */
/* IMPLEMENTATION                                                            */
/* ------------------------------------------------------------------------- */

// NewCommitParser returns a new Conventional Commits parser.
func newCommitParser() CommitParser {
	return &CommitParserPlugin{}
}

// Parse analyzes a slice of commit messages and infers the semver bump type.
// It returns "major", "minor", "patch", or an error if no inference is possible.
func (p *CommitParserPlugin) Parse(commits []string) (string, error) {
	hasBreaking := false
	hasFeat := false
	hasFix := false

	for _, commit := range commits {
		commit = strings.ToLower(commit)

		if strings.Contains(commit, "breaking change") {
			hasBreaking = true
			continue
		}
		if strings.HasPrefix(commit, "feat") {
			hasFeat = true
			continue
		}
		if strings.HasPrefix(commit, "fix") {
			hasFix = true
			continue
		}
	}

	switch {
	case hasBreaking:
		return "major", nil
	case hasFeat:
		return "minor", nil
	case hasFix:
		return "patch", nil
	default:
		return "", errors.New("no bump type could be inferred")
	}
}

/* ------------------------------------------------------------------------- */
/* REGISTRATION                                                              */
/* ------------------------------------------------------------------------- */

// Register registers the commit parser plugin with the semver plugin system.
func Register() {
	RegisterCommitParserFn(&CommitParserPlugin{})
}
