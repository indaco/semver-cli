package version

import (
	"testing"
)

// TestGetVersion checks if the GetVersion function correctly retrieves the embedded version.
func TestGetVersion(t *testing.T) {
	expectedVersion := "0.5.0-beta.1+build.1"

	got := GetVersion()
	if got != expectedVersion {
		t.Errorf("GetVersion() = %q; want %q", got, expectedVersion)
	}
}
