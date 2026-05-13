package pty

import (
	"bytes"
	"fmt"
)

// Many TUI libraries (termenv via Bubble Tea, among others) start by emitting
// terminal-capability queries and block their first render until they get
// replies. tea-eyes' pty driver therefore acts as a tiny terminal emulator on
// the master side: it scans the child's output for known queries and writes
// canned responses back.
//
// maxQueryLen is the longest query pattern we recognise; the scanner keeps
// this many trailing bytes around so a query split across two read chunks is
// still matched on the next pass.
const maxQueryLen = 16

// scanForQueryReplies returns the byte sequences that should be written back
// to the pty master in response to any queries present in chunk.
func scanForQueryReplies(chunk []byte, _, height int) [][]byte {
	if height <= 0 {
		height = 24
	}

	var out [][]byte

	// CSI 6n — Device Status Report, request cursor position.
	// Reply with row/col at the bottom-left so layout code that uses it gets a
	// plausible answer.
	if bytes.Contains(chunk, []byte("\x1b[6n")) {
		out = append(out, fmt.Appendf(nil, "\x1b[%d;1R", height))
	}

	// CSI 5n — Device Status Report, "are you OK?". Reply with "I'm OK".
	if bytes.Contains(chunk, []byte("\x1b[5n")) {
		out = append(out, []byte("\x1b[0n"))
	}

	// CSI c / CSI 0c — Primary Device Attributes.
	if bytes.Contains(chunk, []byte("\x1b[c")) || bytes.Contains(chunk, []byte("\x1b[0c")) {
		out = append(out, []byte("\x1b[?1;2c"))
	}

	// CSI >c — Secondary Device Attributes.
	if bytes.Contains(chunk, []byte("\x1b[>c")) || bytes.Contains(chunk, []byte("\x1b[>0c")) {
		out = append(out, []byte("\x1b[>0;10;0c"))
	}

	// OSC 10 ? — foreground color query. Reply with white.
	if bytes.Contains(chunk, []byte("\x1b]10;?")) {
		out = append(out, []byte("\x1b]10;rgb:ffff/ffff/ffff\x1b\\"))
	}

	// OSC 11 ? — background color query. Reply with black so termenv classifies
	// the terminal as a dark theme (a stable, common default).
	if bytes.Contains(chunk, []byte("\x1b]11;?")) {
		out = append(out, []byte("\x1b]11;rgb:0000/0000/0000\x1b\\"))
	}

	return out
}
