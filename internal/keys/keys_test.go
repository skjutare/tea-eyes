package keys

import (
	"bytes"
	"testing"
)

func TestParse_Literals(t *testing.T) {
	cases := []struct {
		in   string
		want []byte
	}{
		{"a", []byte("a")},
		{"Z", []byte("Z")},
		{"7", []byte("7")},
		{"/", []byte("/")},
		{"hello world", []byte("hello world")},
		{"å", []byte("å")},
		{"漢", []byte("漢")},
	}
	for _, c := range cases {
		got, err := Parse(c.in)
		if err != nil {
			t.Errorf("Parse(%q) error: %v", c.in, err)
			continue
		}
		if !bytes.Equal(got, c.want) {
			t.Errorf("Parse(%q) = %v, want %v", c.in, got, c.want)
		}
	}
}

func TestParse_Specials(t *testing.T) {
	cases := map[string][]byte{
		"enter":     {'\r'},
		"ENTER":     {'\r'},
		"escape":    {0x1b},
		"esc":       {0x1b},
		"space":     {' '},
		"tab":       {'\t'},
		"backspace": {0x7f},
		"up":        {0x1b, '[', 'A'},
		"down":      {0x1b, '[', 'B'},
		"left":      {0x1b, '[', 'D'},
		"right":     {0x1b, '[', 'C'},
		"home":      {0x1b, '[', 'H'},
		"end":       {0x1b, '[', 'F'},
		"pgup":      {0x1b, '[', '5', '~'},
		"pgdown":    {0x1b, '[', '6', '~'},
		"delete":    {0x1b, '[', '3', '~'},
		"insert":    {0x1b, '[', '2', '~'},
		"f1":        {0x1b, 'O', 'P'},
		"f4":        {0x1b, 'O', 'S'},
		"f5":        {0x1b, '[', '1', '5', '~'},
		"f12":       {0x1b, '[', '2', '4', '~'},
	}
	for in, want := range cases {
		got, err := Parse(in)
		if err != nil {
			t.Errorf("Parse(%q) error: %v", in, err)
			continue
		}
		if !bytes.Equal(got, want) {
			t.Errorf("Parse(%q) = %v, want %v", in, got, want)
		}
	}
}

func TestParse_Ctrl(t *testing.T) {
	cases := map[string][]byte{
		"ctrl+c":     {0x03},
		"ctrl+a":     {0x01},
		"ctrl+z":     {0x1a},
		"CTRL+C":     {0x03},
		"ctrl+@":     {0x00},
		"ctrl+space": {0x00},
		"ctrl+[":     {0x1b},
	}
	for in, want := range cases {
		got, err := Parse(in)
		if err != nil {
			t.Errorf("Parse(%q) error: %v", in, err)
			continue
		}
		if !bytes.Equal(got, want) {
			t.Errorf("Parse(%q) = %v, want %v", in, got, want)
		}
	}
}

func TestParse_Alt(t *testing.T) {
	cases := map[string][]byte{
		"alt+x":     {0x1b, 'x'},
		"alt+enter": {0x1b, '\r'},
		"meta+a":    {0x1b, 'a'},
	}
	for in, want := range cases {
		got, err := Parse(in)
		if err != nil {
			t.Errorf("Parse(%q) error: %v", in, err)
			continue
		}
		if !bytes.Equal(got, want) {
			t.Errorf("Parse(%q) = %v, want %v", in, got, want)
		}
	}
}

func TestParse_Shift(t *testing.T) {
	cases := map[string][]byte{
		"shift+tab": {0x1b, '[', 'Z'},
		"shift+a":   {'A'},
	}
	for in, want := range cases {
		got, err := Parse(in)
		if err != nil {
			t.Errorf("Parse(%q) error: %v", in, err)
			continue
		}
		if !bytes.Equal(got, want) {
			t.Errorf("Parse(%q) = %v, want %v", in, got, want)
		}
	}
}

func TestParse_Combined(t *testing.T) {
	cases := map[string][]byte{
		"ctrl+alt+l": {0x1b, 0x0c},
		"alt+ctrl+l": {0x1b, 0x0c},
	}
	for in, want := range cases {
		got, err := Parse(in)
		if err != nil {
			t.Errorf("Parse(%q) error: %v", in, err)
			continue
		}
		if !bytes.Equal(got, want) {
			t.Errorf("Parse(%q) = %v, want %v", in, got, want)
		}
	}
}

func TestParse_Errors(t *testing.T) {
	bad := []string{
		"",
		"ctrl+",
		"ctrl+ctrl+a",
		"ctrl+meow",
		"shift+enter",
	}
	for _, in := range bad {
		if _, err := Parse(in); err == nil {
			t.Errorf("Parse(%q) expected error, got nil", in)
		}
	}
}

func TestParse_LiteralWithPlus(t *testing.T) {
	got, err := Parse("a+b")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !bytes.Equal(got, []byte("a+b")) {
		t.Errorf("Parse(\"a+b\") = %v, want literal a+b", got)
	}
}

func TestParseSequence(t *testing.T) {
	got, err := ParseSequence([]string{"hello", "space", "world", "enter"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []byte("hello world\r")
	if !bytes.Equal(got, want) {
		t.Errorf("got %q, want %q", got, want)
	}

	if _, err := ParseSequence([]string{"a", "ctrl+meow"}); err == nil {
		t.Error("expected error from bad sequence")
	}
}
