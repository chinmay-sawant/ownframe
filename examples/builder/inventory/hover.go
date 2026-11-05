package main

import (
	"context"
	"errors"

	"github.com/chinmay-sawant/ownframe"
)

// handleHover mirrors the loglens hover precedent: the canvas renderer never
// paints HTML title tooltips, so the hovered bin cell title is shown in the
// bin-detail line via a direct patch. A refused patch falls back to a full
// refresh. A direct patch paints by itself, so this always returns false.
func (a *app) handleHover(ctx context.Context, previous, next ownframe.Box) (bool, error) {
	_ = previous
	want := ""
	id := ""
	if title, ok := binTitleByID(next.ID); ok {
		want = title
		id = next.ID
	}
	if id == a.hoverBin {
		return false, nil
	}
	a.hoverBin = id
	text := want
	if text == "" {
		text = binSummary()
	}
	if err := a.page.PatchElement(ctx, "bin-detail", ownframe.ElementEdit{Text: &text}); err != nil {
		if errors.Is(err, ownframe.ErrPatchRefused) {
			return false, a.refresh(ctx)
		}
		return false, err
	}
	return false, nil
}
