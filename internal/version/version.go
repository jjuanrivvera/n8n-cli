// Package version exposes build metadata injected at link time via -ldflags.
package version

import (
	"fmt"
	"runtime"
)

// These are overridden at build time:
//
//	go build -ldflags "-X .../internal/version.Version=1.2.3 -X ....Commit=abc -X ....BuildDate=..."
var (
	Version   = "dev"
	Commit    = "none"
	BuildDate = "unknown"
)

// Short returns just the semantic version (e.g. "1.2.3").
func Short() string { return Version }

// Info returns a one-line human-readable build description.
func Info() string {
	return fmt.Sprintf("n8nctl %s (commit %s, built %s, %s/%s, %s)",
		Version, Commit, BuildDate, runtime.GOOS, runtime.GOARCH, runtime.Version())
}

// UserAgent is the value sent in the User-Agent header on every API request.
func UserAgent() string {
	return fmt.Sprintf("n8nctl/%s (%s/%s)", Version, runtime.GOOS, runtime.GOARCH)
}
