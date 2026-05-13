//go:build teaeyes

package main

import (
	"io"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/exp/teatest"
	"github.com/muesli/termenv"
)

// TestHelloTUI_HandWrittenGolden mirrors the kind of test a user would write
// directly with teatest, independent of the tea-eyes harness. Run with:
//
//	go test -tags teaeyes ./examples/hello-tui -run TestHelloTUI_HandWrittenGolden
//
// Behavior tests should pin the Ascii color profile for determinism.
func TestHelloTUI_HandWrittenGolden(t *testing.T) {
	lipgloss.SetColorProfile(termenv.Ascii)

	tm := teatest.NewTestModel(t, model{}, teatest.WithInitialTermSize(40, 8))
	tm.Send(tea.KeyMsg{Type: tea.KeySpace})
	time.Sleep(20 * time.Millisecond)
	_ = tm.Quit()

	out, _ := io.ReadAll(tm.FinalOutput(t, teatest.WithFinalTimeout(2*time.Second)))
	if !strings.Contains(string(out), "Counter: 1") {
		t.Fatalf("expected counter in output, got:\n%s", string(out))
	}
}
