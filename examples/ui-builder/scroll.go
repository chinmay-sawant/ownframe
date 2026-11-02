package main

import (
	"context"
	"fmt"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/ui"
)

const (
	builderScrollTrackID     = "builder-scroll-track"
	builderScrollThumbID     = "builder-scroll-thumb"
	builderActivePanelPrefix = "builder-active-panel-"
	builderScrollStep        = 48.0
	builderScrollThumbMin    = 24
	builderHeaderHeight      = 68
	builderFooterHeight      = 34
	builderPanePadding       = 16
	builderSidebarWidth      = 205
	builderContentGap        = 16
	builderScrollGutter      = 14
)

// scrollViewport gives the active panel its own clipped scroll area. The
// surrounding titlebar, sidebar, and notice footer remain in normal layout.
func (a *builderApp) scrollViewport(content *ui.Element) *ui.Element {
	inner := ui.Div().Position(ui.PositionRelative).Top(ui.Px(-float64(a.centerScroll))).
		Padding(ui.Edges(ui.Px(0), ui.Px(builderScrollGutter), ui.Px(0), ui.Px(0))).
		BoxSizing(ui.BoxSizingBorderBox).Children(content)
	return ui.Div().Flex(ui.RawValue("1")).
		MinWidth(ui.Px(0)).MinHeight(ui.Px(0)).Position(ui.PositionRelative).
		Overflow(ui.OverflowHidden).
		Children(inner, a.centerScrollTrack())
}

func (a *builderApp) centerScrollTrack() *ui.Element {
	return ui.Div().ID(builderScrollTrackID).Position(ui.PositionAbsolute).
		Top(ui.Px(0)).Right(ui.Px(3)).Width(ui.Px(7)).Height(ui.Percent(100)).
		Children(a.centerScrollThumb())
}

func (a *builderApp) centerScrollThumb() *ui.Element {
	top, height, visible := a.centerThumb()
	if !visible {
		return nil
	}
	return ui.Div().ID(builderScrollThumbID).Position(ui.PositionAbsolute).
		Top(ui.Px(float64(top))).Left(ui.Px(0)).Width(ui.Percent(100)).
		Height(ui.Px(float64(height))).ZIndex(ui.Number(2)).
		BorderRadius(ui.Px(4)).BackgroundColor(ui.RawValue("var(--of-scrollbar,#8394ac)"))
}

func (a *builderApp) handleWheel(ctx context.Context, x, y, dx, dy float64) (bool, error) {
	if dx == 0 && dy == 0 || !a.overCenterViewport(x, y) {
		return false, nil
	}
	a.centerWheelRemainder += -dy * builderScrollStep
	step := int(a.centerWheelRemainder)
	a.centerWheelRemainder -= float64(step)
	if step == 0 {
		return true, nil
	}
	before := a.centerScroll
	a.centerScroll = clampCenterScroll(a.centerScroll+step, a.centerContentH, a.centerViewH)
	if before == a.centerScroll {
		return true, nil
	}
	return true, a.refresh(ctx)
}

func (a *builderApp) overCenterViewport(x, y float64) bool {
	viewport := a.centerViewport()
	return x >= viewport.X && x <= viewport.X+viewport.W &&
		y >= viewport.Y && y <= viewport.Y+viewport.H
}

func (a *builderApp) startCenterScrollDrag(_ context.Context, box ownframe.Box) (bool, error) {
	if box.ID != builderScrollThumbID {
		return false, nil
	}
	a.centerDragging = true
	a.centerDragRemainder = 0
	return true, nil
}

func (a *builderApp) moveCenterScrollDrag(ctx context.Context, _, dy float64) error {
	if !a.centerDragging {
		return nil
	}
	travel := a.centerViewH - a.centerThumbH
	maxOffset := max(a.centerContentH-a.centerViewH, 0)
	if travel <= 0 || maxOffset == 0 {
		return nil
	}
	a.centerDragRemainder += dy * float64(maxOffset) / float64(travel)
	step := int(a.centerDragRemainder)
	a.centerDragRemainder -= float64(step)
	if step == 0 {
		return nil
	}
	before := a.centerScroll
	a.centerScroll = clampCenterScroll(a.centerScroll+step, a.centerContentH, a.centerViewH)
	if before == a.centerScroll {
		return nil
	}
	return a.refresh(ctx)
}

func (a *builderApp) endCenterScrollDrag(context.Context) error {
	a.centerDragging = false
	a.centerDragRemainder = 0
	return nil
}

func (a *builderApp) updateCenterScrollMetrics() {
	viewport := a.centerViewport()
	viewH, contentH := int(viewport.H), int(viewport.H)
	if viewH > 0 {
		bottom := viewport.Y + viewport.H
		boxes := a.page.Boxes()
		panelID := builderActivePanelPrefix + a.tab
		panelH := 0.0
		for i, box := range boxes {
			if box.ID != panelID {
				continue
			}
			parent := box.ParentIndex
			for parent >= 0 && parent < len(boxes) {
				panel := boxes[parent]
				if panel.X == viewport.X && panel.Y == viewport.Y {
					panelH = max(panelH, panel.H)
				}
				next := panel.ParentIndex
				if next >= parent || next == i {
					break
				}
				parent = next
			}
			bottom = max(bottom, viewport.Y+panelH+float64(a.centerScroll))
			break
		}
		contentH = int(bottom - viewport.Y)
	}
	a.centerViewH, a.centerContentH = viewH, contentH
	a.centerScroll = clampCenterScroll(a.centerScroll, a.centerContentH, a.centerViewH)
	_, a.centerThumbH, _ = a.centerThumb()
}

func (a *builderApp) centerViewport() ownframe.Box {
	width, height := a.page.Size()
	left := builderPanePadding*3 + builderSidebarWidth + builderContentGap
	viewport := ownframe.Box{
		X: float64(left), Y: float64(builderHeaderHeight + builderPanePadding),
		W: float64(max(width-left-builderPanePadding, 0)),
		H: float64(max(height-builderHeaderHeight-builderFooterHeight-2*builderPanePadding, 0)),
	}
	for _, box := range a.page.Boxes() {
		if box.ID != builderScrollTrackID {
			continue
		}
		viewport.Y, viewport.H = box.Y, box.H
		break
	}
	return viewport
}

func (a *builderApp) refreshOnResize(ctx context.Context) error {
	width, height := a.page.Size()
	resized := width != a.layoutWidth || height != a.layoutHeight
	if !resized && !a.centerMeasurePending {
		return nil
	}
	a.layoutWidth, a.layoutHeight = width, height
	beforeTop, beforeHeight, beforeVisible := a.centerThumb()
	if resized {
		if err := a.view.Update(ctx, a.screen()); err != nil {
			return fmt.Errorf("update builder after resize: %w", err)
		}
	}
	a.updateCenterScrollMetrics()
	a.centerMeasurePending = false
	afterTop, afterHeight, afterVisible := a.centerThumb()
	if beforeTop == afterTop && beforeHeight == afterHeight && beforeVisible == afterVisible {
		return nil
	}
	if err := a.view.Update(ctx, a.screen()); err != nil {
		return fmt.Errorf("update builder scrollbar: %w", err)
	}
	return nil
}

func (a *builderApp) centerThumb() (top, height int, visible bool) {
	track := a.centerViewH
	content := a.centerContentH
	if track < 1 || content <= track {
		return 0, 0, false
	}
	height = min(max(track*track/content, builderScrollThumbMin), track)
	travel := track - height
	maxOffset := content - track
	top = travel * min(max(a.centerScroll, 0), maxOffset) / maxOffset
	return top, height, true
}

func clampCenterScroll(offset, content, view int) int {
	return min(max(offset, 0), max(content-view, 0))
}
