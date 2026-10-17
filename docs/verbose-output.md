<!--
File: docs/verbose-output.md
Description: Implementation detail behind --verbose rendering: the
  ASCII-only exception, color rules, and terminal-width truncation.
Date: 2026-10-17
-->

# --verbose output: color and width rules

The box-drawing characters are used unconditionally (this is the one
deliberate exception to pkglist's otherwise ASCII-only output - JSON and
source code stay ASCII-only; this doc only departs from that to quote the
table's actual output in the readme). Color is layered independently on
top: when stdout is a terminal, the border is tan, headers are bold, the
`Version` column is tan, and `Reason` is colored by its value (green
`manual`, yellow `auto`, red `unknown`). Color is skipped automatically
when piped to a file or another program, or when the `NO_COLOR`
environment variable (https://no-color.org) is set; `--color`/`--no-color`
override that detection either way.

When stdout is a terminal, the table is also kept within its actual width
(read via the `TIOCGWINSZ` ioctl) so a long `Package` or `Version` value
never wraps the box onto extra lines - those two columns are shrunk and
elided with `...` as needed, shortest-value columns (`Provider`, `Reason`)
are never touched. Redirecting to a file or another program (`pkglist
--verbose | less -R`, `pkglist --verbose > out.txt`) disables truncation
entirely, since there is no fixed terminal width to wrap against.

See [../internal/render/table.go](../internal/render/table.go) and
[../internal/render/color.go](../internal/render/color.go) for the
implementation.
