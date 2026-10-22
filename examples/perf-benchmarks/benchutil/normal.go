package benchutil

import (
	"context"

	"github.com/chinmay-sawant/ownframe"
)

// NormalView is the Benchmark A data: a click counter plus a bound name.
type NormalView struct {
	Count int
	Name  string
}

// Normal is the small mostly-idle desktop form of Benchmark A.
type Normal struct {
	page *ownframe.Page
	view NormalView
}

// NewNormal parses the form page and wires its two buttons.
func NewNormal() (*Normal, error) {
	page, err := ownframe.New(ownframe.Config{DevTools: true,
		Title: "Normal", HTML: normalHTML,
		Width: normalWidth, Height: normalHeight, MinWidth: normalMinWidth, MinHeight: normalMinHeight,
		Perf: true,
	})
	if err != nil {
		return nil, err
	}

	n := &Normal{page: page}
	page.Handle(ownframe.Handlers{Click: n.onClick})
	page.SetData(&n.view)

	return n, nil
}

// Page returns the page Run and Serve display.
func (n *Normal) Page() *ownframe.Page { return n.page }

// Redraw renders the current view.
func (n *Normal) Redraw(ctx context.Context) error { return n.page.Redraw(ctx) }

// onClick bumps the counter on #inc and clears it on #reset.
func (n *Normal) onClick(_ context.Context, box ownframe.Box) error {
	switch box.ID {
	case "inc":
		n.view.Count++
	case "reset":
		n.view.Count = 0
	default:
		return nil
	}

	n.page.SetData(&n.view)
	n.page.Testing().Invalidate("count")

	return nil
}
