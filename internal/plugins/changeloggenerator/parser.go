package changeloggenerator

import (
	"regexp"
	"strings"
)

// ParsedCommit represents a fully parsed conventional commit.
type ParsedCommit struct {
	CommitInfo
	Type        string // feat, fix, docs, etc.
	Scope       string // Optional scope in parentheses
	Description string // The commit description after the colon
	Breaking    bool   // Has breaking change indicator (! or BREAKING CHANGE footer)
	PRNumber    string // Extracted PR/MR number if present
}

// Regex patterns for conventional commit parsing.
var (
	// Matches: type(scope)!: description or type!: description or type: description
	conventionalCommitRe = regexp.MustCompile(`^(\w+)(?:\(([^)]+)\))?(!)?:\s*(.+)$`)

	// Matches: (#123) or (closes #123) etc at end of message
	prNumberRe = regexp.MustCompile(`\(?#(\d+)\)?`)
)

// ParseConventionalCommit parses a commit message into its components.
// Returns nil if the commit doesn't follow conventional commit format.
func ParseConventionalCommit(commit CommitInfo) *ParsedCommit {
	matches := conventionalCommitRe.FindStringSubmatch(commit.Subject)
	if matches == nil {
		// Not a conventional commit, return with just the subject as description
		return &ParsedCommit{
			CommitInfo:  commit,
			Type:        "",
			Description: commit.Subject,
		}
	}

	parsed := &ParsedCommit{
		CommitInfo:  commit,
		Type:        strings.ToLower(matches[1]),
		Scope:       matches[2],
		Breaking:    matches[3] == "!",
		Description: matches[4],
	}

	// Extract PR number from description and remove it from the description text
	if prMatches := prNumberRe.FindStringSubmatch(parsed.Description); len(prMatches) == 2 {
		parsed.PRNumber = prMatches[1]
		// Remove the PR reference from description to avoid duplication in output
		parsed.Description = strings.TrimSpace(prNumberRe.ReplaceAllString(parsed.Description, ""))
	}

	return parsed
}

// ParseCommits parses a slice of CommitInfo into ParsedCommits.
func ParseCommits(commits []CommitInfo) []*ParsedCommit {
	parsed := make([]*ParsedCommit, 0, len(commits))
	for _, c := range commits {
		parsed = append(parsed, ParseConventionalCommit(c))
	}
	return parsed
}

// FilterCommits filters out commits matching exclude patterns.
func FilterCommits(commits []*ParsedCommit, excludePatterns []string) []*ParsedCommit {
	if len(excludePatterns) == 0 {
		return commits
	}

	// Compile patterns
	patterns := make([]*regexp.Regexp, 0, len(excludePatterns))
	for _, p := range excludePatterns {
		re, err := regexp.Compile(p)
		if err != nil {
			continue // Skip invalid patterns
		}
		patterns = append(patterns, re)
	}

	filtered := make([]*ParsedCommit, 0, len(commits))
	for _, c := range commits {
		excluded := false
		for _, re := range patterns {
			if re.MatchString(c.Subject) {
				excluded = true
				break
			}
		}
		if !excluded {
			filtered = append(filtered, c)
		}
	}

	return filtered
}

// GroupedCommit represents a commit with its group assignment.
type GroupedCommit struct {
	*ParsedCommit
	GroupLabel string
	GroupIcon  string
	GroupOrder int
}

// GroupCommitsResult contains grouped commits and any skipped non-conventional commits.
type GroupCommitsResult struct {
	Grouped                map[string][]*GroupedCommit
	SkippedNonConventional []*ParsedCommit
}

// GroupCommits groups parsed commits by their type using the configured groups.
// The order is derived from the position in the groups slice (index) unless
// explicitly overridden by the Order field (if > 0).
// If includeNonConventional is true, commits without a type are included in "Other Changes".
// If false, they are returned in SkippedNonConventional for warning purposes.
func GroupCommits(commits []*ParsedCommit, groups []GroupConfig) map[string][]*GroupedCommit {
	result := GroupCommitsWithOptions(commits, groups, false)
	return result.Grouped
}

// GroupCommitsWithOptions groups commits with configurable handling of non-conventional commits.
func GroupCommitsWithOptions(commits []*ParsedCommit, groups []GroupConfig, includeNonConventional bool) GroupCommitsResult {
	result := GroupCommitsResult{
		Grouped:                make(map[string][]*GroupedCommit),
		SkippedNonConventional: make([]*ParsedCommit, 0),
	}

	// Compile group patterns with derived order from index
	type compiledGroup struct {
		GroupConfig
		re    *regexp.Regexp
		order int
	}
	compiledGroups := make([]compiledGroup, 0, len(groups))
	for i, g := range groups {
		re, err := regexp.Compile(g.Pattern)
		if err != nil {
			continue
		}
		// Use explicit Order if set (> 0), otherwise derive from array position
		order := i
		if g.Order > 0 {
			order = g.Order
		}
		compiledGroups = append(compiledGroups, compiledGroup{GroupConfig: g, re: re, order: order})
	}

	for _, commit := range commits {
		matched := false
		for _, group := range compiledGroups {
			// Match against the commit type (e.g., "feat", "fix")
			// or the full subject for non-conventional commits
			matchTarget := commit.Type
			if matchTarget == "" {
				matchTarget = commit.Subject
			}

			if group.re.MatchString(matchTarget) {
				gc := &GroupedCommit{
					ParsedCommit: commit,
					GroupLabel:   group.Label,
					GroupIcon:    group.Icon,
					GroupOrder:   group.order,
				}
				result.Grouped[group.Label] = append(result.Grouped[group.Label], gc)
				matched = true
				break
			}
		}

		// Handle unmatched commits
		if !matched {
			switch {
			case commit.Type != "":
				// Conventional commit with unrecognized type -> "Other"
				gc := &GroupedCommit{
					ParsedCommit: commit,
					GroupLabel:   "Other",
					GroupIcon:    "",
					GroupOrder:   999,
				}
				result.Grouped["Other"] = append(result.Grouped["Other"], gc)
			case includeNonConventional:
				// Non-conventional commit included in "Other Changes"
				gc := &GroupedCommit{
					ParsedCommit: commit,
					GroupLabel:   "Other Changes",
					GroupIcon:    "",
					GroupOrder:   1000,
				}
				result.Grouped["Other Changes"] = append(result.Grouped["Other Changes"], gc)
			default:
				// Non-conventional commit skipped (tracked for warning)
				result.SkippedNonConventional = append(result.SkippedNonConventional, commit)
			}
		}
	}

	return result
}

// SortedGroupKeys returns group labels sorted by their order.
func SortedGroupKeys(grouped map[string][]*GroupedCommit) []string {
	type groupInfo struct {
		label string
		order int
	}

	infos := make([]groupInfo, 0, len(grouped))
	for label, commits := range grouped {
		if len(commits) > 0 {
			infos = append(infos, groupInfo{label: label, order: commits[0].GroupOrder})
		}
	}

	// Simple insertion sort (groups are small)
	for i := 1; i < len(infos); i++ {
		for j := i; j > 0 && infos[j].order < infos[j-1].order; j-- {
			infos[j], infos[j-1] = infos[j-1], infos[j]
		}
	}

	keys := make([]string, len(infos))
	for i, info := range infos {
		keys[i] = info.label
	}
	return keys
}
