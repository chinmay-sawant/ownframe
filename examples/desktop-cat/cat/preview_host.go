package cat

import (
	"context"
	"image"

	"github.com/chinmay-sawant/ownframe"
)

func (p *preview) Click(ctx context.Context, x, y float64) error {
	return p.Testing().Click(ctx, x, y)
}

func (p *preview) Hover(ctx context.Context, x, y float64) error {
	return p.Testing().Hover(ctx, x, y)
}

func (p *preview) Press(ctx context.Context, x, y float64) error {
	return p.Testing().Press(ctx, x, y)
}

func (p *preview) Release(ctx context.Context) error {
	return p.Testing().Release(ctx)
}

func (p *preview) Type(ctx context.Context, text string) error {
	return p.Testing().Type(ctx, text)
}

func (p *preview) KeyDown(ctx context.Context, key string) error {
	return p.Testing().KeyDown(ctx, key)
}

func (p *preview) KeyUp(ctx context.Context, key string) error {
	return p.Testing().KeyUp(ctx, key)
}

func (p *preview) Backspace(ctx context.Context) error {
	return p.Testing().Backspace(ctx)
}

func (p *preview) DeleteWord(ctx context.Context) error {
	return p.Testing().DeleteWord(ctx)
}

func (p *preview) Submit(ctx context.Context) error {
	return p.Testing().Submit(ctx)
}

func (p *preview) Image() image.Image { return p.Testing().Image() }

func (p *preview) Display() *ownframe.Display { return p.Testing().Display() }

func (p *preview) PNG() []byte { return p.Testing().PNG() }

func (p *preview) Generation() uint64 { return p.Testing().Generation() }
