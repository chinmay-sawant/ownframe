package benchutil

import (
	"context"
	"strconv"

	"github.com/chinmay-sawant/ownframe"
)

// Row is one Benchmark B list row.
type Row struct {
	Num   string
	Title string
	Body  string
}

// LargeView is the Benchmark B data: a generated text-heavy list.
type LargeView struct {
	Total int
	Rows  []Row
}

// Large is the thousand-row scrolling page of Benchmark B.
type Large struct {
	page *ownframe.Page
	view LargeView
}

// NewLarge builds a page with n generated rows.
func NewLarge(n int) (*Large, error) {
	page, err := ownframe.New(ownframe.Config{DevTools: true,
		Title: "Large", HTML: largeHTML,
		Width: 800, Height: 600, MinWidth: 320, MinHeight: 240,
		Perf: true,
	})
	if err != nil {
		return nil, err
	}

	l := &Large{page: page, view: makeView(n)}
	page.SetData(l.view)

	return l, nil
}

// Page returns the page Run and Serve display.
func (l *Large) Page() *ownframe.Page { return l.page }

// Redraw renders the list.
func (l *Large) Redraw(ctx context.Context) error { return l.page.Redraw(ctx) }

// SetSize stores the frame size used by the next Redraw.
func (l *Large) SetSize(width, height int) { l.page.SetSize(width, height) }

// makeView generates n text-heavy rows.
func makeView(n int) LargeView {
	v := LargeView{Total: n, Rows: make([]Row, n)}

	for i := range v.Rows {
		v.Rows[i] = Row{
			Num:   "#" + strconv.Itoa(i+1),
			Title: "Item " + strconv.Itoa(i+1),
			Body:  "Lorem ipsum dolor sit amet, consectetur adipiscing elit.",
		}
	}

	return v
}
