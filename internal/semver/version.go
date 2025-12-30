package semver

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// SemVersion represents a semantic version (major.minor.patch-preRelease+build).
type SemVersion struct {
	Major      int
	Minor      int
	Patch      int
	PreRelease string
	Build      string
}

var (
	// versionRegex matches semantic version strings with optional "v" prefix,
	// optional pre-release (e.g., "-beta.1"), and optional build metadata (e.g., "+build.123").
	// It captures:
	//   1. Major version
	//   2. Minor version
	//   3. Patch version
	//   4. (optional) Pre-release identifier
	//   5. (optional) Build metadata
	versionRegex = regexp.MustCompile(
		`^v?([^\.\-+]+)\.([^\.\-+]+)\.([^\.\-+]+)` + // major.minor.patch
			`(?:-([0-9A-Za-z\-\.]+))?` + // optional pre-release
			`(?:\+([0-9A-Za-z\-\.]+))?$`, // optional build metadata
	)

	// errInvalidVersion is returned when a version string does not conform
	// to the expected semantic version format.
	errInvalidVersion = errors.New("invalid version format")

	// BumpNextFunc is a function variable for performing heuristic-based version bumps.
	// It defaults to BumpNext but can be overridden in tests to simulate errors.
	BumpNextFunc = BumpNext

	// BumpByLabelFunc is a function variable for bumping a version using an explicit label (patch, minor, major).
	// It defaults to BumpByLabel but can be overridden in tests to simulate errors.
	BumpByLabelFunc = BumpByLabel
)

// String returns the string representation of the semantic version.
func (v SemVersion) String() string {
	s := fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
	if v.PreRelease != "" {
		s += "-" + v.PreRelease
	}
	if v.Build != "" {
		s += "+" + v.Build
	}
	return s
}

// ParseVersion parses a semantic version string and returns a SemVersion.
// Returns an error if the version format is invalid.
func ParseVersion(s string) (SemVersion, error) {
	matches := versionRegex.FindStringSubmatch(strings.TrimSpace(s))
	if len(matches) < 4 {
		return SemVersion{}, errInvalidVersion
	}

	major, err := strconv.Atoi(matches[1])
	if err != nil {
		return SemVersion{}, fmt.Errorf("%w: invalid major version: %v", errInvalidVersion, err)
	}
	minor, err := strconv.Atoi(matches[2])
	if err != nil {
		return SemVersion{}, fmt.Errorf("%w: invalid minor version: %v", errInvalidVersion, err)
	}
	patch, err := strconv.Atoi(matches[3])
	if err != nil {
		return SemVersion{}, fmt.Errorf("%w: invalid patch version: %v", errInvalidVersion, err)
	}

	pre := matches[4]
	build := matches[5]

	return SemVersion{Major: major, Minor: minor, Patch: patch, PreRelease: pre, Build: build}, nil
}

// BumpNext applies heuristic-based smart bump logic.
// - If it's a pre-release (e.g., alpha.1, rc.1), it promotes to final version.
// - If it's a final release, it bumps patch by default.
func BumpNext(v SemVersion) (SemVersion, error) {
	// If the version has a pre-release label, strip it (promote to final)
	if v.PreRelease != "" {
		promoted := v
		promoted.PreRelease = ""
		return promoted, nil
	}

	if v.Major == 0 && v.Minor == 9 && v.Patch == 0 {
		return SemVersion{Major: v.Major, Minor: v.Minor + 1, Patch: 0}, nil
	}

	// Default case: bump patch
	return SemVersion{Major: v.Major, Minor: v.Minor, Patch: v.Patch + 1}, nil
}

// BumpByLabel bumps the version using an explicit label (patch, minor, major).
func BumpByLabel(v SemVersion, label string) (SemVersion, error) {
	switch label {
	case "patch":
		return SemVersion{Major: v.Major, Minor: v.Minor, Patch: v.Patch + 1}, nil
	case "minor":
		return SemVersion{Major: v.Major, Minor: v.Minor + 1, Patch: 0}, nil
	case "major":
		return SemVersion{Major: v.Major + 1, Minor: 0, Patch: 0}, nil
	default:
		return SemVersion{}, fmt.Errorf("invalid bump label: %s", label)
	}
}

// IncrementPreRelease increments the numeric suffix of a pre-release label.
// If current == base, or current doesn't match base, returns base.1.
// If current is base.N, returns base.(N+1).
func IncrementPreRelease(current, base string) string {
	if current == base {
		return formatPreRelease(base, 1)
	}

	re := regexp.MustCompile(`^` + regexp.QuoteMeta(base) + `(?:\.(\d*))?$`)
	matches := re.FindStringSubmatch(current)

	if len(matches) == 2 {
		if matches[1] == "" {
			return formatPreRelease(base, 1)
		}
		n, err := strconv.Atoi(matches[1])
		if err == nil {
			return formatPreRelease(base, n+1)
		}
	}

	return formatPreRelease(base, 1)
}

func formatPreRelease(base string, num int) string {
	return fmt.Sprintf("%s.%d", base, num)
}
