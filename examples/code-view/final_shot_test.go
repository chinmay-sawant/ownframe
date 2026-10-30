package main

import (
	"context"
	"errors"
	"image"
	"image/png"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/chinmay-sawant/ownframe"
	"github.com/hajimehoshi/ebiten/v2"
)

const (
	finalShotWidth  = 1500
	finalShotHeight = 900
)

// finalShotSuite captures every final shot in one Ebiten session. Ebiten
// allows a single RunGame per process, so four separate RunGame tests panic
// with "MarkGuest is called more than once" on the second test. One suite
// also keeps `make test` (xvfb) and the display-less examples-race shards
// on the same skip path as the other screenshot helpers.
type finalShotSuite struct {
	inners     []ebiten.Game
	paths      []string
	index      int
	captured   bool
	captureErr error
}

func (g *finalShotSuite) Layout(_, _ int) (int, int) { return finalShotWidth, finalShotHeight }
func (g *finalShotSuite) Update() error {
	if g.index >= len(g.inners) {
		return ebiten.Termination
	}
	if err := g.inners[g.index].Update(); err != nil {
		return err
	}
	if g.captured {
		g.index++
		g.captured = false
		if g.index >= len(g.inners) {
			return ebiten.Termination
		}
	}
	return nil
}
func (g *finalShotSuite) Draw(dst *ebiten.Image) {
	if g.index >= len(g.inners) {
		return
	}
	g.inners[g.index].Draw(dst)
	if g.captured || g.captureErr != nil {
		return
	}
	g.captured = true
	g.captureErr = writeFinalShot(dst, g.paths[g.index])
}

func writeFinalShot(dst *ebiten.Image, path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	w, h := dst.Bounds().Dx(), dst.Bounds().Dy()
	pix := make([]byte, w*h*4)
	dst.ReadPixels(pix)
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	copy(img.Pix, pix)
	f, err := os.Create(path) //nolint:gosec // test screenshot helper writes a fixed /tmp fixture path
	if err != nil {
		return err
	}
	if err := png.Encode(f, img); err != nil {
		_ = f.Close()
		return err
	}
	return f.Close()
}

func finalShotSuiteRun(t *testing.T, apps []*App, paths []string) {
	t.Helper()
	ebiten.SetWindowSize(finalShotWidth, finalShotHeight)
	ctx := context.Background()
	inners := make([]ebiten.Game, len(apps))
	for i, app := range apps {
		inners[i] = ownframe.NewGame(ctx, app.Page())
	}
	g := &finalShotSuite{inners: inners, paths: paths}
	if err := ebiten.RunGame(g); err != nil && !errors.Is(err, ebiten.Termination) {
		t.Fatalf("RunGame: %v", err)
	}
	if g.captureErr != nil {
		t.Fatalf("capture: %v", g.captureErr)
	}
	if g.index != len(inners) {
		t.Fatalf("captured %d of %d shots", g.index, len(inners))
	}
}

func finalDoc(t *testing.T, name string) []byte {
	t.Helper()
	clean := filepath.Clean(name)
	if clean != name || clean == "." || clean == ".." || strings.ContainsAny(clean, `/\`) {
		t.Fatalf("finalDoc: bad doc name %q", name)
	}
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller: no caller information")
	}
	src, err := os.ReadFile(filepath.Join(filepath.Dir(file), "..", "..", "documentation", clean))
	if err != nil {
		t.Fatal(err)
	}
	return src
}

func finalFileApp(t *testing.T, zoom float64) (*App, context.Context) {
	t.Helper()
	src := finalDoc(t, "interaction.md")
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "interaction.md"), src, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "a.txt"), []byte("x\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	app, err := New(dir)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	app.openFile("interaction.md")
	app.SetDevTools(false)
	if app.wrap {
		app.toggleWrap()
	}
	if err := app.setZoom(ctx, zoom); err != nil {
		t.Fatal(err)
	}
	if err := app.Page().Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	return app, ctx
}

func finalDiffApp(t *testing.T, zoom float64) (*App, context.Context) {
	t.Helper()
	src := finalDoc(t, "screen.md")
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "screen.md"), src, 0o600); err != nil {
		t.Fatal(err)
	}
	runGit := func(args ...string) {
		t.Helper()
		c := exec.CommandContext(context.Background(), "git", args...) //nolint:gosec // test helper runs the fixed git binary with literal subcommands on a temp repo
		c.Dir = dir
		c.Env = append(os.Environ(), "GIT_CONFIG_NOSYSTEM=1", "HOME="+dir)
		if out, err := c.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v %s", args, err, out)
		}
	}
	runGit("init", "-q")
	runGit("config", "user.email", "t@t.t")
	runGit("config", "user.name", "t")
	runGit("add", ".")
	runGit("commit", "-qm", "init")
	lines := strings.Split(string(src), "\n")
	lines[33] = strings.Replace(lines[33], "picks the mode", "picks the display mode", 1)
	lines[33] = strings.Replace(lines[33], "keeps the display list", "keeps the full display list", 1)
	if err := os.WriteFile(filepath.Join(dir, "screen.md"), []byte(strings.Join(lines, "\n")), 0o600); err != nil {
		t.Fatal(err)
	}
	app, err := New(dir)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	app.reloadChanges(ctx, false)
	app.SetDevTools(false)
	if err := app.onClick(ctx, ownframe.Box{Action: "mode-changes"}); err != nil {
		t.Fatal(err)
	}
	if app.wrap {
		app.toggleWrap()
	}
	if err := app.setZoom(ctx, zoom); err != nil {
		t.Fatal(err)
	}
	if err := app.Page().Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	return app, ctx
}

// TestFinalShots captures the four review fixtures in one RunGame session.
// It stays out of the default run like the other screenshot helpers: set
// OWNFRAME_CODE_VIEW_SCREENSHOT=1 and run under xvfb (or a desktop session).
func TestFinalShots(t *testing.T) {
	if os.Getenv("OWNFRAME_CODE_VIEW_SCREENSHOT") != "1" {
		t.Skip("set OWNFRAME_CODE_VIEW_SCREENSHOT=1 to capture the final shots")
	}
	if runtime.GOOS == "linux" && os.Getenv("DISPLAY") == "" && os.Getenv("WAYLAND_DISPLAY") == "" {
		t.Skip("needs a display (DISPLAY or WAYLAND_DISPLAY); run under xvfb-run")
	}
	file100, _ := finalFileApp(t, 1)
	file87, _ := finalFileApp(t, 0.87)
	diff100, _ := finalDiffApp(t, 1)
	diff87, _ := finalDiffApp(t, 0.87)
	finalShotSuiteRun(t,
		[]*App{file100, file87, diff100, diff87},
		[]string{
			"/tmp/opencode/final-file-100.png",
			"/tmp/opencode/final-file-87.png",
			"/tmp/opencode/final-diff-100.png",
			"/tmp/opencode/final-diff-87.png",
		},
	)
}
