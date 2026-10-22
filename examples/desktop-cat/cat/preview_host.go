package cat

import (
	"context"
	"image"

	"github.com/chinmay-sawant/ownframe"
)

func (p *preview) Click(ctx context.Context, x, y float64) error {
	return p.Page.Click(ctx, x, y)
}

func (p *preview) Hover(ctx context.Context, x, y float64) error {
	return p.Page.Hover(ctx, x, y)
}

func (p *preview) Press(ctx context.Context, x, y float64) error {
	return p.Page.Press(ctx, x, y)
}

func (p *preview) Release(ctx context.Context) error {
	return p.Page.Release(ctx)
}

func (p *preview) Type(ctx context.Context, text string) error {
	return p.Page.Type(ctx, text)
}

func (p *preview) KeyDown(ctx context.Context, key string) error {
	return p.Page.KeyDown(ctx, key)
}

func (p *preview) KeyUp(ctx context.Context, key string) error {
	return p.Page.KeyUp(ctx, key)
}

func (p *preview) Backspace(ctx context.Context) error {
	return p.Page.Backspace(ctx)
}

func (p *preview) DeleteWord(ctx context.Context) error {
	return p.Page.DeleteWord(ctx)
}

func (p *preview) Submit(ctx context.Context) error {
	return p.Page.Submit(ctx)
}

func (p *preview) Title() string { return p.Page.Title() }

func (p *preview) Boxes() []ownframe.Box { return p.Page.Boxes() }

func (p *preview) Copy(ctx context.Context) (string, bool, error) {
	return p.Page.Copy(ctx)
}

func (p *preview) Cut(ctx context.Context) (string, bool, error) {
	return p.Page.Cut(ctx)
}

func (p *preview) Paste(ctx context.Context, text string) error {
	return p.Page.Paste(ctx, text)
}

func (p *preview) SelectAll(ctx context.Context) error {
	return p.Page.SelectAll(ctx)
}

func (p *preview) Undo(ctx context.Context) error { return p.Page.Undo(ctx) }

func (p *preview) Redo(ctx context.Context) error { return p.Page.Redo(ctx) }

func (p *preview) Image() image.Image { return p.Page.Image() }

func (p *preview) Display() *ownframe.Display { return p.Page.Display() }

func (p *preview) PNG() []byte { return p.Page.PNG() }

func (p *preview) Generation() uint64 { return p.Page.Generation() }
