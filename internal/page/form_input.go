package page

import "context"

func (p *Page) clickControl(ctx context.Context, box Box) error {
	c, _ := p.control(box.ID)
	if c.Tag == "button" && c.Disabled {
		return nil
	}
	if c.Type == "file" && !c.Disabled {
		return p.clickFile(ctx, box, c)
	}

	if !c.Disabled && willChange(c) {
		if err := p.beforeEdit(ctx, box.ID); err != nil {
			return err
		}
	}

	oldFocus := p.currentFocus()

	changed := false
	if !c.Disabled {
		changed = p.activate(box.ID)
	}

	if p.currentFocus() != oldFocus {
		p.markPair(oldFocus, p.currentFocus())
	}

	if changed {
		if c.Type == "checkbox" || c.Type == "radio" {
			p.markToggle(box.ID)
		}

		if next, ok := p.control(box.ID); ok {
			bindWrite(p, next)
		}

		if err := p.change(ctx, box.ID); err != nil {
			return err
		}
	}

	if p.handlers.Click != nil {
		if err := p.handlers.Click(ctx, box); err != nil {
			return err
		}
	}

	return p.Redraw(ctx)
}

func (p *Page) blurForm() {
	if p.form == nil {
		return
	}

	old := p.form.focusID
	if old != "" {
		p.form.lastFocusID = old
	}
	p.form.focusID = ""
	p.clearRange()
	p.markPair(old, "")
}

// currentFocus is the focused control id, or "" when none is focused.
func (p *Page) currentFocus() string {
	if p.form == nil {
		return ""
	}

	return p.form.focusID
}

func (p *Page) formControl(id string) bool {
	if id == "" {
		return false
	}

	_, ok := p.control(id)

	return ok
}

func canEdit(c Control) bool {
	if c.Tag == "select" || c.Type == "checkbox" || c.Type == "radio" {
		return false
	}

	return c.Tag == "textarea" || textLike(c.Type)
}
