package main

import (
	"context"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/ui"
	"github.com/chinmay-sawant/ownframe/ui/components"
)

func (a *app) paneScrollbar(id, thumb, viewport, label string, state *components.ScrollState, track int) *ui.Element {
	return components.Scrollbar(components.ScrollbarOptions{
		ID: id, ThumbID: thumb, ControlsID: viewport, Label: label, State: state, TrackPixels: track,
		OnChange: func(ctx context.Context) error { a.page.Form().RequestFocus(thumb); return a.refresh(ctx) },
	})
}

func (a *app) startScrollDrag(_ context.Context, box ownframe.Box) (bool, error) {
	if a.settings {
		return false, nil
	}
	switch box.ID {
	case "mail-list-thumb":
		return a.listScroll.BeginDrag(a.listViewH), nil
	case "mail-reader-thumb":
		return a.readerScroll.BeginDrag(a.readerScroll.Metrics().Viewport), nil
	default:
		return false, nil
	}
}

func (a *app) moveScrollDrag(ctx context.Context, _, dy float64) error {
	if a.settings {
		a.listScroll.EndDrag()
		a.readerScroll.EndDrag()
		return nil
	}
	list := a.listScroll.DragBy(dy, a.listViewH)
	reader := a.readerScroll.DragBy(dy, a.readerScroll.Metrics().Viewport)
	if list || reader {
		return a.refresh(ctx)
	}
	return nil
}

func (a *app) endScrollDrag(context.Context) error {
	a.listScroll.EndDrag()
	a.readerScroll.EndDrag()
	return nil
}
