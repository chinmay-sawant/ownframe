package page_test

import (
	"context"
	"github.com/chinmay-sawant/ownframe/internal/page"
	"os"
	"testing"
	"time"
)

func TestReloadDoesNotFireChange(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	path := writeSource(t, "index.html", `<input id="note" type="text">`)
	p := newFilePage(t, page.Config{File: path, Width: 320, Height: 200})

	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	changes := 0
	p.Handle(page.Handlers{Change: func(context.Context, page.Box) error {
		changes++

		return nil
	}})

	rewriteSource(t, path, `<input id="note" type="text"><p id="extra">extra</p>`)

	if _, err := p.PollReload(ctx); err != nil {
		t.Fatal(err)
	}

	if changes != 0 {
		t.Fatalf("changes = %d, want none for a reload", changes)
	}
}

func TestPollReloadDeletedFileKeepsPicture(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	path := writeSource(t, "index.html", `<p id="p">one</p>`)
	p := watchedPage(t, path)
	gen := p.Generation()

	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}

	changed, err := p.PollReload(ctx)
	if changed || err == nil {
		t.Fatalf("first poll = %v, %v, want a stat error", changed, err)
	}

	if p.Generation() != gen || p.Display() == nil {
		t.Fatal("the last good picture was dropped")
	}

	changed, err = p.PollReload(ctx)
	if changed || err != nil {
		t.Fatalf("second poll = %v, %v, want the error reported once", changed, err)
	}

	rewriteSource(t, path, `<p id="p">back, longer</p>`)

	changed, err = p.PollReload(ctx)
	if !changed || err != nil {
		t.Fatalf("recreated poll = %v, %v", changed, err)
	}

	if p.HTML() != `<p id="p">back, longer</p>` {
		t.Fatalf("HTML = %q", p.HTML())
	}
}

func TestReloadKeepsHistory(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	path := writeSource(t, "index.html", `<p id="p">one</p>`)
	p := watchedPage(t, path)

	if err := p.Load(ctx, `<p id="p">two</p>`); err != nil {
		t.Fatal(err)
	}

	rewriteSource(t, path, `<p id="p">three</p>`)

	changed, err := p.PollReload(ctx)
	if !changed || err != nil {
		t.Fatalf("poll = %v, %v", changed, err)
	}

	if got := p.HTML(); got != `<p id="p">three</p>` {
		t.Fatalf("html = %q", got)
	}

	if err := p.Back(ctx); err != nil {
		t.Fatal(err)
	}

	if got := p.HTML(); got != `<p id="p">one</p>` {
		t.Fatalf("back = %q", got)
	}
}

func TestDisableHotReloadConfig(t *testing.T) {
	t.Parallel()

	path := writeSource(t, "index.html", `<p>one</p>`)
	p := newFilePage(t, page.Config{File: path, Width: 320, Height: 200, DisableHotReload: true})

	if p.Watching() {
		t.Fatal("watch is on with DisableHotReload")
	}
}

func TestReloadStatsCount(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	path := writeSource(t, "index.html", `<p id="p">one</p>`)
	p := watchedPage(t, path)

	if got := p.Stats().Reloads; got != 0 {
		t.Fatalf("reloads = %d, want 0", got)
	}

	rewriteSource(t, path, `<p id="p">two</p>`)

	if changed, err := p.PollReload(ctx); !changed || err != nil {
		t.Fatalf("poll = %v, %v", changed, err)
	}

	st := p.Stats()
	if st.Reloads != 1 || st.LastReloadError != "" {
		t.Fatalf("stats = %d reloads, error %q", st.Reloads, st.LastReloadError)
	}

	rewriteSource(t, path, `{{`)

	if _, err := p.PollReload(ctx); err == nil {
		t.Fatal("a broken template did not report an error")
	}

	if p.Stats().LastReloadError == "" {
		t.Fatal("reload error not recorded")
	}
}

func TestReloadKeepsTypedField(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	path := writeSource(t, "index.html", `<input id="note" type="text"><p id="p">before</p>`)
	p := watchedPage(t, path)

	x, y := boxCenter(t, p, "note")
	if err := p.Click(ctx, x, y); err != nil {
		t.Fatal(err)
	}

	if err := p.Type(ctx, "hello"); err != nil {
		t.Fatal(err)
	}

	gen := p.Generation()
	next := `<input id="note" type="text"><p id="p">after</p><p id="extra">an extra paragraph</p>`
	rewriteSource(t, path, next)

	changed, err := p.PollReload(ctx)
	if !changed || err != nil {
		t.Fatalf("poll = %v, %v", changed, err)
	}

	if p.FormValue("note") != "hello" {
		t.Fatalf("value = %q", p.FormValue("note"))
	}

	if p.FocusedField() != "note" {
		t.Fatalf("focus = %q", p.FocusedField())
	}

	if p.Generation() != gen+1 {
		t.Fatalf("generation moved by %d", p.Generation()-gen)
	}
}

func TestReloadDropsRemovedFieldFocus(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	path := writeSource(t, "index.html", `<input id="note" type="text">`)
	p := watchedPage(t, path)

	x, y := boxCenter(t, p, "note")
	if err := p.Click(ctx, x, y); err != nil {
		t.Fatal(err)
	}

	rewriteSource(t, path, `<p id="p">no field</p>`)

	changed, err := p.PollReload(ctx)
	if !changed || err != nil {
		t.Fatalf("poll = %v, %v", changed, err)
	}

	if p.FocusedField() != "" {
		t.Fatalf("focus = %q, want none", p.FocusedField())
	}
}

func TestSetHotReloadOff(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	path := writeSource(t, "index.html", `<p id="p">one</p>`)
	p := watchedPage(t, path)

	if !p.Watching() {
		t.Fatal("file page does not watch")
	}

	p.SetHotReload(false)
	if p.Watching() {
		t.Fatal("watch stayed on")
	}

	rewriteSource(t, path, `<p id="p">two</p>`)

	changed, err := p.PollReload(ctx)
	if changed || err != nil {
		t.Fatalf("off poll = %v, %v", changed, err)
	}

	p.SetHotReload(true)

	changed, err = p.PollReload(ctx)
	if !changed || err != nil {
		t.Fatalf("on poll = %v, %v", changed, err)
	}
}

func TestStringPageDoesNotWatch(t *testing.T) {
	t.Parallel()

	p := newFilePage(t, page.Config{HTML: `<p id="p">one</p>`, Width: 320, Height: 200})

	if p.Watching() {
		t.Fatal("string page watches")
	}

	changed, err := p.PollReload(context.Background())
	if changed || err != nil {
		t.Fatalf("poll = %v, %v", changed, err)
	}
}

func TestPollReloadSecondEditWins(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	path := writeSource(t, "index.html", `<p id="p">one</p>`)
	p := watchedPage(t, path)
	gen := p.Generation()

	rewriteSource(t, path, `<p id="p">two</p>`)
	rewriteSource(t, path, `<p id="p">three, longer</p>`)

	changed, err := p.PollReload(ctx)
	if !changed || err != nil {
		t.Fatalf("poll = %v, %v", changed, err)
	}

	if p.HTML() != `<p id="p">three, longer</p>` {
		t.Fatalf("HTML = %q", p.HTML())
	}

	if p.Generation() != gen+1 {
		t.Fatalf("generation moved by %d, want one redraw", p.Generation()-gen)
	}
}

func TestPollReloadPendingRetry(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	good := `<p id="p">good</p>`
	path := writeSource(t, "index.html", good)
	p := watchedPage(t, path)
	gen := p.Generation()

	rewriteSource(t, path, `<p id="p">{{.Broken</p>`)

	changed, err := p.PollReload(ctx)
	if changed || err == nil {
		t.Fatalf("bad poll = %v, %v, want a parse error", changed, err)
	}

	if p.HTML() != good {
		t.Fatalf("HTML = %q, want the last good source", p.HTML())
	}

	if p.Generation() != gen {
		t.Fatal("a parse error redrew the page")
	}

	changed, err = p.PollReload(ctx)
	if changed || err == nil {
		t.Fatalf("retry poll = %v, %v, want the same parse error", changed, err)
	}

	rewriteSource(t, path, `<p id="p">fixed</p>`)

	changed, err = p.PollReload(ctx)
	if !changed || err != nil {
		t.Fatalf("fixed poll = %v, %v", changed, err)
	}

	if p.HTML() != `<p id="p">fixed</p>` {
		t.Fatalf("HTML = %q", p.HTML())
	}

	if p.Generation() != gen+1 {
		t.Fatal("the fix did not redraw exactly once")
	}
}

// rewriteSource replaces a watched file and bumps the mtime when the stat
// did not change.
func rewriteSource(t *testing.T, path, body string) {
	t.Helper()

	before, err := os.Stat(path)
	if err != nil {
		before = nil
	}

	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}

	after, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}

	if before != nil && before.ModTime().Equal(after.ModTime()) && before.Size() == after.Size() {
		stamp := after.ModTime().Add(time.Second)
		if err := os.Chtimes(path, stamp, stamp); err != nil {
			t.Fatal(err)
		}
	}
}

// watchedPage opens a file-backed page and draws it once.
func watchedPage(t *testing.T, path string) *page.Page {
	t.Helper()

	p := newFilePage(t, page.Config{File: path, Width: 320, Height: 200})
	if err := p.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}

	return p
}

func TestPollReloadNoChangeDoesNotRedraw(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	p := watchedPage(t, writeSource(t, "index.html", `<p id="p">one</p>`))
	gen := p.Generation()

	for i := 0; i < 2; i++ {
		changed, err := p.PollReload(ctx)
		if changed || err != nil {
			t.Fatalf("poll = %v, %v", changed, err)
		}
	}

	if p.Generation() != gen {
		t.Fatalf("generation = %d, want %d", p.Generation(), gen)
	}
}

func TestPollReloadOneChangedByteRedrawsOnce(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	path := writeSource(t, "index.html", `<p id="p">one</p>`)
	p := watchedPage(t, path)
	gen := p.Generation()

	rewriteSource(t, path, `<p id="p">ono</p>`)

	changed, err := p.PollReload(ctx)
	if !changed || err != nil {
		t.Fatalf("poll = %v, %v", changed, err)
	}

	if p.HTML() != `<p id="p">ono</p>` {
		t.Fatalf("HTML = %q", p.HTML())
	}

	if p.Generation() != gen+1 {
		t.Fatalf("generation = %d, want one redraw", p.Generation()-gen)
	}
}

func newThemedFilePage(t *testing.T, path string) *page.Page {
	t.Helper()

	p := newFilePage(t, page.Config{
		HTML:      themePage,
		ThemeFile: path,
		Width:     320,
		Height:    200,
	})
	if err := p.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}

	return p
}

func TestPollReloadThemeFile(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	path := writeSource(t, "theme.css", `#swatch{background:#ff0000}`)
	p := newThemedFilePage(t, path)

	if r, g, b := swatchRGB(t, p); r < 1 || g > 0 || b > 0 {
		t.Fatalf("swatch = %v,%v,%v, want red", r, g, b)
	}

	gen := p.Generation()
	rewriteSource(t, path, `#swatch{background:#00ff00}`)

	changed, err := p.PollReload(ctx)
	if !changed || err != nil {
		t.Fatalf("poll = %v, %v", changed, err)
	}

	if r, g, b := swatchRGB(t, p); g < 1 || r > 0 || b > 0 {
		t.Fatalf("swatch = %v,%v,%v, want green", r, g, b)
	}

	if p.Generation() != gen+1 {
		t.Fatalf("generation moved by %d", p.Generation()-gen)
	}
}

func TestPollReloadThemeErrorKeepsOld(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	path := writeSource(t, "theme.css", `#swatch{background:#ff0000}`)
	p := newThemedFilePage(t, path)
	gen := p.Generation()

	rewriteSource(t, path, `}`)

	changed, err := p.PollReload(ctx)
	if changed || err == nil {
		t.Fatalf("bad theme poll = %v, %v, want a parse error", changed, err)
	}

	if r, g, b := swatchRGB(t, p); r < 1 || g > 0 || b > 0 {
		t.Fatalf("swatch = %v,%v,%v, want the old red", r, g, b)
	}

	if p.Generation() != gen {
		t.Fatal("a broken theme redrew the page")
	}
}
