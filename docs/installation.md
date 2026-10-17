<!--
File: docs/installation.md
Description: Install and shell-completion setup for pkglist.
Date: 2026-10-17
-->

# Installation

## Install (user-local, no sudo)

```sh
cp pkglist "$HOME/.local/bin/pkglist"
chmod 755 "$HOME/.local/bin/pkglist"
```

`~/.local/bin` must be on `PATH` (it is, by default, on most Debian
desktop setups).

## Shell completion

`-p`/`--provider` tab-completes against the providers actually detected on
the current machine (via `pkglist --providers`), including comma-separated
lists (e.g. `--provider=apt,pi<TAB>` -> `pipx`).

### Bash (requires the `bash-completion` package)

```sh
mkdir -p "$HOME/.local/share/bash-completion/completions"
cp completions/pkglist.bash "$HOME/.local/share/bash-completion/completions/pkglist"
```

Picked up automatically in new shells (no sourcing needed, no sudo).

### Zsh

```sh
mkdir -p "$HOME/.zsh/completions"
cp completions/_pkglist "$HOME/.zsh/completions/_pkglist"
```

Then make sure that directory is in `fpath` before `compinit` runs, e.g. in
`~/.zshrc`:

```sh
fpath=("$HOME/.zsh/completions" $fpath)
autoload -U compinit && compinit
```
