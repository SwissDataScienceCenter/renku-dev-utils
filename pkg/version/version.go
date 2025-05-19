package version

// Version of renku-dev-utils
var Version string = "DEV"

// Version suffix of renku-dev-utils
var VersionSuffix string = ""

func init() {
	if VersionSuffix != "" {
		Version = Version + "-" + VersionSuffix
	}
}
