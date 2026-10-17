// File: internal/render/json.go
// Description: Machine-readable JSON output for pkglist (the default, non
//   --verbose mode).
// Date: 2026-10-17

// Package render formats a package list for either machine consumption
// (JSON) or a human reading it in a terminal (a colored ASCII table).
package render

import (
	"encoding/json"
	"io"

	"pkglist/internal/provider"
)

// JSON writes pkgs to w as an indented JSON array. A nil slice is written
// as "[]" rather than JSON "null", so scripts never have to special-case
// an empty result.
func JSON(w io.Writer, pkgs []provider.Package) error {
	if pkgs == nil {
		pkgs = []provider.Package{}
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(pkgs)
}
