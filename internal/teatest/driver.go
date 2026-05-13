// Package teatest generates and runs in-process Bubble Tea harnesses using
// the white-box TeaEyesNewModel pattern. It writes a build-tagged _test.go
// file into the user's package, compiles it once via `go test -c -tags
// teaeyes`, caches the resulting binary under
// $XDG_CACHE_HOME/tea-eyes/teatest/, and drives it with runtime knobs
// supplied through environment variables.
package teatest

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
	"sort"
	"strconv"
	"strings"
)

// Driver runs in-process Bubble Tea harnesses by generating a build-tagged
// test file in the user's package, compiling it once with `go test -c`, and
// executing the cached binary against runtime inputs supplied via env vars.
type Driver struct {
	cacheDir string
}

// NewDriver returns a Driver rooted at cacheDir. Empty means the default
// ($XDG_CACHE_HOME/tea-eyes/teatest).
func NewDriver(cacheDir string) *Driver {
	if cacheDir == "" {
		cacheDir = DefaultCacheDir()
	}
	return &Driver{cacheDir: cacheDir}
}

// CacheDir returns the cache directory.
func (d *Driver) CacheDir() string { return d.cacheDir }

// DefaultCacheDir is the on-disk location of cached teatest harness binaries.
func DefaultCacheDir() string {
	if xdg := os.Getenv("XDG_CACHE_HOME"); xdg != "" {
		return filepath.Join(xdg, "tea-eyes", "teatest")
	}
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return filepath.Join(os.TempDir(), "tea-eyes", "teatest")
	}
	return filepath.Join(home, ".cache", "tea-eyes", "teatest")
}

// GoldenOpts configures a tui_test_golden invocation.
type GoldenOpts struct {
	PackagePath  string   // directory containing the user's Bubble Tea package
	ModelFunc    string   // exported constructor name; default "TeaEyesNewModel"
	Keys         []string // key sequence
	Width        int      // terminal columns, default 80
	Height       int      // terminal rows, default 24
	ColorProfile string   // "Ascii"|"ANSI"|"ANSI256"|"TrueColor"; default "Ascii"
	GoldenFile   string   // golden file path
	UpdateGolden bool     // when true, overwrite the golden with the current output
}

// GoldenResult is the outcome of RunGolden.
type GoldenResult struct {
	FinalOutput string
	Match       bool
	Diff        string // unified-ish diff when Match is false
	Created     bool   // true when GoldenFile did not exist and was just written
	GoldenPath  string
}

// InspectOpts configures a tui_inspect_model invocation.
type InspectOpts struct {
	PackagePath  string
	ModelFunc    string
	Keys         []string
	Width        int
	Height       int
	ColorProfile string
}

// InspectResult is the outcome of Inspect.
type InspectResult struct {
	ModelJSON   string
	ViewText    string
	FinalOutput string
}

func applyDefaults(modelFunc string, width, height int, profile string) (string, int, int, string) {
	if modelFunc == "" {
		modelFunc = "TeaEyesNewModel"
	}
	if width == 0 {
		width = 80
	}
	if height == 0 {
		height = 24
	}
	if profile == "" {
		profile = "Ascii"
	}
	return modelFunc, width, height, profile
}

// RunGolden generates the harness (if needed), invokes it, and compares the
// captured output against opts.GoldenFile.
func (d *Driver) RunGolden(ctx context.Context, opts GoldenOpts) (GoldenResult, error) {
	if opts.PackagePath == "" {
		return GoldenResult{}, errors.New("teatest: package_path is required")
	}
	if opts.GoldenFile == "" {
		return GoldenResult{}, errors.New("teatest: golden_file is required")
	}

	modelFunc, width, height, profile := applyDefaults(opts.ModelFunc, opts.Width, opts.Height, opts.ColorProfile)

	out, err := d.runHarness(ctx, opts.PackagePath, modelFunc, harnessInvocation{
		Op:           "golden",
		Keys:         opts.Keys,
		Width:        width,
		Height:       height,
		ColorProfile: profile,
	})
	if err != nil {
		return GoldenResult{}, err
	}

	final := out.FinalOutput

	res := GoldenResult{
		FinalOutput: final,
		GoldenPath:  opts.GoldenFile,
	}

	existing, statErr := os.ReadFile(opts.GoldenFile)
	missing := errors.Is(statErr, os.ErrNotExist)
	if !missing && statErr != nil {
		return GoldenResult{}, fmt.Errorf("teatest: read golden %q: %w", opts.GoldenFile, statErr)
	}

	if missing || opts.UpdateGolden {
		if mkErr := os.MkdirAll(filepath.Dir(opts.GoldenFile), 0o750); mkErr != nil {
			return GoldenResult{}, fmt.Errorf("teatest: mkdir for golden: %w", mkErr)
		}
		if wrErr := os.WriteFile(opts.GoldenFile, []byte(final), 0o600); wrErr != nil {
			return GoldenResult{}, fmt.Errorf("teatest: write golden %q: %w", opts.GoldenFile, wrErr)
		}
		res.Match = true
		res.Created = missing
		return res, nil
	}

	if string(existing) == final {
		res.Match = true
		return res, nil
	}
	res.Match = false
	res.Diff = unifiedDiff(string(existing), final)
	return res, nil
}

// Inspect generates the harness, invokes it in "inspect" mode, and returns
// the JSON-encoded exported fields of the final model plus its current view.
func (d *Driver) Inspect(ctx context.Context, opts InspectOpts) (InspectResult, error) {
	if opts.PackagePath == "" {
		return InspectResult{}, errors.New("teatest: package_path is required")
	}
	modelFunc, width, height, profile := applyDefaults(opts.ModelFunc, opts.Width, opts.Height, opts.ColorProfile)

	out, err := d.runHarness(ctx, opts.PackagePath, modelFunc, harnessInvocation{
		Op:           "inspect",
		Keys:         opts.Keys,
		Width:        width,
		Height:       height,
		ColorProfile: profile,
	})
	if err != nil {
		return InspectResult{}, err
	}
	return InspectResult{
		ModelJSON:   out.ModelJSON,
		ViewText:    out.ViewText,
		FinalOutput: out.FinalOutput,
	}, nil
}

type harnessInvocation struct {
	Op           string
	Keys         []string
	Width        int
	Height       int
	ColorProfile string
}

type harnessOutput struct {
	FinalOutput string `json:"final_output"`
	ModelJSON   string `json:"model_json"`
	ViewText    string `json:"view_text"`
}

func (d *Driver) runHarness(
	ctx context.Context,
	pkgPath, modelFunc string,
	inv harnessInvocation,
) (harnessOutput, error) {
	absPkg, err := filepath.Abs(pkgPath)
	if err != nil {
		return harnessOutput{}, fmt.Errorf("teatest: resolve package path %q: %w", pkgPath, err)
	}
	info, err := os.Stat(absPkg)
	if err != nil {
		return harnessOutput{}, fmt.Errorf("teatest: stat package path %q: %w", absPkg, err)
	}
	if !info.IsDir() {
		return harnessOutput{}, fmt.Errorf("teatest: package_path must be a directory, got %s", absPkg)
	}

	pkgName, err := detectPackageName(absPkg)
	if err != nil {
		return harnessOutput{}, err
	}

	moduleRoot, err := findModuleRoot(absPkg)
	if err != nil {
		return harnessOutput{}, err
	}

	if mkErr := os.MkdirAll(d.cacheDir, 0o750); mkErr != nil {
		return harnessOutput{}, fmt.Errorf("teatest: create cache dir: %w", mkErr)
	}

	binPath, err := d.ensureHarnessBinary(ctx, absPkg, moduleRoot, pkgName, modelFunc)
	if err != nil {
		return harnessOutput{}, err
	}

	keysJSON := "[]"
	if len(inv.Keys) > 0 {
		b, marshalErr := json.Marshal(inv.Keys)
		if marshalErr != nil {
			return harnessOutput{}, fmt.Errorf("teatest: marshal keys: %w", marshalErr)
		}
		keysJSON = string(b)
	}

	cmd := exec.CommandContext(ctx, binPath, "-test.run", "^TestTeaEyesHarness$", "-test.v")
	cmd.Env = append(os.Environ(),
		"TEA_EYES_OP="+inv.Op,
		"TEA_EYES_KEYS="+keysJSON,
		"TEA_EYES_WIDTH="+strconv.Itoa(inv.Width),
		"TEA_EYES_HEIGHT="+strconv.Itoa(inv.Height),
		"TEA_EYES_COLOR="+inv.ColorProfile,
	)
	raw, err := cmd.CombinedOutput()
	if err != nil {
		return harnessOutput{}, fmt.Errorf("teatest: harness binary failed: %w\noutput:\n%s", err, string(raw))
	}

	jsonLine, ok := extractResultLine(string(raw))
	if !ok {
		return harnessOutput{}, fmt.Errorf("teatest: result sentinel not found in harness output:\n%s", string(raw))
	}
	var out harnessOutput
	if jsonErr := json.Unmarshal([]byte(jsonLine), &out); jsonErr != nil {
		return harnessOutput{}, fmt.Errorf("teatest: parse harness JSON: %w\nline: %s", jsonErr, jsonLine)
	}
	return out, nil
}

func extractResultLine(combined string) (string, bool) {
	_, after, ok := strings.Cut(combined, resultSentinelStart)
	if !ok {
		return "", false
	}
	rest := after
	before, _, ok := strings.Cut(rest, resultSentinelEnd)
	if !ok {
		return "", false
	}
	return before, true
}

func (d *Driver) ensureHarnessBinary(
	ctx context.Context,
	absPkg, moduleRoot, pkgName, modelFunc string,
) (string, error) {
	key, err := harnessCacheKey(absPkg, moduleRoot, pkgName, modelFunc)
	if err != nil {
		return "", err
	}
	binPath := filepath.Join(d.cacheDir, "harness-"+key+binExt())
	if info, statErr := os.Stat(binPath); statErr == nil && info.Size() > 0 {
		return binPath, nil
	}

	harness, err := renderHarness(pkgName, modelFunc)
	if err != nil {
		return "", err
	}
	harnessPath := filepath.Join(absPkg, harnessFileName)
	if wrErr := os.WriteFile(harnessPath, harness, 0o600); wrErr != nil {
		return "", fmt.Errorf("teatest: write harness file %q: %w", harnessPath, wrErr)
	}
	defer os.Remove(harnessPath)

	cmd := exec.CommandContext(ctx, "go", "test", "-tags", "teaeyes", "-c", "-o", binPath, absPkg)
	cmd.Dir = moduleRoot
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf(
			"teatest: building harness for %s failed: %w\n"+
				"this usually means the package does not define %s() tea.Model under build tag `teaeyes`.\n"+
				"compiler output:\n%s",
			absPkg, err, modelFunc, string(out))
	}
	return binPath, nil
}

func harnessCacheKey(absPkg, moduleRoot, pkgName, modelFunc string) (string, error) {
	h := sha256.New()
	fmt.Fprintf(h, "pkg=%s\nmodel=%s\nname=%s\nmodroot=%s\n", absPkg, modelFunc, pkgName, moduleRoot)

	files, err := goFilesIn(absPkg)
	if err != nil {
		return "", err
	}
	sort.Strings(files)
	for _, f := range files {
		b, readErr := os.ReadFile(f)
		if readErr != nil {
			return "", fmt.Errorf("teatest: read %s: %w", f, readErr)
		}
		fmt.Fprintf(h, "FILE %s %d\n", filepath.Base(f), len(b))
		h.Write(b)
	}

	for _, mf := range []string{"go.mod", "go.sum"} {
		path := filepath.Join(moduleRoot, mf)
		if b, readErr := os.ReadFile(path); readErr == nil {
			fmt.Fprintf(h, "MOD %s %d\n", mf, len(b))
			h.Write(b)
		}
	}
	return hex.EncodeToString(h.Sum(nil))[:32], nil
}

func goFilesIn(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("teatest: read package dir: %w", err)
	}
	var out []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasSuffix(name, ".go") {
			continue
		}
		if name == harnessFileName {
			continue
		}
		out = append(out, filepath.Join(dir, name))
	}
	return out, nil
}

func detectPackageName(dir string) (string, error) {
	files, err := goFilesIn(dir)
	if err != nil {
		return "", err
	}
	for _, f := range files {
		name, ok := readPackageName(f)
		if ok {
			return name, nil
		}
	}
	return "", fmt.Errorf("teatest: no Go files with a package clause found in %s", dir)
}

func readPackageName(path string) (string, bool) {
	b, err := os.ReadFile(path)
	if err != nil {
		return "", false
	}
	for line := range strings.SplitSeq(string(b), "\n") {
		t := strings.TrimSpace(line)
		if strings.HasPrefix(t, "package ") {
			parts := strings.Fields(t)
			// "package <name>" — need at least two tokens.
			const minPackageTokens = 2
			if len(parts) >= minPackageTokens {
				return strings.TrimSuffix(parts[1], ";"), true
			}
		}
	}
	return "", false
}

func findModuleRoot(start string) (string, error) {
	dir := start
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("teatest: no go.mod found above %s — package must live inside a Go module", start)
		}
		dir = parent
	}
}

func binExt() string {
	if isWindows() {
		return ".exe"
	}
	return ""
}

func isWindows() bool {
	return os.PathSeparator == '\\'
}
