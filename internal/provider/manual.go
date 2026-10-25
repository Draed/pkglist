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
	"syscall"
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
//
// Binaries dropped into /usr/local/bin or ~/.local/bin are not guaranteed
// to treat --version as a no-op flag: a personal script or wrapper may not
// recognize it and fall through to its normal interactive behavior (e.g.
// launching a GUI/TUI prompt). The probe therefore runs the candidate in
// its own session (Setsid) so it has no controlling terminal to manipulate,
// and on timeout kills the whole process group, not just the direct child,
// so a GUI/TUI process it spawned cannot survive as an orphan.
func probeVersion(parent context.Context, path string) string {
	ctx, cancel := context.WithTimeout(parent, perBinaryVersionTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, path, "--version")
	out := &limitedBuffer{max: maxVersionProbeOutput}
	cmd.Stdout = out
	cmd.Stderr = out
	cmd.Stdin = nil
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	cmd.Cancel = func() error {
		return syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
	}

	if err := cmd.Run(); err != nil {
		return "unknown"
	}

	firstLine, _, _ := strings.Cut(out.buf.String(), "\n")
	firstLine = strings.TrimSpace(firstLine)
	if !looksLikeVersion(firstLine) {
		return "unknown"
	}
	const maxVersionLen = 200
	if len(firstLine) > maxVersionLen {
		firstLine = firstLine[:maxVersionLen]
	}
	return firstLine
}

// looksLikeVersion reports whether a --version probe's first output line is
// plausible as an actual version string, rather than some other kind of
// output produced by a script/binary that does not understand --version at
// all and ignores it, falling through to its normal behavior instead of
// erroring (e.g. mako-notifications-status, a waybar status script dropped
// into ~/.local/bin that does not parse its arguments and always prints its
// own JSON status line - {"text": "...", "class": "empty", ...} - which
// exits 0 just like a real --version flag would, so the exit-code check
// above alone cannot catch it). A real --version line is never JSON/XML and
// never a "this flag is not recognized" usage/error message.
func looksLikeVersion(s string) bool {
	if s == "" {
		return false
	}
	switch s[0] {
	case '{', '[', '<', '"', '\'':
		return false
	}
	lower := strings.ToLower(s)
	for _, marker := range []string{
		"usage:", "usage ",
		"unknown option", "unrecognized option", "invalid option",
		"unknown flag", "unrecognized flag",
	} {
		if strings.Contains(lower, marker) {
			return false
		}
	}
	for _, r := range s {
		if r < 0x20 && r != '\t' {
			return false
		}
	}
	return true
}
