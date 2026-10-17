// File: internal/provider/pipx.go
// Description: pipx provider - lists Python CLI applications installed via
//   pipx, including packages injected into their venvs.
// Date: 2026-10-17

package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
)

// PipxProvider reads installed packages from `pipx list --json`.
type PipxProvider struct{}

func (p *PipxProvider) Name() string { return "pipx" }

func (p *PipxProvider) Available() bool {
	_, err := exec.LookPath("pipx")
	return err == nil
}

type pipxPackageRef struct {
	Package        string `json:"package"`
	PackageVersion string `json:"package_version"`
}

type pipxListOutput struct {
	Venvs map[string]struct {
		Metadata struct {
			MainPackage      pipxPackageRef            `json:"main_package"`
			InjectedPackages map[string]pipxPackageRef `json:"injected_packages"`
		} `json:"metadata"`
	} `json:"venvs"`
}

func (p *PipxProvider) List(ctx context.Context) ([]Package, error) {
	cmd := exec.CommandContext(ctx, "pipx", "list", "--json")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("pipx list --json failed: %w", err)
	}

	var parsed pipxListOutput
	if err := json.Unmarshal(out, &parsed); err != nil {
		return nil, fmt.Errorf("parsing pipx json output: %w", err)
	}

	var pkgs []Package
	for _, venv := range parsed.Venvs {
		if mp := venv.Metadata.MainPackage; mp.Package != "" {
			pkgs = append(pkgs, Package{Provider: p.Name(), Name: mp.Package, Version: mp.PackageVersion, InstallReason: InstallReasonManual})
		}
		for _, ip := range venv.Metadata.InjectedPackages {
			if ip.Package != "" {
				pkgs = append(pkgs, Package{Provider: p.Name(), Name: ip.Package, Version: ip.PackageVersion, InstallReason: InstallReasonManual})
			}
		}
	}
	return pkgs, nil
}
