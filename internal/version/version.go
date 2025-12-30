package version

import "fmt"

var (
	// Version is the application version.
	// This is set at build time using ldflags:
	//   go build -ldflags "-X main.version.Version=v0.1.0"
	Version = "dev"

	// GitCommit is the git commit hash at build time.
	GitCommit = "unknown"

	// BuildTime is the build timestamp.
	BuildTime = "unknown"

	// GoVersion is the Go version used to build.
	GoVersion = "unknown"
)

// VersionInfo returns complete version information.
type VersionInfo struct {
	Version   string
	GitCommit string
	BuildTime string
	GoVersion string
}

// GetVersionInfo returns the complete version information.
func GetVersionInfo() VersionInfo {
	return VersionInfo{
		Version:   Version,
		GitCommit: GitCommit,
		BuildTime: BuildTime,
		GoVersion: GoVersion,
	}
}

// String returns a formatted version string.
func (v VersionInfo) String() string {
	return v.Version
}

// FullString returns a formatted version string with all details.
func (v VersionInfo) FullString() string {
	return fmt.Sprintf("%s (commit: %s, built: %s, go: %s)",
		v.Version, v.GitCommit, v.BuildTime, v.GoVersion)
}

// FormatVersion formats a version with optional details.
func FormatVersion(version, commit, buildTime, goVersion string) string {
	return fmt.Sprintf("%s (commit: %s, built: %s, go: %s)",
		version, commit, buildTime, goVersion)
}
