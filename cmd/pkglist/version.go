// File: cmd/pkglist/version.go
// Description: pkglist's own version string. Overridden at build time via
//   `-ldflags "-X main.version=<tag>"` (see .goreleaser.yaml for the
//   GoReleaser release build, and the vesta repo's go_bin_cont_builder
//   invocation for the ansible-driven container build) - left at "dev"
//   for a plain `go build` with no ldflags.
// Date: 2026-10-25

package main

var version = "dev"
