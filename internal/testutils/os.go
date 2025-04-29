package testutils

import (
	"os"
	"strings"
)

// IsWindows returns true if the current OS is Windows.
func IsWindows() bool {
	return strings.Contains(strings.ToLower(os.Getenv("OS")), "windows")
}
