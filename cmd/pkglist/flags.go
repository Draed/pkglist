// File: cmd/pkglist/flags.go
// Description: Command-line flag definitions for pkglist.
// Date: 2026-10-17

package main

import (
	"flag"
	"os"
)

func parseFlags(verbose *bool, providerFlag *string, listProviders, printProviderID, help, noColor, forceColor *bool) {
	flag.BoolVar(verbose, "verbose", false, "human-readable ASCII table instead of JSON")
	flag.BoolVar(verbose, "v", false, "shorthand for --verbose")
	flag.StringVar(providerFlag, "provider", "", "comma-separated providers to query (default: all detected)")
	flag.StringVar(providerFlag, "p", "", "shorthand for --provider")
	flag.BoolVar(listProviders, "list-providers", false, "list known providers and whether they are detected, then exit")
	flag.BoolVar(printProviderID, "providers", false, "print only the detected provider names, one per line (for shell completion/scripting), then exit")
	flag.BoolVar(noColor, "no-color", false, "disable colored --verbose output")
	flag.BoolVar(forceColor, "color", false, "force colored --verbose output even when stdout is not a terminal")
	flag.BoolVar(help, "help", false, "show usage and exit")
	flag.BoolVar(help, "h", false, "shorthand for --help")
	flag.Usage = func() { printUsage(os.Stderr) }
	flag.Parse()
}
