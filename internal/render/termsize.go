// File: internal/render/termsize.go
// Description: Terminal width detection (Linux TIOCGWINSZ ioctl) used to
//   keep --verbose table rows from wrapping mid-line in narrow terminals.
// Date: 2026-10-17

package render

import (
	"os"
	"syscall"
	"unsafe"
)

type winsize struct {
	Row, Col, Xpixel, Ypixel uint16
}

// TerminalWidth reports the current column width of stdout's controlling
// terminal. The second return value is false when stdout is not an actual
// terminal (redirected to a file or pipe) or the width could not be
// determined - callers should skip width-based truncation in that case,
// since there is no fixed terminal width to wrap against.
func TerminalWidth() (int, bool) {
	if !isTerminal(os.Stdout) {
		return 0, false
	}

	var ws winsize
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, os.Stdout.Fd(), uintptr(syscall.TIOCGWINSZ), uintptr(unsafe.Pointer(&ws)))
	if errno != 0 || ws.Col == 0 {
		return 0, false
	}
	return int(ws.Col), true
}
