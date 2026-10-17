// File: internal/provider/mise.go
// Description: mise provider - lists language runtimes/tools installed and
//   activated via mise (asdf-style version manager).
// Date: 2026-10-17

package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
)

// MiseProvider reads installed tool versions from `mise ls --json`.
//
// mise's JSON schema has changed across releases, so List tries the
// current documented schema (a map of tool name to a list of version
// entries) first and falls back to a flat array schema before giving up.
type MiseProvider struct{}

func (p *MiseProvider) Name() string { return "mise" }

func (p *MiseProvider) Available() bool {
	_, err := exec.LookPath("mise")
	return err == nil
}

type miseVersionEntry struct {
	Version string `json:"version"`
}

type miseFlatEntry struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

func (p *MiseProvider) List(ctx context.Context) ([]Package, error) {
	cmd := exec.CommandContext(ctx, "mise", "ls", "--json")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("mise ls --json failed: %w", err)
	}

	var byTool map[string][]miseVersionEntry
	if err := json.Unmarshal(out, &byTool); err == nil {
		var pkgs []Package
		for tool, versions := range byTool {
			for _, v := range versions {
				if v.Version == "" {
					continue
				}
				pkgs = append(pkgs, Package{Provider: p.Name(), Name: tool, Version: v.Version, InstallReason: InstallReasonManual})
			}
		}
		return pkgs, nil
	}

	var flat []miseFlatEntry
	if err := json.Unmarshal(out, &flat); err == nil {
		var pkgs []Package
		for _, e := range flat {
			if e.Version == "" {
				continue
			}
			pkgs = append(pkgs, Package{Provider: p.Name(), Name: e.Name, Version: e.Version, InstallReason: InstallReasonManual})
		}
		return pkgs, nil
	}

	return nil, fmt.Errorf("parsing mise json output: unrecognized schema")
}
