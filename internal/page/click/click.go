// Package click runs the page click sequence.
// It does not import package page.
package click

import (
	"context"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/fetch"
	"github.com/chinmay-sawant/ownframe/internal/page/form"
)

// Surface is the page state one click reads and the actions it takes.
// The page adapter delegates each method. This package does not import page.
type Surface interface {
	HitBox(x, y float64) (layout.Box, bool)
	BoxIDAt(x, y float64) string
	BoxByID(id string) layout.Box
	ConsumePress(id string) bool
	FormControl(id string) bool
	Control(id string) (form.Control, bool)
	Lookup(action string) (string, bool)
	Load(ctx context.Context, html string) error
	ClickControl(ctx context.Context, box layout.Box) error
	BlurForm()
	HasOpenLink() bool
	OpenLinkAt(ctx context.Context, x, y float64) (bool, error)
	HasClick() bool
	HandlerClick(ctx context.Context, box layout.Box) error
	ClickPainted(ctx context.Context, box layout.Box) (bool, error)
	Redraw(ctx context.Context) error
	RedrawIfChanged(ctx context.Context) error
}

// Run hit-tests the last picture.
// The innermost box is the last one in document order that contains the point.
// A box without an id falls back to the innermost id-bearing element under the
// point, so a click on a child icon reaches its control.
// A field is activated. An enabled button can follow a route.
// A disabled control is not toggled and does not blur.
// Any other hit blurs the form. A routed data-action then loads that HTML
// and skips the handler. Run draws the page again.
func Run(ctx context.Context, s Surface, x, y float64) error {
	if err := useContext(ctx); err != nil {
		return err
	}

	box, ok := s.HitBox(x, y)
	if ok && box.ID == "" {
		if id := s.BoxIDAt(x, y); id != "" {
			box = s.BoxByID(id)
		}
	}
	if ok && s.ConsumePress(box.ID) {
		return s.Redraw(ctx)
	}

	if ok && s.FormControl(box.ID) {
		c, _ := s.Control(box.ID)
		if c.Tag == "button" && !c.Disabled && box.Action != "" {
			if html, ok := s.Lookup(box.Action); ok {
				return s.Load(ctx, html)
			}
		}

		return s.ClickControl(ctx, box)
	}

	if ok {
		s.BlurForm()
		if box.Action != "" {
			if html, ok := s.Lookup(box.Action); ok {
				return s.Load(ctx, html)
			}
		}
	}

	if s.HasOpenLink() && (!ok || box.Action == "") {
		if handled, err := s.OpenLinkAt(ctx, x, y); handled {
			return err
		}
	}

	if ok && s.HasClick() {
		if err := s.HandlerClick(ctx, box); err != nil {
			return err
		}
	}

	if ok {
		painted, err := s.ClickPainted(ctx, box)
		if err != nil {
			return err
		}
		if painted {
			return nil
		}
	}

	// The handler may have changed the data, which makes the source differ
	// and redraws. A click that changed nothing visible skips the layout. A
	// refused patch forgot the render key, so this redraw lays the page out
	// again and the tree matches the display.
	return s.RedrawIfChanged(ctx)
}

func useContext(ctx context.Context) error {
	if ctx == nil {
		return fetch.ErrNilContext
	}

	return ctx.Err()
}
