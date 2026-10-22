package page

import (
	"context"
	"errors"
	"fmt"

	"github.com/chinmay-sawant/ownframe/internal/filepick"
)

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
// Cancel or an empty path keeps the old value, still runs Click, and
// redraws. Any other picker error redraws the focus and returns wrapped.
func (p *Page) clickFile(ctx context.Context, box Box, c Control) error {
	oldFocus := p.currentFocus()
	p.activate(box.ID)
	if p.currentFocus() != oldFocus {
		p.markPair(oldFocus, p.currentFocus())
	}
	if p.picker != nil {
		if err := p.applyPicked(ctx, box, c); err != nil {
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

func (p *Page) applyPicked(ctx context.Context, box Box, c Control) error {
	path, err := p.picker(ctx, p.title)
	if err != nil && !errors.Is(err, filepick.ErrCanceled) {
		if rerr := p.Redraw(ctx); rerr != nil {
			err = errors.Join(err, rerr)
		}

		return fmt.Errorf("ownframe: %w", err)
	}
	if err != nil || path == "" || path == c.Value {
		return nil
	}
	if err := p.beforeEdit(ctx, box.ID); err != nil {
		return err
	}
	c.Value = path
	p.formState().byID[box.ID] = c
	bindWrite(p, c)
	p.markPending(box.ID)

	return p.change(ctx, box.ID)
}
