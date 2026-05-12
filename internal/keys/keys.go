package keys

import (
	"errors"
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

var specials = map[string][]byte{
	"enter":     {'\r'},
	"return":    {'\r'},
	"escape":    {0x1b},
	"esc":       {0x1b},
	"space":     {' '},
	"tab":       {'\t'},
	"backspace": {0x7f},
	"up":        {0x1b, '[', 'A'},
	"down":      {0x1b, '[', 'B'},
	"right":     {0x1b, '[', 'C'},
	"left":      {0x1b, '[', 'D'},
	"home":      {0x1b, '[', 'H'},
	"end":       {0x1b, '[', 'F'},
	"pgup":      {0x1b, '[', '5', '~'},
	"pgdown":    {0x1b, '[', '6', '~'},
	"delete":    {0x1b, '[', '3', '~'},
	"del":       {0x1b, '[', '3', '~'},
	"insert":    {0x1b, '[', '2', '~'},
	"ins":       {0x1b, '[', '2', '~'},
	"f1":        {0x1b, 'O', 'P'},
	"f2":        {0x1b, 'O', 'Q'},
	"f3":        {0x1b, 'O', 'R'},
	"f4":        {0x1b, 'O', 'S'},
	"f5":        {0x1b, '[', '1', '5', '~'},
	"f6":        {0x1b, '[', '1', '7', '~'},
	"f7":        {0x1b, '[', '1', '8', '~'},
	"f8":        {0x1b, '[', '1', '9', '~'},
	"f9":        {0x1b, '[', '2', '0', '~'},
	"f10":       {0x1b, '[', '2', '1', '~'},
	"f11":       {0x1b, '[', '2', '3', '~'},
	"f12":       {0x1b, '[', '2', '4', '~'},
}

func isModifier(s string) bool {
	switch s {
	case "ctrl", "alt", "shift", "meta":
		return true
	}
	return false
}

// Parse turns a single key spec into the byte sequence sent to a pty.
//
// Forms accepted:
//   - a special name: "enter", "tab", "f5", "pgup" (case-insensitive)
//   - a literal rune or string with no '+': "a", "/", "hello world"
//   - a modified key: "ctrl+c", "alt+x", "ctrl+alt+l", "shift+tab"
//
// A string containing '+' is treated as modified only if it starts with a
// recognized modifier (ctrl/alt/shift/meta); otherwise it's a literal.
func Parse(s string) ([]byte, error) {
	if s == "" {
		return nil, errors.New("keys: empty key string")
	}
	if !strings.Contains(s, "+") {
		if b, ok := specials[strings.ToLower(s)]; ok {
			return append([]byte(nil), b...), nil
		}
		return []byte(s), nil
	}

	parts := strings.Split(s, "+")
	if !isModifier(strings.ToLower(parts[0])) {
		return []byte(s), nil
	}

	mods := map[string]bool{}
	i := 0
	for i < len(parts)-1 {
		p := strings.ToLower(parts[i])
		if !isModifier(p) {
			return nil, fmt.Errorf("keys: unexpected token %q in %q (expected modifier or final key)", parts[i], s)
		}
		if mods[p] {
			return nil, fmt.Errorf("keys: duplicate modifier %q in %q", p, s)
		}
		mods[p] = true
		i++
	}
	keyName := parts[i]
	if keyName == "" {
		return nil, fmt.Errorf("keys: missing key after modifier in %q", s)
	}

	var base []byte
	if b, ok := specials[strings.ToLower(keyName)]; ok {
		base = append([]byte(nil), b...)
	} else if utf8.RuneCountInString(keyName) == 1 {
		base = []byte(keyName)
	} else {
		return nil, fmt.Errorf("keys: unknown key %q in %q", keyName, s)
	}

	if mods["shift"] {
		switch {
		case strings.EqualFold(keyName, "tab"):
			base = []byte{0x1b, '[', 'Z'}
		case utf8.RuneCountInString(keyName) == 1:
			r, _ := utf8.DecodeRuneInString(keyName)
			base = []byte(string(unicode.ToUpper(r)))
		default:
			return nil, fmt.Errorf("keys: shift modifier not supported with %q", keyName)
		}
		delete(mods, "shift")
	}

	if mods["ctrl"] {
		if len(base) != 1 {
			return nil, fmt.Errorf("keys: ctrl modifier requires a single character, got %q", keyName)
		}
		c := base[0]
		switch {
		case c >= 'a' && c <= 'z':
			base = []byte{c - 'a' + 1}
		case c >= 'A' && c <= 'Z':
			base = []byte{c - 'A' + 1}
		case c == '@', c == ' ':
			base = []byte{0x00}
		case c == '[':
			base = []byte{0x1b}
		case c == '\\':
			base = []byte{0x1c}
		case c == ']':
			base = []byte{0x1d}
		case c == '^':
			base = []byte{0x1e}
		case c == '_', c == '?':
			base = []byte{0x1f}
		default:
			return nil, fmt.Errorf("keys: ctrl modifier not supported with %q", keyName)
		}
		delete(mods, "ctrl")
	}

	if mods["alt"] || mods["meta"] {
		base = append([]byte{0x1b}, base...)
	}
	return base, nil
}

// ParseSequence parses each entry of keys and concatenates the resulting bytes.
// Error messages include the offending index.
func ParseSequence(seq []string) ([]byte, error) {
	var out []byte
	for i, s := range seq {
		b, err := Parse(s)
		if err != nil {
			return nil, fmt.Errorf("keys[%d]: %w", i, err)
		}
		out = append(out, b...)
	}
	return out, nil
}
