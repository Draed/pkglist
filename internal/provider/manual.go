// File: internal/provider/manual.go
// Description: manual provider - lists executables dropped directly into
//   /usr/local/bin and ~/.local/bin that are not already tracked by another
//   provider (apt never installs into /usr/local by Debian policy; the
//   pipx/mise/uv exclusions below cover the known shim locations under
//   ~/.local). Versions are best-effort, obtained by running each
//   executable with --version under a short bounded timeout.
// Date: 2026-10-17

package provider

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// perBinaryVersionTimeout bounds a single --version probe so one
// unresponsive binary cannot stall the manual provider.
const perBinaryVersionTimeout = 2 * time.Second

// maxVersionProbeOutput caps how much of a binary's --version output is
// kept in memory; anything beyond this is discarded, not buffered.
const maxVersionProbeOutput = 4096

// manualExcludeSubstrings are path fragments of directories already owned
// by another provider's shims/symlinks (pipx, mise, uv tool, npm). A
// manual-scan candidate resolving into one of these is skipped so it is
// not double-reported under both providers. This list is best-effort, not
// exhaustive.
var manualExcludeSubstrings = []string{
	"/pipx/venvs/",
	"/.local/share/mise/",
	"/.local/share/uv/tools/",
	"/.cargo/bin/",
	"/node_modules/.bin/",
}

// ManualProvider scans fixed, well-known "drop a binary here" directories
// for executables that no package manager is tracking.
type ManualProvider struct{}

func (p *ManualProvider) Name() string { return "manual" }

func (p *ManualProvider) dirs() []string {
	var dirs []string
	dirs = append(dirs, "/usr/local/bin")
	if home, err := os.UserHomeDir(); err == nil {
		dirs = append(dirs, filepath.Join(home, ".local", "bin"))
	}
	return dirs
}

func (p *ManualProvider) Available() bool {
	for _, dir := range p.dirs() {
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			return true
		}
	}
	return false
}

func (p *ManualProvider) List(ctx context.Context) ([]Package, error) {
	var pkgs []Package

	for _, dir := range p.dirs() {
		entries, err := os.ReadDir(dir)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, fmt.Errorf("reading %s: %w", dir, err)
		}

		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}

			rawPath := filepath.Join(dir, entry.Name())
			realPath, err := filepath.EvalSymlinks(rawPath)
			if err != nil {
				continue
			}
			if isManagedPath(realPath) {
				continue
			}

			info, err := os.Stat(realPath)
			if err != nil || !info.Mode().IsRegular() || info.Mode().Perm()&0111 == 0 {
				continue
			}

			pkgs = append(pkgs, Package{
				Provider:      p.Name(),
				Name:          entry.Name(),
				Version:       probeVersion(ctx, realPath),
				InstallReason: InstallReasonManual,
			})
		}
	}

	return pkgs, nil
}

func isManagedPath(path string) bool {
	for _, substr := range manualExcludeSubstrings {
		if strings.Contains(path, substr) {
			return true
		}
	}
	return false
}

// limitedBuffer caps how many bytes it retains, silently discarding the
// rest while still reporting a successful write so the copying goroutines
// started by exec.Cmd never see a write error.
type limitedBuffer struct {
	buf bytes.Buffer
	max int
}

func (b *limitedBuffer) Write(p []byte) (int, error) {
	remaining := b.max - b.buf.Len()
	if remaining > 0 {
		if remaining > len(p) {
			remaining = len(p)
		}
		b.buf.Write(p[:remaining])
	}
	return len(p), nil
}

// probeVersion runs "<path> --version" under a short timeout with no
// stdin and bounded output, and returns the first line of its output.
// Any failure (timeout, non-zero exit, missing --version support) falls
// back to "unknown" rather than propagating an error - a single
// uncooperative binary should never fail the whole provider, and its
// error/usage text (e.g. "unknown long flag '--version'") is never a
// package version, so it must never be shown as one.
func probeVersion(parent context.Context, path string) string {
	ctx, cancel := context.WithTimeout(parent, perBinaryVersionTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, path, "--version")
	out := &limitedBuffer{max: maxVersionProbeOutput}
	cmd.Stdout = out
	cmd.Stderr = out
	cmd.Stdin = nil

	if err := cmd.Run(); err != nil {
		return "unknown"
	}

	firstLine, _, _ := strings.Cut(out.buf.String(), "\n")
	firstLine = strings.TrimSpace(firstLine)
	if firstLine == "" {
		return "unknown"
	}
	const maxVersionLen = 200
	if len(firstLine) > maxVersionLen {
		firstLine = firstLine[:maxVersionLen]
	}
	return firstLine
}
