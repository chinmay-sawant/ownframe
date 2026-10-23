package render

import (
	"context"

	"github.com/chinmay-sawant/blinkless/css"
	"github.com/chinmay-sawant/blinkless/layout"
)

// RelayoutElement lays one retained element out again under engineMu.
// The error is the one layout.RelayoutElement returned, with no extra wrap.
func RelayoutElement(
	ctx context.Context,
	doc *css.Document,
	prev *layout.Display,
	elementID string,
	options layout.Options,
) (*layout.ElementPatch, error) {
	engineMu.Lock()
	defer engineMu.Unlock()

	return layout.RelayoutElement(ctx, doc, prev, elementID, options)
}
