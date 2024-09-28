package version

import (
	_ "embed"
	"strings"
)

var (
	// go:embed VERSION
	version string
	VERSION = strings.TrimSpace(version)
)
