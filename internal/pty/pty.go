// Package pty wraps github.com/creack/pty to spawn a TUI binary attached to a
// pseudo-terminal of fixed size, deliver keystrokes, and collect the rendered
// output bytes.
package pty

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/creack/pty"
)

// Driver is a thin coordinator around creack/pty. The zero value is not
// usable; construct via New.
type Driver struct{}

// New returns a ready-to-use Driver. It does no setup so the call is cheap.
func New() *Driver { return &Driver{} }

// SpawnOpts describes how to launch the TUI under test.
type SpawnOpts struct {
	Command  string
	Args     []string
	Width    int      // columns; defaults to 80
	Height   int      // rows; defaults to 24
	Env      []string // child env; defaults to os.Environ()
	Cwd      string   // working directory; defaults to caller's
	SettleMs int      // wait after start and between key sends; defaults to 300
}

const (
	defaultWidth    = 80
	defaultHeight   = 24
	defaultSettleMs = 300
	cleanupGrace    = 500 * time.Millisecond
	readChunkSize   = 4096
)

// Capture spawns the command attached to a pty of the requested size, writes
// each key sequence with a settle delay in between, then terminates the
// process and returns the accumulated raw pty output.
//
// The reader runs in a goroutine to keep the pty buffer drained, so child
// processes that emit a lot of output won't block.
//
//nolint:gocognit,funlen // intrinsic complexity: pty reader + writer + ctx + child-exit must coexist
func (d *Driver) Capture(ctx context.Context, opts SpawnOpts, keys [][]byte) ([]byte, error) {
	if opts.Width <= 0 {
		opts.Width = defaultWidth
	}
	if opts.Height <= 0 {
		opts.Height = defaultHeight
	}
	if opts.SettleMs <= 0 {
		opts.SettleMs = defaultSettleMs
	}
	if opts.Command == "" {
		return nil, errors.New("pty: Command is required")
	}

	if _, err := exec.LookPath(opts.Command); err != nil {
		return nil, fmt.Errorf("pty: command not found: %q (PATH=%s)", opts.Command, os.Getenv("PATH"))
	}

	//nolint:gosec // G204: launching a subprocess from caller-supplied args is this package's purpose
	cmd := exec.CommandContext(ctx, opts.Command, opts.Args...)
	env := opts.Env
	if env == nil {
		env = os.Environ()
	}
	if !hasEnvKey(env, "TERM") {
		env = append(env, "TERM=xterm-256color")
	}
	cmd.Env = env
	if opts.Cwd != "" {
		cmd.Dir = opts.Cwd
	}

	ptyFile, err := pty.StartWithSize(cmd, &pty.Winsize{
		Cols: uint16(opts.Width),  //nolint:gosec // G115: width is caller-bounded to terminal columns
		Rows: uint16(opts.Height), //nolint:gosec // G115: height is caller-bounded to terminal rows
	})
	if err != nil {
		return nil, fmt.Errorf("pty: failed to start %q under a pseudo-terminal: %w", opts.Command, err)
	}

	var (
		mu       sync.Mutex
		buf      bytes.Buffer
		readDone = make(chan struct{})
		scanPos  int
	)
	go func() {
		defer close(readDone)
		chunk := make([]byte, readChunkSize)
		for {
			n, readErr := ptyFile.Read(chunk)
			if n > 0 {
				mu.Lock()
				buf.Write(chunk[:n])
				// Scan recent bytes for terminal capability queries that the
				// TUI may be blocking on (termenv emits these on startup) and
				// reply on the master side so the child can proceed.
				b := buf.Bytes()
				safeEnd := len(b)
				if safeEnd > scanPos {
					replies := scanForQueryReplies(b[scanPos:safeEnd], opts.Width, opts.Height)
					scanPos = max(safeEnd-maxQueryLen, 0)
					mu.Unlock()
					for _, r := range replies {
						_, _ = ptyFile.Write(r)
					}
				} else {
					mu.Unlock()
				}
			}
			if readErr != nil {
				if !errors.Is(readErr, io.EOF) && !errors.Is(readErr, fs.ErrClosed) {
					// EIO is normal when the slave side closes on darwin/linux;
					// nothing actionable.
					_ = readErr
				}
				return
			}
		}
	}()

	waitDone := make(chan error, 1)
	go func() { waitDone <- cmd.Wait() }()

	settle := time.Duration(opts.SettleMs) * time.Millisecond
	time.Sleep(settle)

	select {
	case werr := <-waitDone:
		_ = ptyFile.Close()
		<-readDone
		snap := snapshot(&mu, &buf)
		exitCode := -1
		if cmd.ProcessState != nil {
			exitCode = cmd.ProcessState.ExitCode()
		}
		return snap, fmt.Errorf(
			"pty: process %q exited before initial settle (exit=%d, err=%w)",
			opts.Command,
			exitCode,
			werr,
		)
	case <-ctx.Done():
		_ = cmd.Process.Signal(syscall.SIGTERM)
		<-waitDone
		_ = ptyFile.Close()
		<-readDone
		return snapshot(&mu, &buf), ctx.Err()
	default:
	}

	for i, k := range keys {
		if len(k) == 0 {
			continue
		}
		if _, writeErr := ptyFile.Write(k); writeErr != nil {
			_ = cmd.Process.Signal(syscall.SIGTERM)
			<-waitDone
			_ = ptyFile.Close()
			<-readDone
			return snapshot(&mu, &buf), fmt.Errorf("pty: failed to write key #%d (%d bytes): %w", i, len(k), writeErr)
		}
		time.Sleep(settle)
		select {
		case <-waitDone:
			// Child exited mid-sequence (e.g., quit key). Drain and return.
			_ = ptyFile.Close()
			<-readDone
			return snapshot(&mu, &buf), nil
		default:
		}
	}

	_ = cmd.Process.Signal(syscall.SIGTERM)
	select {
	case <-waitDone:
	case <-time.After(cleanupGrace):
		_ = cmd.Process.Kill()
		<-waitDone
	}
	_ = ptyFile.Close()
	<-readDone

	return snapshot(&mu, &buf), nil
}

func snapshot(mu *sync.Mutex, buf *bytes.Buffer) []byte {
	mu.Lock()
	defer mu.Unlock()
	out := make([]byte, buf.Len())
	copy(out, buf.Bytes())
	return out
}

func hasEnvKey(env []string, key string) bool {
	prefix := key + "="
	for _, e := range env {
		if strings.HasPrefix(e, prefix) {
			return true
		}
	}
	return false
}
