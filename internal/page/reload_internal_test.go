package page

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestReloadKeepsTickAndImages(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	path := writeWatchFile(t, `<p id="p">one</p>`)
	p, err := New(Config{File: path, Width: 320, Height: 200})
	if err != nil {
		t.Fatal(err)
	}

	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	ticks := 0
	p.SetTick(func(context.Context) error {
		ticks++

		return nil
	})
	p.SetImage("logo", []byte("bytes"))

	rewriteWatchFile(t, path, `<p id="p">two, longer</p>`)

	if _, err := p.PollReload(ctx); err != nil {
		t.Fatal(err)
	}

	if err := p.Tick(ctx); err != nil {
		t.Fatal(err)
	}

	if ticks != 1 {
		t.Fatalf("ticks = %d, want 1", ticks)
	}

	data, err := imageBytes(p, "logo")
	if err != nil || string(data) != "bytes" {
		t.Fatalf("image = %q, %v", data, err)
	}
}

func TestPollReloadMemorySource(t *testing.T) {
	ctx := context.Background()
	fake := newMemSource()
	path := "mem://one.html"
	fake.write(path, []byte(`<p id="p">one</p>`))
	installSourceReader(fake)
	defer installSourceReader(nil)
	p, err := New(Config{File: path, Width: 320, Height: 200})
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	gen := p.Generation()
	if changed, err := p.PollReload(ctx); changed || err != nil {
		t.Fatalf("clean poll = %v, %v", changed, err)
	}
	fake.write(path, []byte(`<p id="p">two words longer</p>`))
	changed, err := p.PollReload(ctx)
	if !changed || err != nil {
		t.Fatalf("dirty poll = %v, %v", changed, err)
	}
	if p.HTML() != `<p id="p">two words longer</p>` {
		t.Fatalf("HTML = %q", p.HTML())
	}
	if p.Generation() != gen+1 {
		t.Fatalf("gen = %d, want %d", p.Generation(), gen+1)
	}
}

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

	if activeID(p) != "b" {
		t.Fatalf("active = %q, want b", activeID(p))
	}
}
