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
			a.scrollDirty = true
			return nil
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
	// Wheel input bursts arrive faster than a rebuild. Record the offset and
	// let the frame tick paint it once, so one frame never pays two rebuilds.
	if a.center.Wheel(dy, builderScrollStep) {
		a.scrollDirty = true
		return true, nil
	}
	return true, nil
}

// scrollCenterKey advances the center viewport for a keyboard scroll key and
// defers the repaint to the frame tick, mirroring handleWheel. It reports
// whether it consumed the key; a false return must fall through to the other
// key branches in handleKey.
func (a *builderApp) scrollCenterKey(key string) bool {
	if a.scrollInputBlocked() || a.page.Form().FocusID() != builderScrollThumbID {
		return false
	}
	if a.center.Key(key, int(builderScrollStep)) {
		a.scrollDirty = true
		return true
	}
	return false
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
		a.scrollDirty = true
		return nil
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

// Windowing note: once catalogPanel renders catalogWindowedRows, windowed
// frames keep the top/bottom spacer stubs inside the content box, so the box
// height above still measures the full catalog and this range stays put.
// The spacers are sized for exact preservation up to one container gap of
// layout tolerance (see catalogWindowSpacers); the two overscan rows hide
// the rest, and the max above absorbs any shortfall.

// ensureCatalogWindowing installs the catalog row-window callback once. Page
// panning stays off for this gallery (SetAllowScroll(false) in newBuilderApp),
// so the shell never drives a nonzero page offset; the live scroll driver is
// the center ScrollState, which catalogWindowPlan reads at build time. The
// install keeps parity with the code-view precedent and covers any future
// page-scroll use. It is idempotent: this app never enables page windowing
// for another purpose, so a set flag means our callback is installed.
func (a *builderApp) ensureCatalogWindowing() {
	if a.page == nil || a.page.Scroll().Windowing() {
		return
	}
	a.page.Scroll().SetWindowing(true)
	a.page.SetScrollWindow(a.catalogOnScroll)
}

// catalogOnScroll is the Page.SetScrollWindow callback. It never touches the
// view: it folds the page offset together with the authoritative center
// offset and reports whether that position implies a partial row window, so
// it is safe to run inside the redraw path (applyScrollWindow) and on the
// shell scroll step. See the TODO in catalog_window.go for the render hookup.
func (a *builderApp) catalogOnScroll(offsetY, viewH int) bool {
	if a.tab != "catalog" {
		return false
	}
	headerPx, pitchPx, _, ok := a.catalogWindowGeometry()
	if !ok {
		return false
	}
	if viewH <= 0 {
		viewH = int(a.centerViewport().H)
	}
	totalRows := catalogRowCount(len(catalogEntries()))
	start, end := catalogWindowRange(totalRows, offsetY+a.center.Offset(), viewH, headerPx, pitchPx, catalogWindowOverscanRows)
	return start > 0 || end < totalRows
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
	a.ensureCatalogWindowing()
	width, height := a.page.Size()
	resized := width != a.layoutWidth || height != a.layoutHeight
	if !resized && !a.centerMeasurePending && !a.scrollDirty {
		return nil
	}
	a.layoutWidth, a.layoutHeight = width, height
	before := a.center.Metrics()
	if resized {
		a.catalogGeoMeasured = false
		if err := a.view.Update(ctx, a.screen()); err != nil {
			return fmt.Errorf("update builder after resize: %w", err)
		}
	}
	if a.scrollDirty {
		if err := a.refresh(ctx); err != nil {
			return err
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
