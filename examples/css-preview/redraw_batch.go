package main

import (
	"context"

	"github.com/chinmay-sawant/ownframe"
)

// redrawBatch defers page draws across fn, then draws once.
func redrawBatch(ctx context.Context, p *ownframe.Page, fn func() error) (err error) {
	p.BeginRedrawBatch()
	defer func() {
		if endErr := p.EndRedrawBatch(ctx); err == nil {
			err = endErr
		}
	}()
	return fn()
}
