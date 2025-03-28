package semver

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
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
}

const VersionFilePerm = 0600

var (
	// versionRegex matches semantic version strings with optional pre-release.
	versionRegex = regexp.MustCompile(`^(\d+)\.(\d+)\.(\d+)(?:-([\w.-]+))?$`)

	// execCommand is used to run external commands (e.g., git).
	// It can be overridden in tests.
	execCommand = exec.Command

	errInvalidVersion = errors.New("invalid version format")
)

// String returns the string representation of the semantic version.
func (v SemVersion) String() string {
	if v.PreRelease != "" {
		return fmt.Sprintf("%d.%d.%d-%s", v.Major, v.Minor, v.Patch, v.PreRelease)
	}
	return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
}

// InitializeVersionFile initializes a .version file at the given path.
// If the file already exists, it does nothing.
// If not, it attempts to use the latest git tag (if valid), or falls back to 0.1.0.
func InitializeVersionFile(path string) error {
	if _, err := os.Stat(path); err == nil {
		return nil // Already exists
	}

	cmd := execCommand("git", "describe", "--tags", "--abbrev=0")
	output, err := cmd.Output()
	version := SemVersion{Major: 0, Minor: 1, Patch: 0} // Default

	if err == nil {
		tag := strings.TrimSpace(string(output))
		tag = strings.TrimPrefix(tag, "v")

		if parsed, parseErr := parseVersion(tag); parseErr == nil {
			version = parsed
		}
	}

	return SaveVersion(path, version)
}

// ReadVersion reads a version string from the given file and parses it into a SemVersion.
func ReadVersion(path string) (SemVersion, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return SemVersion{}, err
	}
	return parseVersion(string(data))
}

// SaveVersion writes a SemVersion to the given file path.
func SaveVersion(path string, version SemVersion) error {
	return os.WriteFile(path, []byte(version.String()+"\n"), VersionFilePerm)
}

// UpdateVersion bumps the version at the given file path by the specified level:
// "patch", "minor", or "major". It resets any pre-release identifiers.
func UpdateVersion(path, level string) error {
	version, err := ReadVersion(path)
	if err != nil {
		return err
	}

	switch level {
	case "patch":
		version.Patch++
		version.PreRelease = ""
	case "minor":
		version.Minor++
		version.Patch = 0
		version.PreRelease = ""
	case "major":
		version.Major++
		version.Minor = 0
		version.Patch = 0
		version.PreRelease = ""
	default:
		return errors.New("unknown level: " + level)
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

// parseVersion parses a semantic version string and returns a SemVersion.
// Returns an error if the version format is invalid.
func parseVersion(s string) (SemVersion, error) {
	matches := versionRegex.FindStringSubmatch(strings.TrimSpace(s))
	if len(matches) < 4 {
		return SemVersion{}, errInvalidVersion
	}

	// The regex ensures these will always be digits, but parse defensively
	major, err := strconv.Atoi(matches[1])
	if err != nil {
		return SemVersion{}, fmt.Errorf("invalid major version: %w", err)
	}
	minor, err := strconv.Atoi(matches[2])
	if err != nil {
		return SemVersion{}, fmt.Errorf("invalid minor version: %w", err)
	}
	patch, err := strconv.Atoi(matches[3])
	if err != nil {
		return SemVersion{}, fmt.Errorf("invalid patch version: %w", err)
	}
	pre := matches[4]

	return SemVersion{major, minor, patch, pre}, nil
}

func formatPreRelease(base string, num int) string {
	return fmt.Sprintf("%s.%d", base, num)
}
