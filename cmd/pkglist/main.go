// File: cmd/pkglist/main.go
// Description: pkglist entry point - a cross-provider wrapper that lists
//   every installed package (apt/dpkg, pipx, mise, go, uv, flatpak,
//   manual) with its version. Default output is machine-readable JSON;
//   --verbose prints a colored Unicode box-drawing table for interactive
//   use (see internal/render/table.go for the ASCII-only exception note).
// Date: 2026-10-17

package main

import (
	"fmt"
	"io"
	"os"

	"pkglist/internal/provider"
	"pkglist/internal/render"
)

func main() {
	var (
		verbose         bool
		providerFlag    string
		listProviders   bool
		printProviderID bool
		help            bool
		noColor         bool
		forceColor      bool
	)

	parseFlags(&verbose, &providerFlag, &listProviders, &printProviderID, &help, &noColor, &forceColor)

	if help {
		printUsage(os.Stdout)
		return
	}

	providers := provider.All()

	if listProviders {
		provider.PrintStatus(providers)
		return
	}

	if printProviderID {
		for _, p := range providers {
			if p.Available() {
				fmt.Println(p.Name())
			}
		}
		return
	}

	selected, err := provider.Select(providers, providerFlag)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}

	pkgs, hadErrors := provider.Collect(selected)
	provider.Sort(pkgs)

	if verbose {
		width, _ := render.TerminalWidth()
		render.Table(os.Stdout, pkgs, render.UseColor(noColor, forceColor), width)
	} else if err := render.JSON(os.Stdout, pkgs); err != nil {
		fmt.Fprintln(os.Stderr, "error: encoding JSON output:", err)
		os.Exit(1)
	}

	if hadErrors {
		os.Exit(2)
	}
}

func printUsage(w io.Writer) {
	fmt.Fprint(w, `pkglist - list installed packages across every detected package provider

Usage:
  pkglist [flags]

Flags:
  -v, --verbose          human-readable box-drawing table instead of JSON
  -p, --provider NAMES   comma-separated providers to query (default: all detected)
      --list-providers   list known providers and whether they are detected, then exit
      --providers        print only the detected provider names, one per line
                          (used by shell completion/scripting), then exit
      --color            force colored --verbose output, even when stdout is not a terminal
      --no-color         disable colored --verbose output
  -h, --help             show this usage text and exit

Shell completion:
  See completions/pkglist.bash and completions/_pkglist for --provider
  autocompletion limited to the providers actually detected on this system.

Output:
  Default output is a JSON array of
  {"provider","name","version","install_reason"} objects on stdout,
  suitable for piping into jq/python/scripts. install_reason is "manual",
  "auto" (apt only, from apt-mark), or "unknown". Warnings about a
  provider that failed to run go to stderr and do not interrupt the other
  providers; the process then exits with status 2 instead of 0.

  --verbose output is colored automatically when stdout is a terminal and
  the NO_COLOR environment variable is unset (https://no-color.org);
  --color/--no-color override that detection.

Known providers: apt, pipx, mise, go, uv, flatpak, manual

Examples:
  pkglist
  pkglist --verbose
  pkglist --provider=apt,pipx
  pkglist --verbose | less -R
  pkglist | jq '.[] | select(.provider == "apt")'
`)
}
