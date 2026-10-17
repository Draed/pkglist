// File: internal/provider/uv.go
// Description: uv provider - lists CLI tools installed via `uv tool
//   install` (uv's pipx-equivalent).
// Date: 2026-10-17

package provider

import (
	"bufio"
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// UvProvider reads installed tools from `uv tool list`. uv has no --format
// json for this subcommand, so the plain text output is parsed instead:
// an unindented "<name> v<version>" header line per tool, followed by
// "- <executable>" lines that are not package entries and are skipped.
type UvProvider struct{}

func (p *UvProvider) Name() string { return "uv" }

func (p *UvProvider) Available() bool {
	_, err := exec.LookPath("uv")
	return err == nil
}

func (p *UvProvider) List(ctx context.Context) ([]Package, error) {
	cmd := exec.CommandContext(ctx, "uv", "tool", "list")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("uv tool list failed: %w", err)
	}

	var pkgs []Package
	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	for scanner.Scan() {
		line := strings.TrimRight(scanner.Text(), " \t")
		if line == "" || line == "No tools installed" || strings.HasPrefix(line, "- ") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		name := fields[0]
		version := "unknown"
		if len(fields) >= 2 {
			version = strings.TrimPrefix(fields[1], "v")
		}
		pkgs = append(pkgs, Package{Provider: p.Name(), Name: name, Version: version, InstallReason: InstallReasonManual})
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("reading uv tool list output: %w", err)
	}
	return pkgs, nil
}
