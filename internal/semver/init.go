package semver

import (
	"context"
)

// InitializeVersionFileWithFeedback initializes the version file and returns
// whether a new file was created.
// This function uses the legacy InitializeVersionFileFunc for backward compatibility.
// For better testability, use VersionManager.InitializeWithFeedback() instead.
func InitializeVersionFileWithFeedback(path string) (created bool, err error) {
	// Use the new VersionManager with a background context
	return defaultManager.InitializeWithFeedback(context.Background(), path)
}
