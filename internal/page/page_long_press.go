package page

import "context"

// LongPress handles a press held in place at a point. A text control under
// the point selects the word there, the way a double click does, and claims
// the press. Otherwise a Handlers.LongPress runs on the innermost box under
// the point, found the way Click finds it, and the press is claimed. A nil
// handler reports false, so the window keeps its ordinary gesture.
func (p *Page) LongPress(ctx context.Context, x, y float64) (bool, error) {
	if err := useContext(ctx); err != nil {
		return false, err
	}

	id := p.boxIDAt(x, y)
	if c, ok := p.control(id); ok && !c.Disabled && canEdit(c) {
		return true, p.SelectWordAt(ctx, x, y)
	}

	box, ok := p.hitBox(x, y)
	if !ok || p.handlers.LongPress == nil {
		return false, nil
	}

	if box.ID == "" {
		if id := p.boxIDAt(x, y); id != "" {
			box = p.boxByID(id)
		}
	}

	if err := p.handlers.LongPress(ctx, box); err != nil {
		return true, err
	}

	return true, p.Redraw(ctx)
}
