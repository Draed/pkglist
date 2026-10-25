<!--
File: readme.md
Description: Usage and build instructions for pkglist, a cross-provider
  installed package listing tool for Debian-based systems.
Date: 2026-10-17
-->

# pkglist

Lists every installed package across all detected package providers on a
Debian-based system, with name and version, in a single machine-readable
stream.

## Documentation

- [Installation and shell completion](docs/installation.md)
- [Project structure](docs/project-structure.md)
- [manual provider caveats](docs/manual-provider.md)
- [--verbose color and width rules](docs/verbose-output.md)
- [Security notes](docs/security.md)

## Supported providers

- `apt` / `dpkg` (core Debian packages)
- `pipx` (Python CLI apps, including packages injected into their venvs)
- `mise` (language runtimes/tools)
- `go` (binaries installed via `go install`, versioned from the module
  info embedded in each binary - no execution of the binary involved)
- `uv` (CLI tools installed via `uv tool install`)
- `flatpak` (installed apps; runtimes/SDKs are intentionally excluded)
- `manual` (executables dropped directly into `/usr/local/bin` and
  `~/.local/bin` that no other provider above already tracks - see
  [manual provider caveats](docs/manual-provider.md))

A provider is only queried if its binary is found on `PATH`
(`dpkg-query`, `pipx`, `mise`, `go`, `uv`, `flatpak`; `manual` instead
checks that at least one of its scanned directories exists). Run
`pkglist --list-providers` to see what was detected on the current
machine.

## Build

```sh
go build -o pkglist ./cmd/pkglist
```

See [docs/installation.md](docs/installation.md) for installing the
built binary and its shell completions.

## Usage

```text
pkglist [flags]

Flags:
  -v, --verbose          human-readable box-drawing table instead of JSON
  -p, --provider NAMES   comma-separated providers to query (default: all detected)
      --list-providers   list known providers and whether they are detected, then exit
      --providers        print only the detected provider names, one per line
      --color            force colored --verbose output, even when stdout is not a terminal
      --no-color         disable colored --verbose output
  -V, --version          print version information and exit
  -h, --help             show usage text and exit
```

### Machine-readable (default)

```sh
pkglist
```

Prints a JSON array of objects on stdout:

```json
[
  {
    "provider": "apt",
    "name": "bash",
    "version": "5.2.21-2+deb12u1",
    "install_reason": "manual"
  }
]
```

`install_reason` is `"manual"` (explicitly requested), `"auto"` (pulled in
only as a dependency), or `"unknown"`. Only `apt` can actually tell these
apart - it reads the same data as `apt-mark showmanual`/`apt-mark
showauto`. Every other provider here only ever installs what the user
explicitly named (there is no dependency-of-a-dependency concept in pipx,
uv tool, go install, flatpak app installs, mise, or a binary dropped into
`~/.local/bin`), so they always report `"manual"`.

Pipe it into `jq`, Python, or any script:

```sh
pkglist | jq '.[] | select(.provider == "apt")'
pkglist --provider=pipx | jq -r '.[].name'
pkglist | jq '.[] | select(.install_reason == "manual" and .provider == "apt")'
```

### Human-readable

```sh
pkglist --verbose
```

Prints a rounded Unicode box-drawing table (no gridlines between columns),
styled after the `git-who` CLI's contributor table:

```text
╭──────────────────────────────────────────────╮
│ Provider  Package    Version          Reason │
├──────────────────────────────────────────────┤
│ apt       bash       5.2.37-2+b9      manual │
│ apt       libc6-dev  2.41-12+deb13u3  auto   │
│ pipx      ansible    14.2.0           manual │
╰──────────────────────────────────────────────╯
Total: 3 package(s)
```

See [--verbose color and width rules](docs/verbose-output.md) for exactly
how color and terminal-width truncation are applied.

### Version

```sh
pkglist --version
```

Prints `pkglist <tag>`, e.g. `pkglist v1.0.2`. The version string is
injected at build time via `-ldflags "-X main.version=<tag>"` (see
`.goreleaser.yaml`); a plain `go build -o pkglist ./cmd/pkglist` with no
ldflags prints `pkglist dev` instead.

### Exit codes

- `0` - success
- `1` - usage error (e.g. unknown `--provider` name)
- `2` - one or more providers failed to run; packages from the other
  providers are still printed. Warnings go to stderr.
