package version

import (
	"fmt"
	"runtime/debug"
)

// Version of renku-dev-utils
var Version string = "DEV"

// Version suffix of renku-dev-utils
var VersionSuffix string = ""

func init() {
	if VersionSuffix != "" {
		Version = Version + "-" + VersionSuffix
	}
}

func BB() (version string) {
	bi, ok := debug.ReadBuildInfo()
	if !ok {
		return "unknown"
	}
	version = bi.Main.Version
	for _, setting := range bi.Settings {
		fmt.Printf("%s: %s\n", setting.Key, setting.Value)
	}
	return version
}
