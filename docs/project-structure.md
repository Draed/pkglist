<!--
File: docs/project-structure.md
Description: Source layout of the pkglist Go module.
Date: 2026-10-17
-->

# Project structure

```text
cmd/pkglist/           main package: flag parsing and orchestration only
internal/provider/     Package/Provider types, the registry, and one file
                        per provider (apt.go, pipx.go, mise.go, go.go,
                        uv.go, flatpak.go, manual.go)
internal/render/        JSON output and the colored Unicode box table
completions/            bash and zsh completion scripts
```

`internal/` keeps these packages private to this module - nothing outside
pkglist can import them, which is the standard Go way to say "this is
application code, not a library".

Adding another provider (snap, brew, cargo, npm, gem, ...) means
implementing the `Provider` interface in a new file under
[../internal/provider/](../internal/provider/) and registering it in
`All()` in
[../internal/provider/registry.go](../internal/provider/registry.go) - no
other code needs to change, and the bash/zsh completions pick it up
automatically since they query `pkglist --providers` dynamically.
