// File: internal/provider/go.go
// Description: go provider - lists binaries installed via `go install` by
//   reading the module version embedded in each binary's build info.
// Date: 2026-10-17

package provider

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// GoProvider reads the build info embedded in binaries under the Go
// toolchain's bin directory (GOBIN, or GOPATH/bin when GOBIN is unset).
// `go version -m` only inspects the binary's embedded metadata - it never
// executes the binary itself.
type GoProvider struct{}

func (p *GoProvider) Name() string { return "go" }

func (p *GoProvider) Available() bool {
	_, err := exec.LookPath("go")
	return err == nil
}

func (p *GoProvider) binDir(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "go", "env", "-json", "GOBIN", "GOPATH")
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("go env failed: %w", err)
	}
	var env struct {
		GOBIN  string `json:"GOBIN"`
		GOPATH string `json:"GOPATH"`
	}
	if err := json.Unmarshal(out, &env); err != nil {
		return "", fmt.Errorf("parsing go env json output: %w", err)
	}
	if env.GOBIN != "" {
		return env.GOBIN, nil
	}
	if env.GOPATH == "" {
		return "", fmt.Errorf("neither GOBIN nor GOPATH is set")
	}
	return filepath.Join(env.GOPATH, "bin"), nil
}

func (p *GoProvider) List(ctx context.Context) ([]Package, error) {
	dir, err := p.binDir(ctx)
	if err != nil {
		return nil, err
	}
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil, nil
	}

	cmd := exec.CommandContext(ctx, "go", "version", "-m", dir)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("go version -m failed: %w", err)
	}

	return parseGoVersionM(out), nil
}

// parseGoVersionM parses `go version -m <dir>` output. Each binary starts a
// new block with an unindented "<path>: go<toolchain>" line; the following
// tab-indented lines describe it, and the "mod" line carries the main
// module's version.
func parseGoVersionM(out []byte) []Package {
	var pkgs []Package
	var currentPath, currentVersion string
	flush := func() {
		if currentPath == "" {
			return
		}
		version := currentVersion
		if version == "" {
			version = "unknown"
		}
		pkgs = append(pkgs, Package{Provider: "go", Name: filepath.Base(currentPath), Version: version, InstallReason: InstallReasonManual})
	}

	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		if !strings.HasPrefix(line, "\t") {
			flush()
			currentPath, _, _ = strings.Cut(line, ":")
			currentVersion = ""
			continue
		}
		fields := strings.Split(strings.TrimPrefix(line, "\t"), "\t")
		if len(fields) >= 3 && fields[0] == "mod" {
			currentVersion = fields[2]
		}
	}
	flush()
	return pkgs
}
