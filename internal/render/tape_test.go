package render

import (
	"strings"
	"testing"
)

func TestBuildTape_BasicPNG(t *testing.T) {
	opts := RenderOpts{Command: "./hello-tui"}
	opts.applyDefaults()
	tape, err := BuildTape(opts, "/tmp/x.png")
	if err != nil {
		t.Fatalf("BuildTape: %v", err)
	}
	wants := []string{
		`Set FontFamily "JetBrains Mono"`,
		`Set Theme "Dracula"`,
		`Type "exec ./hello-tui"`,
		`Enter`,
		`Screenshot "/tmp/x.png"`,
	}
	for _, w := range wants {
		if !strings.Contains(tape, w) {
			t.Errorf("missing %q in tape:\n%s", w, tape)
		}
	}
}

func TestBuildTape_Keys(t *testing.T) {
	opts := RenderOpts{Command: "./tui", Keys: []string{"tab", "ctrl+c", "hello", "shift+tab"}}
	opts.applyDefaults()
	tape, err := BuildTape(opts, "/tmp/x.png")
	if err != nil {
		t.Fatalf("BuildTape: %v", err)
	}
	for _, want := range []string{"\nTab\n", "\nCtrl+c\n", `Type "hello"`, "Shift+Tab"} {
		if !strings.Contains(tape, want) {
			t.Errorf("missing %q in tape:\n%s", want, tape)
		}
	}
}

func TestBuildTape_GIF(t *testing.T) {
	opts := RenderOpts{Command: "./tui", Format: "gif"}
	opts.applyDefaults()
	tape, err := BuildTape(opts, "/tmp/x.gif")
	if err != nil {
		t.Fatalf("BuildTape: %v", err)
	}
	if !strings.Contains(tape, `Output "/tmp/x.gif"`) {
		t.Errorf("expected Output line to point at gif, got:\n%s", tape)
	}
	if strings.Contains(tape, "Screenshot") {
		t.Errorf("gif tape should not have Screenshot command:\n%s", tape)
	}
}

func TestCacheKey_Deterministic(t *testing.T) {
	a, err := CacheKey(RenderOpts{Command: "x", Width: 80, Height: 24})
	if err != nil {
		t.Fatal(err)
	}
	b, err := CacheKey(RenderOpts{Command: "x"})
	if err != nil {
		t.Fatal(err)
	}
	// defaults make these equivalent
	if a != b {
		t.Errorf("expected same cache key after defaults, got %s vs %s", a, b)
	}
	c, err := CacheKey(RenderOpts{Command: "y"})
	if err != nil {
		t.Fatal(err)
	}
	if a == c {
		t.Errorf("expected different cache keys for different commands")
	}
}

func TestMapKeyToTape_Specials(t *testing.T) {
	cases := map[string]string{
		"enter":    "Enter",
		"f5":       "F5",
		"pgdown":   "PageDown",
		"ctrl+c":   "Ctrl+c",
		"alt+x":    "Alt+x",
		"shift+tab": "Shift+Tab",
	}
	for in, want := range cases {
		got, err := mapKeyToTape(in)
		if err != nil {
			t.Errorf("%q: %v", in, err)
			continue
		}
		if len(got) != 1 || got[0] != want {
			t.Errorf("%q: got %v, want [%s]", in, got, want)
		}
	}
}
