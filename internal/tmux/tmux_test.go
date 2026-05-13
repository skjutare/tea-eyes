package tmux_test

import (
	"context"
	"os/exec"
	"strings"
	"testing"
	"time"

	"gitlab.com/skjutare/tea-eyes/internal/tmux"
)

func requireTmux(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("tmux"); err != nil {
		t.Skip("tmux not installed; skipping tmux driver tests")
	}
}

func ctx(t *testing.T) (context.Context, context.CancelFunc) {
	t.Helper()
	return context.WithTimeout(context.Background(), 15*time.Second)
}

func TestNewSession_CreatesAndKills(t *testing.T) {
	requireTmux(t)
	c, cancel := ctx(t)
	defer cancel()

	d := tmux.New()
	sess, err := d.NewSession(c, tmux.SessionOpts{Width: 80, Height: 24})
	if err != nil {
		t.Fatalf("NewSession: %v", err)
	}
	if !strings.HasPrefix(sess.Name, "teaeyes-") {
		t.Errorf("expected generated session name to start with teaeyes-, got %q", sess.Name)
	}
	ok, err := d.HasSession(c, sess.Name)
	if err != nil || !ok {
		t.Fatalf("HasSession should be true after create: ok=%v err=%v", ok, err)
	}
	if err := sess.Close(c); err != nil {
		t.Fatalf("Close: %v", err)
	}
	ok, _ = d.HasSession(c, sess.Name)
	if ok {
		t.Error("session should be gone after Close on owned session")
	}
}

func TestHasSession_Missing(t *testing.T) {
	requireTmux(t)
	c, cancel := ctx(t)
	defer cancel()

	d := tmux.New()
	ok, err := d.HasSession(c, "teaeyes-definitely-not-here-xyz")
	if err != nil {
		t.Fatalf("HasSession on missing should not error, got %v", err)
	}
	if ok {
		t.Error("HasSession should return false for missing session")
	}
}

func TestCapturePane_RunCommand(t *testing.T) {
	requireTmux(t)
	c, cancel := ctx(t)
	defer cancel()

	d := tmux.New()
	sess, err := d.NewSession(c, tmux.SessionOpts{Width: 80, Height: 24})
	if err != nil {
		t.Fatalf("NewSession: %v", err)
	}
	t.Cleanup(func() { _ = sess.Close(context.Background()) })

	if err := sess.RunCommand(c, "echo hello-tmux-driver"); err != nil {
		t.Fatalf("RunCommand: %v", err)
	}
	// Give the shell a beat to render.
	time.Sleep(400 * time.Millisecond)
	out, err := sess.CapturePane(c, false)
	if err != nil {
		t.Fatalf("CapturePane: %v", err)
	}
	if !strings.Contains(out, "hello-tmux-driver") {
		t.Errorf("expected captured pane to contain echoed string, got:\n%s", out)
	}
}

func TestAttachExisting_PreservesAcrossHandles(t *testing.T) {
	requireTmux(t)
	c, cancel := ctx(t)
	defer cancel()

	d := tmux.New()
	owned, err := d.NewSession(c, tmux.SessionOpts{Name: "teaeyes-attach-test", Width: 80, Height: 24})
	if err != nil {
		t.Fatalf("NewSession: %v", err)
	}
	t.Cleanup(func() { _ = owned.Kill(context.Background()) })

	attached, err := d.AttachExisting(c, "teaeyes-attach-test")
	if err != nil {
		t.Fatalf("AttachExisting: %v", err)
	}
	if attached.Owned {
		t.Error("AttachExisting must not claim ownership")
	}
	if err := attached.Close(c); err != nil {
		t.Fatalf("Close on attached: %v", err)
	}
	ok, _ := d.HasSession(c, "teaeyes-attach-test")
	if !ok {
		t.Error("non-owned Close must not kill the session")
	}
}

func TestNewSession_RejectsBadName(t *testing.T) {
	requireTmux(t)
	c, cancel := ctx(t)
	defer cancel()

	d := tmux.New()
	if _, err := d.NewSession(c, tmux.SessionOpts{Name: "bad name with spaces"}); err == nil {
		t.Fatal("expected error for invalid session name")
	}
}
