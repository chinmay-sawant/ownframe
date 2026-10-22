package pagetest

import (
	"context"
	"image"

	"github.com/chinmay-sawant/ownframe"
)

// Pump runs one host-loop tick on p.
func Pump(ctx context.Context, p *ownframe.Page) error {
	return p.Testing().Tick(ctx)
}

// TakeDirty drains the host dirty bounding box.
func TakeDirty(p *ownframe.Page) (image.Rectangle, bool) {
	return p.Testing().TakeDirty()
}
