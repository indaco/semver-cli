package semver

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// SemVersion represents a semantic version (major.minor.patch-preRelease).
type SemVersion struct {
	Major      int
	Minor      int
	Patch      int
	PreRelease string
	Build      string
}

const VersionFilePerm = 0600

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

	// execCommand is a wrapper for exec.Command used to run external commands (e.g., git).
	// It can be overridden in tests for mocking behavior.
	execCommand = exec.Command

	// InitializeVersionFileFunc is a function variable that initializes the version file.
	// It points to the default initializeVersionFile implementation but can be overridden in tests.
	InitializeVersionFileFunc = initializeVersionFile

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

// InitializeVersionFile initializes a .version file at the given path.
// If the file already exists, it does nothing.
// If not, it attempts to use the latest git tag (if valid), or falls back to 0.1.0.
func initializeVersionFile(path string) error {
	if _, err := os.Stat(path); err == nil {
		return nil // Already exists
	}

	cmd := execCommand("git", "describe", "--tags", "--abbrev=0")
	output, err := cmd.Output()
	version := SemVersion{Major: 0, Minor: 1, Patch: 0} // Default

	if err == nil {
		tag := strings.TrimSpace(string(output))
		tag = strings.TrimPrefix(tag, "v")

		if parsed, parseErr := ParseVersion(tag); parseErr == nil {
			version = parsed
		}
	}

	return SaveVersion(path, version)
}

func InitializeVersionFileWithFeedback(path string) (created bool, err error) {
	if _, err := os.Stat(path); err == nil {
		// File exists → not created; do NOT read or parse
		return false, nil
	}

	err = InitializeVersionFileFunc(path)
	if err != nil {
		return false, err
	}

	return true, nil
}

// ReadVersion reads a version string from the given file and parses it into a SemVersion.
func ReadVersion(path string) (SemVersion, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return SemVersion{}, err
	}
	return ParseVersion(string(data))
}

// SaveVersion writes a SemVersion to the given file path.
func SaveVersion(path string, version SemVersion) error {
	// Ensure parent directory exists
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(version.String()+"\n"), VersionFilePerm)
}

// UpdateVersion updates the semantic version in the given file based on the bump type (patch, minor, major),
// and optionally sets the pre-release and build metadata strings.
// If preserve is true and meta is empty, existing build metadata is retained.
func UpdateVersion(path string, bumpType string, pre string, meta string, preserve bool) error {
	version, err := ReadVersion(path)
	if err != nil {
		return err
	}

	switch bumpType {
	case "patch":
		version.Patch++
	case "minor":
		version.Minor++
		version.Patch = 0
	case "major":
		version.Major++
		version.Minor = 0
		version.Patch = 0
	default:
		return fmt.Errorf("invalid bump type: %s", bumpType)
	}

	// Always reset pre-release when bumping
	version.PreRelease = pre

	// Handle build metadata
	if meta != "" {
		version.Build = meta
	} else if !preserve {
		version.Build = ""
	}

	return SaveVersion(path, version)
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

func formatPreRelease(base string, num int) string {
	return fmt.Sprintf("%s.%d", base, num)
}
