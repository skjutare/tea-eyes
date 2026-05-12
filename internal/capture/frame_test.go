package capture

import (
	"strings"
	"testing"
)

func TestRenderFrame_PlainText(t *testing.T) {
	raw := []byte("hello\r\nworld\r\n")
	f, err := RenderFrame(raw, 20, 5, true)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(f.Text, "hello") || !strings.Contains(f.Text, "world") {
		t.Errorf("frame missing expected lines:\n%s", f.Text)
	}
	if f.Width != 20 || f.Height != 5 {
		t.Errorf("unexpected size: %d×%d", f.Width, f.Height)
	}
}

func TestRenderFrame_StripsSGR(t *testing.T) {
	// red "hi" via SGR — stripped result should just be "hi"
	raw := []byte("\x1b[31mhi\x1b[0m")
	f, err := RenderFrame(raw, 10, 2, true)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(f.Text, "hi") {
		t.Errorf("expected 'hi' at start, got %q", f.Text)
	}
	if strings.Contains(f.Text, "\x1b") {
		t.Errorf("stripped output should not contain ESC: %q", f.Text)
	}
}

func TestRenderFrame_PreservesSGRWhenNotStripping(t *testing.T) {
	raw := []byte("\x1b[31mhi\x1b[0m")
	f, err := RenderFrame(raw, 10, 2, false)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(f.Text, "\x1b[31m") {
		t.Errorf("expected SGR retained, got %q", f.Text)
	}
}

func TestRenderFrame_CursorPositioning(t *testing.T) {
	// Move to row 2, col 1 and write "X"
	raw := []byte("\x1b[2;1HX")
	f, err := RenderFrame(raw, 5, 3, true)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(f.Text, "\n")
	if len(lines) < 2 || !strings.HasPrefix(lines[1], "X") {
		t.Errorf("expected X on row 2, got:\n%s", f.Text)
	}
}
