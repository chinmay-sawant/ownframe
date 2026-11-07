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

// RelayoutElements lays several retained elements out in one engine pass and
// builds the patched display once.
func RelayoutElements(
	ctx context.Context,
	doc *css.Document,
	prev *layout.Display,
	elementIDs []string,
	options layout.Options,
) (*layout.Display, error) {
	engineMu.Lock()
	defer engineMu.Unlock()

	return layout.RelayoutElements(ctx, doc, prev, elementIDs, options)
}

// RelayoutElementsInPlace is RelayoutElements without the whole-display
// clone: the batch commits directly into prev. PatchElements uses it because
// its failure path (failBatch) rolls the tree back and falls back to a full
// reload, which discards the display wholesale — a partially patched prev on
// error is therefore unobservable, while the clone it skips costs megabytes
// on a big page.
func RelayoutElementsInPlace(
	ctx context.Context,
	doc *css.Document,
	prev *layout.Display,
	elementIDs []string,
	options layout.Options,
) (*layout.Display, error) {
	engineMu.Lock()
	defer engineMu.Unlock()

	return layout.RelayoutElementsInPlace(ctx, doc, prev, elementIDs, options)
}
