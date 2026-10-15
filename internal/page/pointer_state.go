package page

import "context"

// Hover sets the hovered element from a point and draws when it changed.
// The innermost box with an id under the point wins; an empty point clears it.
func (p *Page) Hover(ctx context.Context, x, y float64) error {
	if err := useContext(ctx); err != nil {
		return err
	}

	p.hoverX, p.hoverY = x, y

	return p.setHover(ctx, p.boxIDAt(x, y))
}

// Press sets the pressed element from a point and draws when it changed.
func (p *Page) Press(ctx context.Context, x, y float64) error {
	if err := useContext(ctx); err != nil {
		return err
	}

	id := p.boxIDAt(x, y)
	if p.active == id {
		return nil
	}

	p.markPair(p.active, id)
	p.active = id
	p.pressUsed = false
	if p.handlers.Press != nil {
		box := p.boxByID(id)
		if c, ok := p.control(id); !ok || !c.Disabled {
			used, err := p.handlers.Press(ctx, box, x, y)
			if err != nil {
				return err
			}
			p.pressUsed = used
		}
	}

	return p.Redraw(ctx)
}

// Release clears the pressed element and draws when it changed.
func (p *Page) Release(ctx context.Context) error {
	if err := useContext(ctx); err != nil {
		return err
	}

	if p.active == "" {
		return nil
	}

	p.markPair(p.active, "")
	p.active = ""
	p.pressUsed = false

	return p.Redraw(ctx)
}

func (p *Page) setHover(ctx context.Context, id string) error {
	if p.hover == id {
		return nil
	}

	handled, err := p.paintHover(ctx, id)
	if err != nil {
		return err
	}
	p.markPair(p.hover, id)
	p.hover = id
	if handled {
		return nil
	}

	return p.Redraw(ctx)
}

// boxIDAt returns the innermost box with an id that contains the point, or
// "". It delegates to the shared boxAt scan so Hover and Press agree with
// Click on overlapping boxes.
func (p *Page) boxIDAt(x, y float64) string {
	return p.boxAt(x, y).id
}
