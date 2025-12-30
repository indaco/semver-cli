// Package semver provides semantic version parsing, manipulation, and persistence.
//
// This package implements the Semantic Versioning 2.0.0 specification
// (https://semver.org/) for parsing and incrementing semantic versions.
// It includes file-based persistence and git tag integration for
// automatic version initialization.
//
// # Version Format
//
// A semantic version consists of:
//   - Major: Incremented for incompatible API changes
//   - Minor: Incremented for backwards-compatible functionality additions
//   - Patch: Incremented for backwards-compatible bug fixes
//   - PreRelease: Optional pre-release label (e.g., "alpha.1", "rc.2")
//   - Build: Optional build metadata (e.g., "build.123", "sha.abc123")
//
// Examples of valid version strings:
//
//	1.0.0
//	v1.2.3
//	1.0.0-alpha
//	1.0.0-alpha.1
//	1.0.0+build.123
//	1.0.0-beta.2+build.456
//
// # Basic Usage
//
// Parse a version string:
//
//	v, err := semver.ParseVersion("1.2.3-alpha.1")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(v.Major, v.Minor, v.Patch) // 1 2 3
//
// Bump a version:
//
//	v, _ := semver.ParseVersion("1.2.3")
//	v2, _ := semver.BumpByLabel(v, "minor")
//	fmt.Println(v2) // 1.3.0
//
// Read and write version files:
//
//	v, err := semver.ReadVersion(".version")
//	semver.SaveVersion(".version", v)
//
// # Thread Safety
//
// The parsing functions (ParseVersion, BumpByLabel, BumpNext, etc.) are
// safe for concurrent use. File operations should be synchronized by
// the caller if used concurrently.
package semver
