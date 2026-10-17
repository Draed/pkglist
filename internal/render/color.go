// File: internal/render/color.go
// Description: ANSI color support for --verbose output - a tan/cream
//   palette with green/yellow/red used to flag install reason, inspired by
//   the git-who CLI's bordered contributor table. All codes are plain
//   7-bit ASCII escape sequences (ESC + printable ASCII), so output stays
//   within ISO/IEC 646 even with color enabled.
// Date: 2026-10-17

package render

import (
	"os"

	"pkglist/internal/provider"
)

const (
	ansiReset  = "\x1b[0m"
	ansiBold   = "\x1b[1m"
	ansiTan    = "\x1b[33m" // borders and the Version column
	ansiGreen  = "\x1b[32m" // install_reason: manual
	ansiYellow = "\x1b[33m" // install_reason: auto
	ansiRed    = "\x1b[31m" // install_reason: unknown
)

// UseColor decides whether --verbose output should be colored: an explicit
// --no-color/--color flag wins, otherwise color is used only when stdout is
// an actual terminal and the NO_COLOR convention (https://no-color.org)
// has not opted out.
func UseColor(noColor, forceColor bool) bool {
	switch {
	case noColor:
		return false
	case forceColor:
		return true
	case os.Getenv("NO_COLOR") != "":
		return false
	default:
		return isTerminal(os.Stdout)
	}
}

func isTerminal(f *os.File) bool {
	info, err := f.Stat()
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeCharDevice != 0
}

func colorize(enabled bool, code, text string) string {
	if !enabled || text == "" {
		return text
	}
	return code + text + ansiReset
}

func reasonColor(reason string) string {
	switch reason {
	case provider.InstallReasonManual:
		return ansiGreen
	case provider.InstallReasonAuto:
		return ansiYellow
	default:
		return ansiRed
	}
}
