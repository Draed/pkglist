// File: internal/provider/manual_test.go
// Description: regression tests for the manual provider's --version
//   probing, in particular looksLikeVersion - added after
//   mako-notifications-status (a waybar status script that ignores all
//   arguments and always prints its own JSON status line) was found
//   being reported as that package's "version" by probeVersion.
// Date: 2026-10-25

package provider

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestLooksLikeVersion(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want bool
	}{
		{"plain semver", "1.2.3", true},
		{"v-prefixed semver", "v1.2.3", true},
		{"name plus version", "mytool version 2.0.1", true},
		{"empty", "", false},
		{"json object", `{"text": "x", "class": "empty"}`, false},
		{"json array", `["a", "b"]`, false},
		{"xml-ish", "<version>1.0</version>", false},
		{"quoted string", `"1.0.0"`, false},
		{"usage text", "usage: foo [options]", false},
		{"unknown option text", "unknown option '--version'", false},
		{"unrecognized flag text", "unrecognized flag: --version", false},
		{"control character", "1.0\x0130", false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := looksLikeVersion(tc.in); got != tc.want {
				t.Errorf("looksLikeVersion(%q) = %v, want %v", tc.in, got, tc.want)
			}
		})
	}
}

// writeScript creates an executable shell script in dir and returns its path.
func writeScript(t *testing.T, dir, name, body string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte("#!/bin/sh\n"+body), 0o755); err != nil {
		t.Fatalf("writing script %s: %v", path, err)
	}
	return path
}

func TestProbeVersionIgnoresArgsScript(t *testing.T) {
	dir := t.TempDir()
	// mimics mako-notifications-status: ignores --version entirely and
	// always prints its own JSON status line, exiting 0.
	path := writeScript(t, dir, "notifications-status", `printf '{"text": "x", "class": "empty"}\n'`+"\n")

	got := probeVersion(context.Background(), path)
	if got != "unknown" {
		t.Errorf("probeVersion() = %q, want %q", got, "unknown")
	}
}

func TestProbeVersionRealVersionOutput(t *testing.T) {
	dir := t.TempDir()
	path := writeScript(t, dir, "fake-tool", `echo "fake-tool version 3.1.4"`+"\n")

	got := probeVersion(context.Background(), path)
	if got != "fake-tool version 3.1.4" {
		t.Errorf("probeVersion() = %q, want %q", got, "fake-tool version 3.1.4")
	}
}

func TestProbeVersionUnknownFlagUsageText(t *testing.T) {
	dir := t.TempDir()
	path := writeScript(t, dir, "strict-tool", `echo "unknown option '--version'" >&2; exit 0`+"\n")

	got := probeVersion(context.Background(), path)
	if got != "unknown" {
		t.Errorf("probeVersion() = %q, want %q", got, "unknown")
	}
}
