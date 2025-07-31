package main

import (
	"fmt"
	"runtime"
)

// Version information
var (
	// Version is the current version of the application
	Version = "0.1.0"

	// BuildTime is the time the binary was built
	BuildTime = "unknown"

	// CommitHash is the git commit hash the binary was built from
	CommitHash = "unknown"
)

// VersionInfo returns a formatted string with version information
func VersionInfo() string {
	return fmt.Sprintf("%s (built: %s, commit: %s, %s/%s)",
		Version, BuildTime, CommitHash, runtime.GOOS, runtime.GOARCH)
}

// ShortVersionInfo returns just the version number
func ShortVersionInfo() string {
	return Version
}
