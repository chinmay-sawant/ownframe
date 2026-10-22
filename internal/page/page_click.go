package page

import (
	"context"

	"github.com/chinmay-sawant/ownframe/internal/host"
)

// Click hit-tests the last picture.
// The innermost box is the last one in document order that contains the point.
// A box without an id falls back to the innermost id-bearing element under the
// point, so a click on a child icon reaches its control.
// A field is activated. An enabled button can follow a route.
// A disabled control is not toggled and does not blur.
// Any other hit blurs the form. A routed data-action then loads that HTML
// and skips the handler. Click draws the page again.
func (p *Page) Click(ctx context.Context, x, y float64) error {
	if err := useContext(ctx); err != nil {
		return err
	}

	box, ok := p.hitBox(x, y)
	if ok && box.ID == "" {
		if id := p.boxIDAt(x, y); id != "" {
			box = p.boxByID(id)
		}
	}
	if ok && p.consumePress(box.ID) {
		return p.Redraw(ctx)
	}

	if ok && p.formControl(box.ID) {
		c, _ := p.control(box.ID)
		if c.Tag == "button" && !c.Disabled && box.Action != "" {
			if html, ok := p.nav.Lookup(box.Action); ok {
				return p.Load(ctx, html)
			}
		}
		return p.clickControl(ctx, box)
	}

	if ok {
		p.blurForm()
		if box.Action != "" {
			if html, ok := p.nav.Lookup(box.Action); ok {
				return p.Load(ctx, html)
			}
		}
	}

	if p.handlers.OpenLink != nil && (!ok || box.Action == "") {
		if handled, err := p.openLinkAt(ctx, x, y); handled {
			return err
		}
	}

	if ok && p.handlers.Click != nil {
		if err := p.handlers.Click(ctx, box); err != nil {
			return err
		}
	}

	if ok {
		painted, err := p.clickPainted(ctx, box)
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
	return p.RedrawIfChanged(ctx)
}

var _ host.Screen = (*Page)(nil)
