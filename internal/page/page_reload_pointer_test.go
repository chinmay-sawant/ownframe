package page

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func writeWatchFile(t *testing.T, body string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "index.html")
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}

	return path
}

func rewriteWatchFile(t *testing.T, path, body string) {
	t.Helper()

	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
}

func centerOf(t *testing.T, p *Page, id string) (float64, float64) {
	t.Helper()

	for _, b := range p.Boxes() {
		if b.ID == id && b.W > 0 && b.H > 0 {
			return b.X + b.W/2, b.Y + b.H/2
		}
	}

	t.Fatalf("no box id=%q", id)

	return 0, 0
}

func TestReloadClearsStaleHoverActive(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	path := writeWatchFile(t, `<div id="a">A</div><div id="b">B</div>`)
	p, err := New(Config{File: path, Width: 320, Height: 200})
	if err != nil {
		t.Fatal(err)
	}

	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	ax, ay := centerOf(t, p, "a")
	if err := p.Hover(ctx, ax, ay); err != nil {
		t.Fatal(err)
	}

	bx, by := centerOf(t, p, "b")
	if err := p.Press(ctx, bx, by); err != nil {
		t.Fatal(err)
	}

	rewriteWatchFile(t, path, `<div id="b">B</div>`)

	if _, err := p.PollReload(ctx); err != nil {
		t.Fatal(err)
	}

	if p.HoverID() != "" {
		t.Fatalf("hover = %q, want none", p.HoverID())
	}

	if p.activeID() != "b" {
		t.Fatalf("active = %q, want b", p.activeID())
	}
}
