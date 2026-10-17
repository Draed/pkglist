// File: internal/provider/provider.go
// Description: Package provider defines the core Package/Provider types
//   shared by every package source (apt, pipx, mise, go, uv, flatpak,
//   manual).
// Date: 2026-10-17

// Package provider implements every package source pkglist knows how to
// query, behind a single Provider interface.
package provider

import "context"

// Package is one installed package as reported by a provider.
//
// InstallReason is "manual" (explicitly requested by the user), "auto"
// (pulled in only as a dependency of something else), or "unknown" when a
// provider cannot tell the two apart. Only apt distinguishes auto from
// manual (via apt-mark, mirroring `apt-mark showmanual`/`showauto`); every
// other provider here only ever installs what the user explicitly named,
// so they report "manual" unconditionally.
type Package struct {
	Provider      string `json:"provider"`
	Name          string `json:"name"`
	Version       string `json:"version"`
	InstallReason string `json:"install_reason"`
}

// Install reason values shared by every provider.
const (
	InstallReasonManual  = "manual"
	InstallReasonAuto    = "auto"
	InstallReasonUnknown = "unknown"
)

// Provider is a package source (apt, pipx, mise, ...). Adding support for a
// new provider only requires implementing this interface and registering it
// in All.
type Provider interface {
	// Name is the short, lowercase identifier used in output and on the
	// command line (e.g. "apt", "pipx", "mise").
	Name() string
	// Available reports whether the provider's tooling is installed on
	// this system.
	Available() bool
	// List returns every package currently installed through this
	// provider.
	List(ctx context.Context) ([]Package, error)
}
