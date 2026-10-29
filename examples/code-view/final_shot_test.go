package main

import (
	"context"
	"image"
	"image/png"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/chinmay-sawant/ownframe"
	"github.com/hajimehoshi/ebiten/v2"
)

type finalShotGame struct {
	inner ebiten.Game
	tick  int
	path  string
	done  bool
}

func (g *finalShotGame) Layout(w, h int) (int, int) { return g.inner.Layout(w, h) }
func (g *finalShotGame) Update() error {
	if err := g.inner.Update(); err != nil {
		return err
	}
	g.tick++
	if g.done || g.tick >= 30 {
		return ebiten.Termination
	}
	return nil
}
func (g *finalShotGame) Draw(dst *ebiten.Image) {
	g.inner.Draw(dst)
	if g.done {
		return
	}
	g.done = true
	w, h := dst.Bounds().Dx(), dst.Bounds().Dy()
	pix := make([]byte, w*h*4)
	dst.ReadPixels(pix)
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	copy(img.Pix, pix)
	f, err := os.Create(g.path)
	if err != nil {
		return
	}
	_ = png.Encode(f, img)
	_ = f.Close()
}

func finalShot(t *testing.T, app *App, path string) {
	t.Helper()
	ebiten.SetWindowSize(1500, 900)
	g := &finalShotGame{inner: ownframe.NewGame(context.Background(), app.Page()), path: path}
	if err := ebiten.RunGame(g); err != nil && err != ebiten.Termination {
		t.Fatalf("RunGame: %v", err)
	}
}

func finalFileApp(t *testing.T, zoom float64) (*App, context.Context) {
	t.Helper()
	src, err := os.ReadFile("/home/chinmay/ChinmayPersonalProjects/ownframe/documentation/interaction.md")
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "interaction.md"), src, 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "a.txt"), []byte("x\n"), 0644); err != nil {
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
	src, err := os.ReadFile("/home/chinmay/ChinmayPersonalProjects/ownframe/documentation/screen.md")
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "screen.md"), src, 0644); err != nil {
		t.Fatal(err)
	}
	runGit := func(args ...string) {
		t.Helper()
		c := exec.Command("git", args...)
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
	if err := os.WriteFile(filepath.Join(dir, "screen.md"), []byte(strings.Join(lines, "\n")), 0644); err != nil {
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

func TestFinalFile100(t *testing.T) {
	app, _ := finalFileApp(t, 1)
	finalShot(t, app, "/tmp/opencode/final-file-100.png")
}

func TestFinalFile87(t *testing.T) {
	app, _ := finalFileApp(t, 0.87)
	finalShot(t, app, "/tmp/opencode/final-file-87.png")
}

func TestFinalDiff100(t *testing.T) {
	app, _ := finalDiffApp(t, 1)
	finalShot(t, app, "/tmp/opencode/final-diff-100.png")
}

func TestFinalDiff87(t *testing.T) {
	app, _ := finalDiffApp(t, 0.87)
	finalShot(t, app, "/tmp/opencode/final-diff-87.png")
}
