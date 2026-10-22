package page

import "context"

// runTick runs fn then the caret blink, then records in-place op patches
// as dirty rects. A no-op leaves FrameDirty unset so idle park can sleep.
func (p *Page) runTick(ctx context.Context, fn func(context.Context) error) error {
	gen := p.Generation()
	if fn == nil {
		err := p.caretBlink(ctx)
		p.flagTickDirty(gen)

		return err
	}

	snap := p.captureTickOps()
	if err := fn(ctx); err != nil {
		p.freeOps(snap)

		return err
	}

	err := p.caretBlink(ctx)
	p.noteTickDamage(gen, snap)
	p.freeOps(snap)

	return err
}

func (p *Page) captureTickOps() []DisplayOp {
	if p.display == nil {
		return nil
	}

	return append(p.rentOps(len(p.display.Ops)), p.display.Ops...)
}

func (p *Page) noteTickDamage(gen uint64, snap []DisplayOp) {
	if p.Generation() == gen && p.display != nil && snap != nil {
		before := *p.display
		before.Ops = snap
		d, full, _ := diffDisplay(&before, p.display)
		if full {
			p.markFull()
		} else {
			for i := 0; i < d.n; i++ {
				p.markRect(d.rects[i])
			}
		}
	}

	p.flagTickDirty(gen)
}

func (p *Page) flagTickDirty(gen uint64) {
	if p.Generation() != gen || p.dirty.full || len(p.dirty.rects) > 0 {
		p.frameDirty.Store(true)
	}
}
