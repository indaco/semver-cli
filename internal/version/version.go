package version

import (
	_ "embed"
	"strings"
)

//go:embed .version
var version string

func GetVersion() string {
	return strings.TrimSpace(version)
}
