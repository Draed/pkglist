// File: internal/provider/apt.go
// Description: apt/dpkg provider - lists packages installed through the
//   Debian package manager via dpkg-query.
// Date: 2026-10-17

package provider

import (
	"bufio"
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// AptProvider reads installed packages from dpkg's package database.
// dpkg-query is used directly (never a shell string) to avoid any command
// injection surface, and runs under the caller-supplied context so it
// cannot hang the whole collection.
type AptProvider struct{}

func (p *AptProvider) Name() string { return "apt" }

func (p *AptProvider) Available() bool {
	_, err := exec.LookPath("dpkg-query")
	return err == nil
}

func (p *AptProvider) List(ctx context.Context) ([]Package, error) {
	cmd := exec.CommandContext(ctx, "dpkg-query", "-W", "-f=${Package}\t${Version}\t${Status}\n")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("dpkg-query failed: %w", err)
	}

	// apt-mark failures degrade to "unknown" reasons rather than failing
	// the whole provider: the package list itself is still valid without it.
	autoSet, _ := aptMarkSet(ctx, "showauto")
	manualSet, _ := aptMarkSet(ctx, "showmanual")

	var pkgs []Package
	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		fields := strings.SplitN(line, "\t", 3)
		if len(fields) != 3 {
			continue
		}
		name, version, status := fields[0], fields[1], fields[2]

		// Status is "<want> <flag> <status>", e.g. "install ok installed".
		// Packages left in a removed-but-not-purged state ("rc") are
		// skipped since they are not actually installed.
		statusWords := strings.Fields(status)
		if len(statusWords) == 0 || statusWords[len(statusWords)-1] != "installed" {
			continue
		}

		reason := InstallReasonUnknown
		switch {
		case autoSet[name]:
			reason = InstallReasonAuto
		case manualSet[name]:
			reason = InstallReasonManual
		}

		pkgs = append(pkgs, Package{Provider: p.Name(), Name: name, Version: version, InstallReason: reason})
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("reading dpkg-query output: %w", err)
	}
	return pkgs, nil
}

// aptMarkSet runs `apt-mark <which>` (expected to be "showauto" or
// "showmanual") and returns the set of package names it printed. Foreign-
// architecture packages are reported by apt-mark as "name:arch"; the
// ":arch" suffix is stripped so lookups by bare package name still match.
func aptMarkSet(ctx context.Context, which string) (map[string]bool, error) {
	cmd := exec.CommandContext(ctx, "apt-mark", which)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("apt-mark %s failed: %w", which, err)
	}

	set := make(map[string]bool)
	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	for scanner.Scan() {
		name, _, _ := strings.Cut(strings.TrimSpace(scanner.Text()), ":")
		if name != "" {
			set[name] = true
		}
	}
	return set, scanner.Err()
}
