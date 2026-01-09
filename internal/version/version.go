package version

import (
	_ "embed"
	"strings"
)

//go:embed VERSION
var version string

var VERSION = strings.TrimSpace(version)
