// Package render wraps Charm's VHS to produce PNG/GIF screenshots of a TUI
// session. The renderer generates a .tape file from a RenderOpts spec, runs
// the vhs binary, and returns the resulting image bytes. Results are cached
// under $XDG_CACHE_HOME/tea-eyes/renders/ keyed by a SHA-256 of the inputs.
package render

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// RenderOpts describes a single VHS render request.
type RenderOpts struct {
	Command    string   `json:"command"`
	Args       []string `json:"args,omitempty"`
	Keys       []string `json:"keys,omitempty"`
	Width      int      `json:"width"`
	Height     int      `json:"height"`
	FontFamily string   `json:"font_family"`
	FontSize   int      `json:"font_size"`
	Theme      string   `json:"theme"`
	Format     string   `json:"format"` // "png" or "gif"
	Padding    int      `json:"padding"`
	SettleMs   int      `json:"settle_ms"`
	Cwd        string   `json:"cwd,omitempty"`
}

// Render is the result of a successful VHS render.
type Render struct {
	Bytes      []byte
	Mime       string
	Format     string
	CachePath  string
	CacheHit   bool
	TapeSource string
}

// Renderer runs vhs subprocesses and manages the on-disk render cache.
type Renderer struct {
	cacheDir string
	vhsPath  string
}

// NewRenderer returns a Renderer rooted at cacheDir. If cacheDir is empty,
// the default ($XDG_CACHE_HOME/tea-eyes/renders or ~/.cache/tea-eyes/renders)
// is used.
func NewRenderer(cacheDir string) *Renderer {
	if cacheDir == "" {
		cacheDir = DefaultCacheDir()
	}
	return &Renderer{cacheDir: cacheDir}
}

// CacheDir returns the cache directory the renderer writes to.
func (r *Renderer) CacheDir() string { return r.cacheDir }

// DefaultCacheDir returns the platform cache directory for tea-eyes renders.
func DefaultCacheDir() string {
	if xdg := os.Getenv("XDG_CACHE_HOME"); xdg != "" {
		return filepath.Join(xdg, "tea-eyes", "renders")
	}
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return filepath.Join(os.TempDir(), "tea-eyes", "renders")
	}
	return filepath.Join(home, ".cache", "tea-eyes", "renders")
}

// applyDefaults fills zero values with the documented defaults.
func (o *RenderOpts) applyDefaults() {
	if o.Width == 0 {
		o.Width = 80
	}
	if o.Height == 0 {
		o.Height = 24
	}
	if o.FontFamily == "" {
		o.FontFamily = "JetBrains Mono"
	}
	if o.FontSize == 0 {
		o.FontSize = 14
	}
	if o.Theme == "" {
		o.Theme = "Dracula"
	}
	if o.Format == "" {
		o.Format = "png"
	}
	if o.Padding == 0 {
		o.Padding = 20
	}
	if o.SettleMs == 0 {
		o.SettleMs = 300
	}
}

// validate reports any input errors.
func (o RenderOpts) validate() error {
	if strings.TrimSpace(o.Command) == "" {
		return errors.New("render: command is required")
	}
	switch o.Format {
	case "png", "gif":
	default:
		return fmt.Errorf("render: unsupported format %q (want \"png\" or \"gif\")", o.Format)
	}
	return nil
}

// CacheKey returns the SHA-256 hex digest used as the on-disk cache key for
// the given options.
func CacheKey(o RenderOpts) (string, error) {
	o.applyDefaults()
	buf, err := json.Marshal(o)
	if err != nil {
		return "", fmt.Errorf("render: hash opts: %w", err)
	}
	sum := sha256.Sum256(buf)
	return hex.EncodeToString(sum[:]), nil
}

// Render produces an image for opts, consulting the cache unless noCache is
// set.
func (r *Renderer) Render(ctx context.Context, opts RenderOpts, noCache bool) (Render, error) {
	opts.applyDefaults()
	if err := opts.validate(); err != nil {
		return Render{}, err
	}

	key, err := CacheKey(opts)
	if err != nil {
		return Render{}, err
	}
	cachePath := filepath.Join(r.cacheDir, key+"."+opts.Format)
	tapePath := filepath.Join(r.cacheDir, key+".tape")
	mime := mimeFor(opts.Format)

	if !noCache {
		if info, err := os.Stat(cachePath); err == nil && info.Size() > 0 {
			data, err := os.ReadFile(cachePath)
			if err == nil {
				tape, _ := os.ReadFile(tapePath)
				return Render{
					Bytes: data, Mime: mime, Format: opts.Format,
					CachePath: cachePath, CacheHit: true,
					TapeSource: string(tape),
				}, nil
			}
		}
	}

	if err := os.MkdirAll(r.cacheDir, 0o755); err != nil {
		return Render{}, fmt.Errorf("render: create cache dir %q: %w", r.cacheDir, err)
	}

	vhsBin, err := r.resolveVHS()
	if err != nil {
		return Render{}, err
	}

	tape, err := BuildTape(opts, cachePath)
	if err != nil {
		return Render{}, fmt.Errorf("render: build tape: %w", err)
	}
	if err := os.WriteFile(tapePath, []byte(tape), 0o644); err != nil {
		return Render{}, fmt.Errorf("render: write tape: %w", err)
	}

	cmd := exec.CommandContext(ctx, vhsBin, tapePath)
	if opts.Cwd != "" {
		cmd.Dir = opts.Cwd
	}
	out, err := cmd.CombinedOutput()
	if err != nil {
		return Render{}, fmt.Errorf("render: vhs %s failed: %w\ntape:\n%s\noutput:\n%s",
			tapePath, err, tape, string(out))
	}

	info, err := os.Stat(cachePath)
	if err != nil {
		return Render{}, fmt.Errorf("render: vhs did not produce %s (output: %s)", cachePath, string(out))
	}
	if info.Size() == 0 {
		return Render{}, fmt.Errorf("render: vhs produced empty file %s", cachePath)
	}

	data, err := os.ReadFile(cachePath)
	if err != nil {
		return Render{}, fmt.Errorf("render: read output: %w", err)
	}
	return Render{
		Bytes: data, Mime: mime, Format: opts.Format,
		CachePath: cachePath, CacheHit: false,
		TapeSource: tape,
	}, nil
}

// resolveVHS finds the vhs binary on PATH and caches it on the Renderer.
func (r *Renderer) resolveVHS() (string, error) {
	if r.vhsPath != "" {
		return r.vhsPath, nil
	}
	p, err := exec.LookPath("vhs")
	if err != nil {
		return "", fmt.Errorf(
			"render: vhs not found on PATH — install with "+
				"`brew install vhs` or `go install github.com/charmbracelet/vhs@latest` "+
				"(also requires ttyd and ffmpeg). Underlying error: %w", err)
	}
	r.vhsPath = p
	return p, nil
}

// CleanCache removes every file in the renderer's cache directory but leaves
// the directory itself in place. Returns the number of files removed.
func (r *Renderer) CleanCache() (int, error) {
	entries, err := os.ReadDir(r.cacheDir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return 0, nil
		}
		return 0, fmt.Errorf("render: read cache dir: %w", err)
	}
	removed := 0
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if err := os.Remove(filepath.Join(r.cacheDir, e.Name())); err != nil {
			return removed, fmt.Errorf("render: remove %s: %w", e.Name(), err)
		}
		removed++
	}
	return removed, nil
}

func mimeFor(format string) string {
	switch format {
	case "gif":
		return "image/gif"
	default:
		return "image/png"
	}
}
