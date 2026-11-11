package page_test

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/ownframe/internal/page"
)

// newCachePage builds a page for the cache tests.
func newCachePage(t *testing.T, html string) *page.Page {
	t.Helper()

	p, err := page.New(page.Config{HTML: html, Width: 320, Height: 200})
	if err != nil {
		t.Fatal(err)
	}

	return p
}

func TestResizeReusesParseAndCascade(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	p := newCachePage(t, `<p>hello cache</p>`)

	for i := 0; i < 10; i++ {
		p.SetSize(320+i, 200+i)
		if err := p.Redraw(ctx); err != nil {
			t.Fatal(err)
		}
	}

	stats := p.Stats()
	if stats.Parses != 1 || stats.Cascades != 1 {
		t.Fatalf("parses = %d, cascades = %d, want 1 and 1", stats.Parses, stats.Cascades)
	}
	if stats.Redraws != 10 || stats.Repaints != 10 || stats.Layouts != 10 {
		t.Fatalf("redraws = %d, repaints = %d, layouts = %d, want 10 each",
			stats.Redraws, stats.Repaints, stats.Layouts)
	}
	if stats.LastRedraw <= 0 || stats.LastDraw <= 0 {
		t.Fatalf("last redraw = %v, last draw = %v", stats.LastRedraw, stats.LastDraw)
	}
}

type resizeTemplateData struct {
	Reads int
	Value string
}

func (d *resizeTemplateData) Text() string {
	d.Reads++
	return d.Value
}

func TestResizeKeepsCleanTemplateButRelayouts(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	p := newCachePage(t, `<style>#a{width:40px}@media(min-width:700px){#a{width:120px}}</style><p id="a">{{.Text}}</p>`)
	data := &resizeTemplateData{Value: "before"}
	p.SetData(data)
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	reads := data.Reads
	for _, width := range []int{800, 400, 800} {
		p.SetSize(width, 200)
		if err := p.Redraw(ctx); err != nil {
			t.Fatal(err)
		}
		if data.Reads != reads {
			t.Fatalf("resize executed template %d extra times", data.Reads-reads)
		}
		want := 40.0
		if width == 800 {
			want = 120
		}
		if got := viewportBox(t, p, "a").W; got != want {
			t.Fatalf("width %d: media-query box = %g, want %g", width, got, want)
		}
	}
	data.Value = "after"
	p.SetData(data)
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	if data.Reads != reads+1 || viewportBox(t, p, "a").Text != "after" {
		t.Fatal("SetData did not refresh template output after resizing")
	}
}

func TestResizeStillPublishesChangedScrollWindow(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name     string
		reported bool
	}{
		{name: "requests redraw", reported: true},
		{name: "declines redraw", reported: false},
	} {
		t.Run(test.name, func(t *testing.T) {
			p := newCachePage(t, `<p id="a">{{.Text}}</p>`)
			data := &resizeTemplateData{Value: "before"}
			p.SetData(data)
			p.SetWindowing(true)
			lastHeight := 0
			p.SetScrollWindow(func(_, height int) bool {
				if height == lastHeight {
					return false
				}
				lastHeight = height
				data.Value = "initial window"
				if height == 400 {
					data.Value = "resized window"
				}
				return test.reported
			})
			if err := p.Redraw(t.Context()); err != nil {
				t.Fatal(err)
			}
			reads := data.Reads
			p.SetSize(320, 400)
			if err := p.Redraw(t.Context()); err != nil {
				t.Fatal(err)
			}
			if data.Reads != reads+1 || viewportBox(t, p, "a").Text != "resized window" {
				t.Fatal("changed scroll window did not invalidate template output")
			}
		})
	}
}

func TestHoverRelayoutKeepsCache(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	p := newCachePage(t, `<style>#a:hover{color:#ff0000}</style><p id="a">hover</p>`)
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	x, y := boxCenter(t, p, "a")
	if err := p.Hover(ctx, x, y); err != nil {
		t.Fatal(err)
	}

	stats := p.Stats()
	if stats.Parses != 1 || stats.Cascades != 1 {
		t.Fatalf("parses = %d, cascades = %d, want 1 and 1", stats.Parses, stats.Cascades)
	}
	if stats.Layouts != 2 {
		t.Fatalf("layouts = %d, want 2", stats.Layouts)
	}
}
