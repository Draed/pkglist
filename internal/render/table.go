// File: internal/render/table.go
// Description: Human-readable --verbose output: a rounded Unicode
//   box-drawing table (no gridlines between columns) in the style of the
//   git-who CLI's contributor table. Rows are kept within the terminal's
//   width (when known) by shrinking and eliding the Package/Version
//   columns, so a long value never wraps the box onto extra lines.
//
//   NOTE: this intentionally uses non-ASCII box-drawing characters
//   (U+256D-U+2580 range), a deliberate, explicit exception to this
//   project's default ASCII-only output convention - the user asked for
//   a style closer to a reference screenshot and explicitly said not to
//   worry about rendering compatibility. Every other pkglist output
//   (JSON, source, docs) stays ASCII-only as usual.
// Date: 2026-10-17

package render

import (
	"fmt"
	"io"
	"strings"

	"pkglist/internal/provider"
)

var tableHeaders = []string{"Provider", "Package", "Version", "Reason"}

const columnGap = "  "

// flexibleColumns holds the indices of the columns allowed to shrink when
// the table would otherwise exceed the terminal width. Provider and
// Reason are short, closed sets of values (provider names, install
// reasons) and are never truncated.
var flexibleColumns = []int{1, 2}

// minFlexWidth is the smallest a flexible column is ever shrunk to,
// keeping truncated values at least somewhat readable.
const minFlexWidth = 8

// Rounded single-line box-drawing glyphs.
const (
	glyphHorizontal  = "─" // -
	glyphVertical    = "│" // |
	glyphTopLeft     = "╭" // top-left corner
	glyphTopRight    = "╮" // top-right corner
	glyphBottomLeft  = "╰" // bottom-left corner
	glyphBottomRight = "╯" // bottom-right corner
	glyphTeeLeft     = "├" // left tee (header separator)
	glyphTeeRight    = "┤" // right tee (header separator)
)

func tableRowValues(pkg provider.Package) []string {
	return []string{pkg.Provider, pkg.Name, pkg.Version, pkg.InstallReason}
}

// Table writes pkgs to w as a rounded box-drawing table. When color is
// true, the border is tan, headers are bold, the Version column is tan,
// and the Reason column is colored by its value (green manual, yellow
// auto, red unknown). maxWidth caps the total row width (as reported by
// TerminalWidth); pass 0 to disable truncation entirely (e.g. output is
// not an interactive terminal).
func Table(w io.Writer, pkgs []provider.Package, color bool, maxWidth int) {
	widths := fitWidths(columnWidths(pkgs), maxWidth)
	top := colorize(color, ansiTan, horizontalLine(widths, glyphTopLeft, glyphTopRight))
	mid := colorize(color, ansiTan, horizontalLine(widths, glyphTeeLeft, glyphTeeRight))
	bottom := colorize(color, ansiTan, horizontalLine(widths, glyphBottomLeft, glyphBottomRight))

	fmt.Fprintln(w, top)
	fmt.Fprintln(w, tableHeaderRow(widths, color))
	fmt.Fprintln(w, mid)
	for _, pkg := range pkgs {
		fmt.Fprintln(w, tableDataRow(pkg, widths, color))
	}
	fmt.Fprintln(w, bottom)
	fmt.Fprintf(w, "Total: %d package(s)\n", len(pkgs))
}

func columnWidths(pkgs []provider.Package) []int {
	widths := make([]int, len(tableHeaders))
	for i, h := range tableHeaders {
		widths[i] = len(h)
	}
	for _, pkg := range pkgs {
		for i, v := range tableRowValues(pkg) {
			if l := len(v); l > widths[i] {
				widths[i] = l
			}
		}
	}
	return widths
}

// rowWidth is the total on-screen width a table using these column widths
// renders at: the two single-width border glyphs, one padding space on
// each side, the columns themselves, and a columnGap between each pair.
func rowWidth(widths []int) int {
	total := 2 + 2 // border glyphs + inner padding spaces
	for i, w := range widths {
		total += w
		if i > 0 {
			total += len(columnGap)
		}
	}
	return total
}

// fitWidths shrinks the flexible columns (Package, Version) just enough
// that the table fits within maxWidth, splitting the available space
// between them so neither is starved while the other has room to spare.
// maxWidth <= 0 means "no limit": natural widths are returned unchanged.
func fitWidths(natural []int, maxWidth int) []int {
	if maxWidth <= 0 || rowWidth(natural) <= maxWidth {
		return natural
	}

	widths := append([]int(nil), natural...)
	overflow := rowWidth(natural) - maxWidth

	for overflow > 0 {
		// Shrink whichever flexible column currently has the most room
		// to give, one character at a time, until the row fits or every
		// flexible column has hit its floor.
		target := -1
		for _, i := range flexibleColumns {
			if widths[i] <= minFlexWidth {
				continue
			}
			if target == -1 || widths[i] > widths[target] {
				target = i
			}
		}
		if target == -1 {
			break // every flexible column is already at its floor
		}
		widths[target]--
		overflow--
	}
	return widths
}

// truncate shortens s to at most width characters, replacing the cut
// point with "..." when there is room for it.
func truncate(s string, width int) string {
	if len(s) <= width {
		return s
	}
	if width <= 3 {
		if width <= 0 {
			return ""
		}
		return s[:width]
	}
	return s[:width-3] + "..."
}

// horizontalLine spans the full row width: 1 leading space + columns
// joined by columnGap + 1 trailing space, matching tableRow's layout
// exactly so the border lines up with the vertical edges of every row.
func horizontalLine(widths []int, left, right string) string {
	inner := 2 // leading + trailing space inside the two vertical edges
	for i, w := range widths {
		inner += w
		if i > 0 {
			inner += len(columnGap)
		}
	}
	return left + strings.Repeat(glyphHorizontal, inner) + right
}

func tableHeaderRow(widths []int, color bool) string {
	var b strings.Builder
	b.WriteString(glyphVertical + " ")
	for i, h := range tableHeaders {
		if i > 0 {
			b.WriteString(columnGap)
		}
		b.WriteString(colorize(color, ansiBold, padRight(h, widths[i])))
	}
	b.WriteString(" " + glyphVertical)
	return b.String()
}

func tableDataRow(pkg provider.Package, widths []int, color bool) string {
	values := tableRowValues(pkg)
	values[1] = padRight(truncate(values[1], widths[1]), widths[1])
	values[2] = colorize(color, ansiTan, padRight(truncate(values[2], widths[2]), widths[2]))
	values[3] = colorize(color, reasonColor(pkg.InstallReason), padRight(values[3], widths[3]))
	values[0] = padRight(values[0], widths[0])

	var b strings.Builder
	b.WriteString(glyphVertical + " ")
	b.WriteString(strings.Join(values, columnGap))
	b.WriteString(" " + glyphVertical)
	return b.String()
}

func padRight(s string, width int) string {
	if pad := width - len(s); pad > 0 {
		return s + strings.Repeat(" ", pad)
	}
	return s
}
