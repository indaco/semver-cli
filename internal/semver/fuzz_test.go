package semver

import (
	"strings"
	"testing"
)

// FuzzParseVersion tests the version parser with random inputs.
// Run with: go test -fuzz=FuzzParseVersion -fuzztime=30s
func FuzzParseVersion(f *testing.F) {
	// Seed corpus with valid versions
	seeds := []string{
		"1.0.0",
		"0.1.0",
		"1.2.3",
		"10.20.30",
		"1.0.0-alpha",
		"1.0.0-alpha.1",
		"1.0.0-0.3.7",
		"1.0.0-x.7.z.92",
		"1.0.0+20130313144700",
		"1.0.0-beta+exp.sha.5114f85",
		"1.0.0+21AF26D3----117B344092BD",
		"v1.2.3",
		"v1.2.3-alpha.1",
		"v1.2.3-alpha.1+build.123",
		// Edge cases
		"0.0.0",
		"999.999.999",
		"1.0.0-alpha-beta",
		"1.0.0-alpha.beta.gamma",
		// Invalid inputs to test error handling
		"",
		"invalid",
		"1",
		"1.2",
		"1.2.3.4",
		"-1.0.0",
		"1.-2.3",
		"1.2.-3",
		"a.b.c",
		"1.2.3-",
		"1.2.3+",
		"1.2.3-+",
	}

	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, input string) {
		fuzzParseVersionInput(t, input)
	})
}

// fuzzParseVersionInput tests a single input for ParseVersion.
func fuzzParseVersionInput(t *testing.T, input string) {
	t.Helper()

	// The parser should never panic
	v, err := ParseVersion(input)

	if err == nil {
		verifyRoundtrip(t, input, v)
		verifySanity(t, v)
	}

	verifyConsistency(t, input, v, err)
}

// verifyRoundtrip checks that parsing output produces the same version.
func verifyRoundtrip(t *testing.T, input string, v SemVersion) {
	t.Helper()

	output := v.String()
	v2, err2 := ParseVersion(output)
	if err2 != nil {
		t.Errorf("roundtrip failed: %q -> %q -> error: %v", input, output, err2)
		return
	}

	if !versionsEqual(v, v2) {
		t.Errorf("roundtrip mismatch: %q -> %+v -> %q -> %+v", input, v, output, v2)
	}
}

// versionsEqual checks if two versions are equal.
func versionsEqual(v1, v2 SemVersion) bool {
	return v1.Major == v2.Major && v1.Minor == v2.Minor && v1.Patch == v2.Patch &&
		v1.PreRelease == v2.PreRelease && v1.Build == v2.Build
}

// verifySanity checks basic sanity constraints.
func verifySanity(t *testing.T, v SemVersion) {
	t.Helper()

	if v.Major < 0 || v.Minor < 0 || v.Patch < 0 {
		t.Errorf("negative version component: %+v", v)
	}
}

// verifyConsistency checks that parsing is consistent.
func verifyConsistency(t *testing.T, input string, v SemVersion, err error) {
	t.Helper()

	v3, err3 := ParseVersion(input)
	if (err == nil) != (err3 == nil) {
		t.Errorf("inconsistent error state for %q: first=%v, second=%v", input, err, err3)
	}
	if err == nil && (v.Major != v3.Major || v.Minor != v3.Minor || v.Patch != v3.Patch) {
		t.Errorf("inconsistent parse result for %q", input)
	}
}

// FuzzIncrementPreRelease tests the pre-release increment logic with random inputs.
func FuzzIncrementPreRelease(f *testing.F) {
	seeds := []struct {
		current string
		base    string
	}{
		{"alpha", "alpha"},
		{"alpha.1", "alpha"},
		{"alpha.99", "alpha"},
		{"beta", "alpha"},
		{"rc.1", "rc"},
		{"", "alpha"},
		{"alpha.beta", "alpha"},
	}

	for _, seed := range seeds {
		f.Add(seed.current, seed.base)
	}

	f.Fuzz(func(t *testing.T, current, base string) {
		// Should never panic
		result := IncrementPreRelease(current, base)

		// Result should contain the base
		if !strings.HasPrefix(result, base) {
			t.Errorf("result %q should start with base %q", result, base)
		}

		// Result should have format "base.N" where N >= 1
		if !strings.Contains(result, ".") {
			t.Errorf("result %q should contain a dot", result)
		}
	})
}

// FuzzBumpByLabel tests the bump logic with random inputs.
func FuzzBumpByLabel(f *testing.F) {
	seeds := []struct {
		major int
		minor int
		patch int
		label string
	}{
		{1, 2, 3, "patch"},
		{1, 2, 3, "minor"},
		{1, 2, 3, "major"},
		{0, 0, 0, "patch"},
		{999, 999, 999, "patch"},
		{0, 0, 0, "invalid"},
	}

	for _, seed := range seeds {
		f.Add(seed.major, seed.minor, seed.patch, seed.label)
	}

	f.Fuzz(func(t *testing.T, major, minor, patch int, label string) {
		fuzzBumpByLabelInput(t, major, minor, patch, label)
	})
}

// fuzzBumpByLabelInput tests a single input for BumpByLabel.
func fuzzBumpByLabelInput(t *testing.T, major, minor, patch int, label string) {
	t.Helper()

	// Skip invalid inputs
	if !isValidVersionInput(major, minor, patch) {
		return
	}

	v := SemVersion{Major: major, Minor: minor, Patch: patch}
	result, err := BumpByLabel(v, label)

	verifyBumpResult(t, v, result, err, label)
}

// isValidVersionInput checks if version components are valid for testing.
func isValidVersionInput(major, minor, patch int) bool {
	if major < 0 || minor < 0 || patch < 0 {
		return false
	}
	// Cap very large values to avoid overflow issues
	if major > 1000000 || minor > 1000000 || patch > 1000000 {
		return false
	}
	return true
}

// verifyBumpResult checks the result of BumpByLabel.
func verifyBumpResult(t *testing.T, v, result SemVersion, err error, label string) {
	t.Helper()

	validators := map[string]func(){
		"patch": func() { verifyPatchBump(t, v, result, err) },
		"minor": func() { verifyMinorBump(t, v, result, err) },
		"major": func() { verifyMajorBump(t, v, result, err) },
	}

	if validator, ok := validators[label]; ok {
		validator()
	} else if err == nil {
		t.Errorf("expected error for invalid label %q", label)
	}
}

func verifyPatchBump(t *testing.T, v, result SemVersion, err error) {
	t.Helper()
	if err != nil {
		t.Errorf("unexpected error for patch bump: %v", err)
		return
	}
	if result.Patch != v.Patch+1 {
		t.Errorf("patch bump failed: expected %d, got %d", v.Patch+1, result.Patch)
	}
}

func verifyMinorBump(t *testing.T, v, result SemVersion, err error) {
	t.Helper()
	if err != nil {
		t.Errorf("unexpected error for minor bump: %v", err)
		return
	}
	if result.Minor != v.Minor+1 || result.Patch != 0 {
		t.Errorf("minor bump failed: expected %d.%d.0, got %d.%d.%d",
			v.Major, v.Minor+1, result.Major, result.Minor, result.Patch)
	}
}

func verifyMajorBump(t *testing.T, v, result SemVersion, err error) {
	t.Helper()
	if err != nil {
		t.Errorf("unexpected error for major bump: %v", err)
		return
	}
	if result.Major != v.Major+1 || result.Minor != 0 || result.Patch != 0 {
		t.Errorf("major bump failed: expected %d.0.0, got %d.%d.%d",
			v.Major+1, result.Major, result.Minor, result.Patch)
	}
}
