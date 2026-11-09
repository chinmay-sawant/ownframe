package render

import (
	"context"
	"fmt"

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
) (patch *layout.ElementPatch, err error) {
	engineMu.Lock()
	defer engineMu.Unlock()
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("render: relayout element %s panic: %v", elementID, r)
		}
	}()

	return layout.RelayoutElement(ctx, doc, prev, elementID, options)
}

// RelayoutElements lays several retained elements out in one engine pass and
// builds the patched display once.
func RelayoutElements(
	ctx context.Context,
	doc *css.Document,
	prev *layout.Display,
	elementIDs []string,
	options layout.Options,
) (disp *layout.Display, err error) {
	engineMu.Lock()
	defer engineMu.Unlock()
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("render: relayout elements panic: %v", r)
		}
	}()

	return layout.RelayoutElements(ctx, doc, prev, elementIDs, options)
}
