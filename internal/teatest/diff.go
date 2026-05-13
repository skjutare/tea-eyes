package teatest

import (
	"fmt"
	"strings"
)

// unifiedDiff produces a small unified-style diff between want and got.
// It is intentionally minimal — line-by-line, no LCS — because golden files
// are typically tens of lines.
func unifiedDiff(want, got string) string {
	if want == got {
		return ""
	}
	wantLines := strings.Split(want, "\n")
	gotLines := strings.Split(got, "\n")

	var b strings.Builder
	b.WriteString("--- golden\n+++ actual\n")
	maxN := max(len(gotLines), len(wantLines))
	for i := range maxN {
		var w, g string
		var hasW, hasG bool
		if i < len(wantLines) {
			w = wantLines[i]
			hasW = true
		}
		if i < len(gotLines) {
			g = gotLines[i]
			hasG = true
		}
		switch {
		case hasW && hasG && w == g:
			fmt.Fprintf(&b, " %s\n", w)
		case hasW && hasG:
			fmt.Fprintf(&b, "-%s\n+%s\n", w, g)
		case hasW:
			fmt.Fprintf(&b, "-%s\n", w)
		case hasG:
			fmt.Fprintf(&b, "+%s\n", g)
		}
	}
	return b.String()
}
