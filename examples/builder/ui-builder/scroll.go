package main

import (
	"context"
	"fmt"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/components/chadcn"
	"github.com/chinmay-sawant/ownframe/ui"
)

const (
	builderScrollViewportID  = "builder-scroll-viewport"
	builderScrollContentID   = "builder-scroll-content"
	builderScrollTrackID     = "builder-scroll-track"
	builderScrollThumbID     = "builder-scroll-thumb"
	builderActivePanelPrefix = "builder-active-panel-"
	builderScrollStep        = 48.0
	builderHeaderHeight      = 68
	builderFooterHeight      = 34
	builderPanePadding       = 16
	builderSidebarWidth      = 205
	builderContentGap        = 16
	builderScrollGutter      = 14
)

func (a *builderApp) scrollViewport(content *ui.Element) *ui.Element {
	inner := ui.Div().ID(builderScrollContentID).Position(ui.PositionRelative).Top(ui.Px(-float64(a.center.Offset()))).
		Padding(ui.Edges(ui.Px(0), ui.Px(builderScrollGutter), ui.Px(0), ui.Px(0))).BoxSizing(ui.BoxSizingBorderBox).Children(content)
	bar := chadcn.Scrollbar(chadcn.ScrollbarOptions{
		ID: builderScrollTrackID, ThumbID: builderScrollThumbID, ControlsID: builderScrollViewportID, Label: "Component catalog",
		State: &a.center, TrackPixels: a.center.Metrics().Viewport,
		OnChange: func(ctx context.Context) error {
			a.page.Form().RequestFocus(builderScrollThumbID)
			return a.refresh(ctx)
		},
	})
	return ui.Div().ID(builderScrollViewportID).Flex(ui.Number(1)).MinWidth(ui.Px(0)).MinHeight(ui.Px(0)).
		Position(ui.PositionRelative).Overflow(ui.OverflowHidden).Children(inner, bar)
}

func (a *builderApp) scrollInputBlocked() bool {
	return a.tab == "dialog" && a.dialog.open || a.tab == "compare" && a.compare.Preview
}

func (a *builderApp) handleWheel(ctx context.Context, x, y, dx, dy float64) (bool, error) {
	if dx == 0 && dy == 0 {
		return false, nil
	}
	if a.scrollInputBlocked() {
		return true, nil
	}
	if !a.overCenterViewport(x, y) {
		return false, nil
	}
	if a.center.Wheel(dy, builderScrollStep) {
		return true, a.refresh(ctx)
	}
	return true, nil
}

func (a *builderApp) overCenterViewport(x, y float64) bool {
	viewport := a.centerViewport()
	return x >= viewport.X && x <= viewport.X+viewport.W && y >= viewport.Y && y <= viewport.Y+viewport.H
}

func (a *builderApp) startCenterScrollDrag(_ context.Context, box ownframe.Box) (bool, error) {
	if a.scrollInputBlocked() || box.ID != builderScrollThumbID {
		return false, nil
	}
	return a.center.BeginDrag(a.center.Metrics().Viewport), nil
}

func (a *builderApp) moveCenterScrollDrag(ctx context.Context, _, dy float64) error {
	if a.scrollInputBlocked() {
		a.center.EndDrag()
		return nil
	}
	if a.center.DragBy(dy, a.center.Metrics().Viewport) {
		return a.refresh(ctx)
	}
	return nil
}

func (a *builderApp) endCenterScrollDrag(context.Context) error { a.center.EndDrag(); return nil }

func (a *builderApp) updateCenterScrollMetrics() {
	viewport := a.centerViewport()
	content := int(viewport.H)
	for _, box := range a.page.Boxes() {
		if box.ID == builderScrollContentID {
			content = max(content, int(box.H))
			break
		}
	}
	a.center.SetRange(content, int(viewport.H))
}

func (a *builderApp) centerViewport() ownframe.Box {
	for _, box := range a.page.Boxes() {
		if box.ID == builderScrollViewportID {
			return box
		}
	}
	width, height := a.page.Size()
	left := builderPanePadding*3 + builderSidebarWidth + builderContentGap
	return ownframe.Box{X: float64(left), Y: float64(builderHeaderHeight + builderPanePadding),
		W: float64(max(width-left-builderPanePadding, 0)), H: float64(max(height-builderHeaderHeight-builderFooterHeight-2*builderPanePadding, 0))}
}

func (a *builderApp) refreshOnResize(ctx context.Context) error {
	width, height := a.page.Size()
	resized := width != a.layoutWidth || height != a.layoutHeight
	if !resized && !a.centerMeasurePending {
		return nil
	}
	a.layoutWidth, a.layoutHeight = width, height
	before := a.center.Metrics()
	if resized {
		if err := a.view.Update(ctx, a.screen()); err != nil {
			return fmt.Errorf("update builder after resize: %w", err)
		}
	}
	a.updateCenterScrollMetrics()
	a.centerMeasurePending = false
	if before != a.center.Metrics() {
		if err := a.view.Update(ctx, a.screen()); err != nil {
			return fmt.Errorf("update builder scrollbar: %w", err)
		}
	}
	return nil
}
