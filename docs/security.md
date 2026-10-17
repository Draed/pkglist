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
