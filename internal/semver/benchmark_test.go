package semver

import (
	"os"
	"path/filepath"
	"testing"
)

// BenchmarkParseVersion measures version parsing performance.
func BenchmarkParseVersion(b *testing.B) {
	versions := []struct {
		name    string
		version string
	}{
		{"simple", "1.2.3"},
		{"with_v_prefix", "v1.2.3"},
		{"with_prerelease", "1.2.3-alpha.1"},
		{"with_build", "1.2.3+build.123"},
		{"full", "1.2.3-alpha.1+build.123"},
		{"long_prerelease", "1.2.3-alpha.beta.gamma.delta.1"},
	}

	for _, v := range versions {
		b.Run(v.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_, _ = ParseVersion(v.version)
			}
		})
	}
}

// BenchmarkSemVersion_String measures version string generation.
func BenchmarkSemVersion_String(b *testing.B) {
	versions := []struct {
		name    string
		version SemVersion
	}{
		{"simple", SemVersion{Major: 1, Minor: 2, Patch: 3}},
		{"with_prerelease", SemVersion{Major: 1, Minor: 2, Patch: 3, PreRelease: "alpha.1"}},
		{"with_build", SemVersion{Major: 1, Minor: 2, Patch: 3, Build: "build.123"}},
		{"full", SemVersion{Major: 1, Minor: 2, Patch: 3, PreRelease: "alpha.1", Build: "build.123"}},
	}

	for _, v := range versions {
		b.Run(v.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_ = v.version.String()
			}
		})
	}
}

// BenchmarkBumpByLabel measures bump performance.
func BenchmarkBumpByLabel(b *testing.B) {
	v := SemVersion{Major: 1, Minor: 2, Patch: 3}

	labels := []string{"patch", "minor", "major"}
	for _, label := range labels {
		b.Run(label, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_, _ = BumpByLabel(v, label)
			}
		})
	}
}

// BenchmarkBumpNext measures smart bump performance.
func BenchmarkBumpNext(b *testing.B) {
	versions := []struct {
		name    string
		version SemVersion
	}{
		{"final", SemVersion{Major: 1, Minor: 2, Patch: 3}},
		{"prerelease", SemVersion{Major: 1, Minor: 2, Patch: 3, PreRelease: "alpha.1"}},
	}

	for _, v := range versions {
		b.Run(v.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_, _ = BumpNext(v.version)
			}
		})
	}
}

// BenchmarkIncrementPreRelease measures pre-release increment performance.
func BenchmarkIncrementPreRelease(b *testing.B) {
	cases := []struct {
		name    string
		current string
		base    string
	}{
		{"new", "", "alpha"},
		{"same_base", "alpha", "alpha"},
		{"increment", "alpha.1", "alpha"},
		{"high_number", "alpha.999", "alpha"},
	}

	for _, tc := range cases {
		b.Run(tc.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_ = IncrementPreRelease(tc.current, tc.base)
			}
		})
	}
}

// BenchmarkReadVersion measures file reading performance.
func BenchmarkReadVersion(b *testing.B) {
	// Create a temp file with a version
	tmpDir := b.TempDir()
	path := filepath.Join(tmpDir, ".version")
	if err := os.WriteFile(path, []byte("1.2.3-alpha.1+build.123\n"), 0600); err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	for b.Loop() {
		_, _ = ReadVersion(path)
	}
}

// BenchmarkSaveVersion measures file writing performance.
func BenchmarkSaveVersion(b *testing.B) {
	tmpDir := b.TempDir()
	path := filepath.Join(tmpDir, ".version")
	v := SemVersion{Major: 1, Minor: 2, Patch: 3, PreRelease: "alpha.1", Build: "build.123"}

	b.ReportAllocs()
	for b.Loop() {
		_ = SaveVersion(path, v)
	}
}

// BenchmarkUpdateVersion measures the full update cycle.
func BenchmarkUpdateVersion(b *testing.B) {
	tmpDir := b.TempDir()
	path := filepath.Join(tmpDir, ".version")
	if err := os.WriteFile(path, []byte("1.2.3\n"), 0600); err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	for i := 0; b.Loop(); i++ {
		// Reset version to avoid unbounded growth
		if i%100 == 0 {
			_ = os.WriteFile(path, []byte("1.2.3\n"), 0600)
		}
		_ = UpdateVersion(path, "patch", "", "", false)
	}
}

// BenchmarkParseVersion_Parallel measures parallel parsing performance.
func BenchmarkParseVersion_Parallel(b *testing.B) {
	version := "1.2.3-alpha.1+build.123"

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = ParseVersion(version)
		}
	})
}
