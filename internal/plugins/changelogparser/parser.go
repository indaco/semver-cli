package changelogparser

import (
	"bufio"
	"errors"
	"io"
	"os"
	"regexp"
	"strings"
)

// Keep a Changelog section patterns for parsing.
var (
	// Matches section headers like "## [Unreleased]" or "## [1.2.3] - 2024-01-15"
	sectionHeaderRe = regexp.MustCompile(`^##\s+\[([^\]]+)\]`)
	// Matches subsection headers like "### Added", "### Fixed", etc.
	subsectionHeaderRe = regexp.MustCompile(`^###\s+(.+)$`)
)

// Function variables for testability.
var (
	openFileFn = os.Open
)

// ChangelogSection represents a parsed section from CHANGELOG.md.
type ChangelogSection struct {
	// Version is the version string (e.g., "Unreleased", "1.2.3")
	Version string
	// Date is the release date (empty for Unreleased)
	Date string
	// Subsections maps subsection names to their content lines
	Subsections map[string][]string
}

// UnreleasedSection represents the parsed Unreleased section with change types.
type UnreleasedSection struct {
	// HasEntries indicates if the section has any content
	HasEntries bool
	// Added contains "Added" subsection entries
	Added []string
	// Changed contains "Changed" subsection entries
	Changed []string
	// Deprecated contains "Deprecated" subsection entries
	Deprecated []string
	// Removed contains "Removed" subsection entries
	Removed []string
	// Fixed contains "Fixed" subsection entries
	Fixed []string
	// Security contains "Security" subsection entries
	Security []string
	// Subsections is a helper map for internal parsing
	Subsections map[string][]string
}

// changelogFileParser parses CHANGELOG.md files in Keep a Changelog format.
type changelogFileParser struct {
	path string
}

// newChangelogFileParser creates a new changelog parser for the given file path.
func newChangelogFileParser(path string) *changelogFileParser {
	return &changelogFileParser{path: path}
}

// ParseUnreleased extracts and parses the Unreleased section from CHANGELOG.md.
func (p *changelogFileParser) ParseUnreleased() (*UnreleasedSection, error) {
	file, err := openFileFn(p.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errors.New("changelog file not found")
		}
		return nil, err
	}
	defer file.Close()

	section, err := p.parseUnreleasedSection(file)
	if err != nil {
		return nil, err
	}

	return section, nil
}

// parseUnreleasedSection reads the file and extracts the Unreleased section.
func (p *changelogFileParser) parseUnreleasedSection(reader io.Reader) (*UnreleasedSection, error) {
	scanner := bufio.NewScanner(reader)
	inUnreleased := false
	currentSubsection := ""
	section := &UnreleasedSection{
		Subsections: make(map[string][]string),
	}

	for scanner.Scan() {
		line := scanner.Text()
		trimmedLine := strings.TrimSpace(line)

		// Check for section header
		if matches := sectionHeaderRe.FindStringSubmatch(line); matches != nil {
			versionName := matches[1]
			if strings.EqualFold(versionName, "Unreleased") {
				inUnreleased = true
				continue
			} else if inUnreleased {
				// We've hit the next version section, stop parsing
				break
			}
		}

		if !inUnreleased {
			continue
		}

		// Check for subsection header (e.g., "### Added")
		if matches := subsectionHeaderRe.FindStringSubmatch(line); matches != nil {
			currentSubsection = strings.TrimSpace(matches[1])
			section.Subsections[currentSubsection] = []string{}
			continue
		}

		// Parse entries within a subsection (lines starting with "- ")
		if currentSubsection != "" && strings.HasPrefix(trimmedLine, "- ") {
			entry := strings.TrimPrefix(trimmedLine, "- ")
			section.Subsections[currentSubsection] = append(section.Subsections[currentSubsection], entry)
			section.HasEntries = true
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	if !inUnreleased {
		return nil, errors.New("unreleased section not found in changelog")
	}

	// Map subsections to specific fields
	p.mapSubsectionsToFields(section)

	return section, nil
}

// mapSubsectionsToFields maps the generic subsections map to specific fields.
func (p *changelogFileParser) mapSubsectionsToFields(section *UnreleasedSection) {
	for name, entries := range section.Subsections {
		normalized := strings.ToLower(strings.TrimSpace(name))
		switch normalized {
		case "added":
			section.Added = entries
		case "changed":
			section.Changed = entries
		case "deprecated":
			section.Deprecated = entries
		case "removed":
			section.Removed = entries
		case "fixed":
			section.Fixed = entries
		case "security":
			section.Security = entries
		}
	}
}

// InferBumpType determines the bump type based on changelog entries.
// Priority: major (Removed/Changed) > minor (Added) > patch (Fixed/Security/Deprecated)
func (s *UnreleasedSection) InferBumpType() (string, error) {
	if !s.HasEntries {
		return "", errors.New("no changelog entries found in unreleased section")
	}

	// Major: Removed or Changed sections (breaking changes)
	if len(s.Removed) > 0 {
		return "major", nil
	}

	// Changed can indicate major or minor depending on content
	// For simplicity, we treat any Changed as major (conservative approach)
	if len(s.Changed) > 0 {
		return "major", nil
	}

	// Minor: Added section (new features)
	if len(s.Added) > 0 {
		return "minor", nil
	}

	// Patch: Fixed, Security, or Deprecated
	if len(s.Fixed) > 0 || len(s.Security) > 0 || len(s.Deprecated) > 0 {
		return "patch", nil
	}

	// No recognized sections with content
	return "", errors.New("no bump type could be inferred from changelog")
}
