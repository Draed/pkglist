<!--
File: docs/manual-provider.md
Description: Caveats and known tradeoffs of the "manual" provider (see
  ../internal/provider/manual.go).
Date: 2026-10-17
-->

# manual provider caveats

- Version is obtained by running `<binary> --version` under a 2-second
  timeout with no stdin and capped output, then taking its first line -
  but only when the command exits successfully. A non-zero exit (e.g.
  Teleport's `tsh`/`tctl`, which reject `--version` and use a `version`
  subcommand instead) reports version `"unknown"` rather than showing the
  tool's error/usage text, since that text is never a real version. One
  known tradeoff: a binary that happens to print a real version line but
  then exits non-zero for an unrelated reason (missing required
  arguments, for instance) is still reported as `"unknown"` - exit status
  is treated as the one reliable signal rather than trying to pattern-
  match error text, which would be far more fragile.
- To avoid double-reporting, candidates whose resolved real path falls
  under a known shim directory (pipx venvs, mise installs, uv tool venvs,
  cargo bin, a local `node_modules/.bin`) are skipped, since another
  provider already reports them. This list is best-effort, not
  exhaustive - see `manualExcludeSubstrings` in
  [../internal/provider/manual.go](../internal/provider/manual.go).
