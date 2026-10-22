package page

import "context"

// PickFunc opens a file dialog and returns the path the person chose.
type PickFunc func(ctx context.Context, title string) (path string, err error)

// SetPicker installs the dialog a file control opens.
func (p *Page) SetPicker(pick PickFunc) {
	if p == nil {
		return
	}
	p.picker = pick
}

// InstallPicker sets the dialog a click on a file control opens.
// Run installs the picker for the current system; tests install a fake.
func InstallPicker(p interface{ SetPicker(PickFunc) }, pick PickFunc) {
	if p != nil {
		p.SetPicker(pick)
	}
}

// clickFile focuses a file control and stores the path the picker returns.
// ErrCanceled, ErrNoDialog, other picker errors, and an empty path only
// move the focus.
func (p *Page) clickFile(ctx context.Context, box Box, c Control) error {
	oldFocus := p.currentFocus()
	p.activate(box.ID)

	if p.currentFocus() != oldFocus {
		p.markPair(oldFocus, p.currentFocus())
	}

	if p.picker != nil {
		path, err := p.picker(ctx, p.title)
		if err == nil && path != "" && path != c.Value {
			if err := p.beforeEdit(ctx, box.ID); err != nil {
				return err
			}

			c.Value = path
			p.formState().byID[box.ID] = c
			bindWrite(p, c)
			p.markPending(box.ID)

			if err := p.change(ctx, box.ID); err != nil {
				return err
			}
		}
	}

	if p.handlers.Click != nil {
		if err := p.handlers.Click(ctx, box); err != nil {
			return err
		}
	}

	return p.Redraw(ctx)
}
