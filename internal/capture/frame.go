// Package capture turns raw pty output bytes into a clean text representation
// of the visible cell grid by feeding them through a virtual terminal
// (hinshun/vt10x).
package capture

import (
	"bytes"
	"strings"

	"github.com/hinshun/vt10x"
)

// Frame is a snapshot of the virtual terminal after replaying raw pty bytes.
type Frame struct {
	Width  int
	Height int
	Text   string // newline-separated lines, no trailing newline
}

// RenderFrame replays raw onto a virtual terminal of width×height and returns
// the visible cell grid. Trailing whitespace is trimmed from each line for
// readability.
//
// If stripANSI is false, raw is returned verbatim (with ANSI escape sequences
// intact). This is useful when the caller needs to inspect color/SGR data; the
// trade-off is that the result is no longer a clean grid.
func RenderFrame(raw []byte, width, height int, stripANSI bool) (Frame, error) {
	if width <= 0 {
		width = 80
	}
	if height <= 0 {
		height = 24
	}

	if !stripANSI {
		return Frame{
			Width:  width,
			Height: height,
			Text:   string(raw),
		}, nil
	}

	term := vt10x.New(vt10x.WithSize(width, height))
	if _, err := term.Write(raw); err != nil {
		return Frame{}, err
	}

	var sb strings.Builder
	sb.Grow((width + 1) * height)
	line := make([]rune, 0, width)
	for y := 0; y < height; y++ {
		line = line[:0]
		for x := 0; x < width; x++ {
			g := term.Cell(x, y)
			c := g.Char
			if c == 0 {
				c = ' '
			}
			line = append(line, c)
		}
		trimmed := strings.TrimRight(string(line), " ")
		sb.WriteString(trimmed)
		if y < height-1 {
			sb.WriteByte('\n')
		}
	}

	return Frame{
		Width:  width,
		Height: height,
		Text:   trimTrailingBlankLines(sb.String()),
	}, nil
}

// trimTrailingBlankLines drops empty lines at the very end. Leading and
// interior blanks are preserved so layout stays recognisable.
func trimTrailingBlankLines(s string) string {
	b := []byte(s)
	for len(b) > 0 && (b[len(b)-1] == '\n' || b[len(b)-1] == ' ') {
		// Trim a single trailing newline group at a time.
		idx := bytes.LastIndexByte(b[:len(b)-1], '\n')
		if idx < 0 {
			break
		}
		// Check if the segment after the last newline is all blanks.
		tail := b[idx+1:]
		allBlank := true
		for _, c := range tail {
			if c != ' ' && c != '\n' {
				allBlank = false
				break
			}
		}
		if !allBlank {
			break
		}
		b = b[:idx]
	}
	return string(b)
}
