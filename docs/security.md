<!--
File: docs/security.md
Description: Security properties of pkglist's provider execution model.
Date: 2026-10-17
-->

# Security notes

- Every provider shells out via `exec.CommandContext` with an explicit
  argument list - no string ever passes through `/bin/sh -c`, so there is
  no command-injection surface.
- Each provider call runs under a 15-second timeout so a stuck subprocess
  cannot hang the whole listing. The `manual` provider additionally bounds
  each individual `--version` probe to 2 seconds with no stdin and a
  4096-byte output cap, since it is the only provider that executes
  binaries it did not choose (apt/pipx/mise/go/uv/flatpak are all
  well-known, trusted package manager CLIs).
- A binary dropped into /usr/local/bin or ~/.local/bin is not guaranteed to
  treat `--version` as inert: a personal script may not recognize the flag
  and fall through to its normal interactive behavior (e.g. launching a
  GUI/TUI prompt). Each `--version` probe therefore runs in its own
  session (`Setsid`), so it has no controlling terminal to put into raw
  mode, and on timeout the whole process group is killed, not just the
  direct child, so a GUI/TUI process it spawned cannot survive as an
  orphan.
- `Setsid` only blocks terminal-based side effects. A probed binary that
  falls back to a GUI dialog instead needs no controlling terminal at all -
  it reaches a window purely through the display server socket. Each probe
  therefore also strips `DISPLAY`, `WAYLAND_DISPLAY`, `XAUTHORITY`,
  `XDG_RUNTIME_DIR` and `DBUS_SESSION_BUS_ADDRESS` from its environment, so
  an X11/Wayland client library cannot connect to a compositor and simply
  fails instead of drawing anything.
