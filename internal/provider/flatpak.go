// File: internal/provider/flatpak.go
// Description: flatpak provider - lists installed Flatpak applications
//   (runtimes/SDKs are intentionally excluded; see Available/List).
// Date: 2026-10-17

package provider

import (
	"bufio"
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// FlatpakProvider reads installed apps via `flatpak list --app
// --columns=application,version`. The --columns flag is flatpak's
// documented machine-readable mode: one row per line, fields separated by
// a single tab character.
type FlatpakProvider struct{}

func (p *FlatpakProvider) Name() string { return "flatpak" }

func (p *FlatpakProvider) Available() bool {
	_, err := exec.LookPath("flatpak")
	return err == nil
}

func (p *FlatpakProvider) List(ctx context.Context) ([]Package, error) {
	cmd := exec.CommandContext(ctx, "flatpak", "list", "--app", "--columns=application,version")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("flatpak list failed: %w", err)
	}

	var pkgs []Package
	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		fields := strings.Split(line, "\t")
		if len(fields) == 0 || fields[0] == "" {
			continue
		}
		version := "unknown"
		if len(fields) >= 2 && fields[1] != "" && fields[1] != "-" {
			version = fields[1]
		}
		pkgs = append(pkgs, Package{Provider: p.Name(), Name: fields[0], Version: version, InstallReason: InstallReasonManual})
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("reading flatpak list output: %w", err)
	}
	return pkgs, nil
}
