// Package version provides build-time version information for the application.
//
// Version is injected at build time via ldflags:
//
//	go build -ldflags="-X github.com/kirasync2748/github-widget/internal/version.Version=v1.0.0"
//
// When not set (e.g. during `go run`), it defaults to "dev".
package version

// Version holds the application version string.
// It is overridden at build time via -ldflags.
var Version = "dev"
