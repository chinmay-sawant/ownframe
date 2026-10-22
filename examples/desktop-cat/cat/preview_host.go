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
	return p.Page.Edit().Hover(ctx, x, y)
}

func (p *preview) Press(ctx context.Context, x, y float64) error {
	return p.Page.Edit().Press(ctx, x, y)
}

func (p *preview) Release(ctx context.Context) error {
	return p.Page.Edit().Release(ctx)
}

func (p *preview) Type(ctx context.Context, text string) error {
	return p.Page.Edit().Type(ctx, text)
}

func (p *preview) KeyDown(ctx context.Context, key string) error {
	return p.Page.Edit().KeyDown(ctx, key)
}

func (p *preview) KeyUp(ctx context.Context, key string) error {
	return p.Page.Edit().KeyUp(ctx, key)
}

func (p *preview) Backspace(ctx context.Context) error {
	return p.Page.Edit().Backspace(ctx)
}

func (p *preview) DeleteWord(ctx context.Context) error {
	return p.Page.Edit().DeleteWord(ctx)
}

func (p *preview) Submit(ctx context.Context) error {
	return p.Page.Edit().Submit(ctx)
}

func (p *preview) Title() string { return p.Page.Title() }

func (p *preview) Boxes() []ownframe.Box { return p.Page.Boxes() }

func (p *preview) Copy(ctx context.Context) (string, bool, error) {
	return p.Page.Edit().Copy(ctx)
}

func (p *preview) Cut(ctx context.Context) (string, bool, error) {
	return p.Page.Edit().Cut(ctx)
}

func (p *preview) Paste(ctx context.Context, text string) error {
	return p.Page.Edit().Paste(ctx, text)
}

func (p *preview) SelectAll(ctx context.Context) error {
	return p.Page.Edit().SelectAll(ctx)
}

func (p *preview) Undo(ctx context.Context) error { return p.Page.Edit().Undo(ctx) }

func (p *preview) Redo(ctx context.Context) error { return p.Page.Edit().Redo(ctx) }

func (p *preview) Image() image.Image { return p.Page.Probe().Image() }

func (p *preview) Display() *ownframe.Display { return p.Page.Probe().Display() }

func (p *preview) PNG() []byte { return p.Page.Probe().PNG() }

func (p *preview) Generation() uint64 { return p.Page.Probe().Generation() }
