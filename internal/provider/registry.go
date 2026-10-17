// File: internal/provider/registry.go
// Description: The provider registry - lists every known provider,
//   resolves the --provider filter against them, queries them, and sorts
//   the merged results.
// Date: 2026-10-17

package provider

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"
)

// Timeout bounds how long a single provider's command may run, so one
// stuck subprocess cannot hang the whole listing.
const Timeout = 15 * time.Second

// All returns every provider pkglist knows about, regardless of whether it
// is actually available on this system.
func All() []Provider {
	return []Provider{
		&AptProvider{},
		&PipxProvider{},
		&MiseProvider{},
		&GoProvider{},
		&UvProvider{},
		&FlatpakProvider{},
		&ManualProvider{},
	}
}

// Select resolves the --provider filter (comma-separated provider names)
// against the known providers. An empty filter selects every available
// provider.
func Select(providers []Provider, filter string) ([]Provider, error) {
	if filter == "" {
		var out []Provider
		for _, p := range providers {
			if p.Available() {
				out = append(out, p)
			}
		}
		return out, nil
	}

	known := make(map[string]Provider, len(providers))
	for _, p := range providers {
		known[p.Name()] = p
	}

	var out []Provider
	for _, raw := range strings.Split(filter, ",") {
		name := strings.ToLower(strings.TrimSpace(raw))
		if name == "" {
			continue
		}
		p, ok := known[name]
		if !ok {
			return nil, fmt.Errorf("unknown provider %q (known: %s)", name, strings.Join(names(providers), ", "))
		}
		if !p.Available() {
			fmt.Fprintf(os.Stderr, "warning: provider %q requested but not available on this system\n", name)
			continue
		}
		out = append(out, p)
	}
	return out, nil
}

func names(providers []Provider) []string {
	out := make([]string, 0, len(providers))
	for _, p := range providers {
		out = append(out, p.Name())
	}
	return out
}

// Collect queries every given provider and merges the results. A provider
// error is reported on stderr but does not stop the other providers from
// being queried; the returned bool reports whether any provider failed.
func Collect(providers []Provider) ([]Package, bool) {
	var all []Package
	hadErrors := false
	for _, p := range providers {
		ctx, cancel := context.WithTimeout(context.Background(), Timeout)
		pkgs, err := p.List(ctx)
		cancel()
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: provider %q: %v\n", p.Name(), err)
			hadErrors = true
			continue
		}
		all = append(all, pkgs...)
	}
	return all, hadErrors
}

// Sort orders packages by provider, then by name (case-insensitive).
func Sort(pkgs []Package) {
	sort.Slice(pkgs, func(i, j int) bool {
		if pkgs[i].Provider != pkgs[j].Provider {
			return pkgs[i].Provider < pkgs[j].Provider
		}
		return strings.ToLower(pkgs[i].Name) < strings.ToLower(pkgs[j].Name)
	})
}

// PrintStatus prints each provider's name and whether it was detected on
// this system.
func PrintStatus(providers []Provider) {
	for _, p := range providers {
		status := "not available"
		if p.Available() {
			status = "available"
		}
		fmt.Printf("%-10s %s\n", p.Name(), status)
	}
}
