package ui

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/chinmay-sawant/ownframe"
)

// seedEntries builds n entries with IDs 1..n.
func seedEntries(n int) []Entry {
	out := make([]Entry, n)
	for i := range out {
		out[i] = Entry{
			ID: int64(i + 1), TimeOK: true, Time: time.Unix(int64(i), 0),
			Source: "app.log", SourceKey: "s1", Severity: "info",
			Text: "entry " + strings.Repeat("m", i%5), Lines: 1,
		}
	}

	return out
}

// newHeadless builds an App with no feed and 100 loaded entries.
func newHeadless(t *testing.T) *App {
	t.Helper()

	a, err := New(Options{})
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { _ = a.Close() })

	a.pager.Load(PageResult{Entries: seedEntries(100), Total: 100, HasOlder: true})
	a.follow.Seen(100)

	return a
}

// boxByID returns the boxes for one id.
func boxByID(boxes []ownframe.Box, id string) (ownframe.Box, bool) {
	for _, b := range boxes {
		if b.ID == id {
			return b, true
		}
	}

	return ownframe.Box{}, false
}

func TestRenderListWindow(t *testing.T) {
	a := newHeadless(t)
	ctx := context.Background()

	a.Pin(0, 400)
	if err := a.draw(ctx); err != nil {
		t.Fatal(err)
	}

	boxes := a.page.Boxes()
	rows := 0

	for _, b := range boxes {
		if strings.HasPrefix(b.ID, "row-") {
			rows++

			if b.H != RowH {
				t.Errorf("row %s height = %v, want %d", b.ID, b.H, RowH)
			}
		}
	}

	if rows == 0 {
		t.Fatal("no rows rendered")
	}

	top, ok := boxByID(boxes, "spacer-top")
	if !ok || top.H != HeaderH {
		t.Fatalf("spacer-top = %#v", top)
	}
}

func TestRenderScrollSlicesWindow(t *testing.T) {
	a := newHeadless(t)
	ctx := context.Background()

	a.page.Scroll().SetScrollOffset(0, 30*RowH)
	if err := a.draw(ctx); err != nil {
		t.Fatal(err)
	}

	if len(a.view.Rows) == 0 || a.view.Rows[0].ID != 21 {
		t.Fatalf("first row = %#v", a.view.Rows[0])
	}

	if a.view.TopPad != HeaderH+20*RowH {
		t.Fatalf("top pad = %d", a.view.TopPad)
	}
}
