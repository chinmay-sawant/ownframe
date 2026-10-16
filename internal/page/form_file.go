package page

import "context"

// PickFunc opens a file dialog and returns the path the person chose.
// ok is false when the person canceled or no dialog is available.
type PickFunc func(ctx context.Context, title string) (path string, ok bool)

// InstallPicker sets the dialog a click on a file control opens.
// Run installs the picker for the current system; tests install a fake.
func InstallPicker(p *Page, pick PickFunc) {
	p.picker = pick
}

// clickFile focuses a file control and stores the path the picker returns.
// A canceled dialog and an empty path only move the focus.
func (p *Page) clickFile(ctx context.Context, box Box, c Control) error {
	oldFocus := p.currentFocus()
	p.activate(box.ID)

	if p.currentFocus() != oldFocus {
		p.markPair(oldFocus, p.currentFocus())
	}

	if p.picker != nil {
		if path, ok := p.picker(ctx, p.title); ok && path != "" && path != c.Value {
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
