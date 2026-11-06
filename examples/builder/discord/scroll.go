package main

import (
	"context"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/components/chadcn"
	"github.com/chinmay-sawant/ownframe/ui"
)

// chatWheelStep scrolls the transcript in CSS pixels per wheel notch,
// matching the mail reader pace.
const chatWheelStep = 48.0

// paneScrollbar overlays a themed scrollbar for the chat viewport.
func (a *app) paneScrollbar() *ui.Element {
	return chadcn.Scrollbar(chadcn.ScrollbarOptions{
		ID:          "dc-scroll-bar",
		ThumbID:     "dc-scroll-thumb",
		ControlsID:  "dc-scroll-viewport",
		Label:       "Messages",
		State:       &a.chatScroll,
		TrackPixels: a.chatViewH,
		OnChange: func(ctx context.Context) error {
			a.page.Form().RequestFocus("dc-scroll-thumb")
			return a.refresh(ctx)
		},
	})
}

// handleWheel scrolls the chat transcript under the cursor. It consumes the
// event so the window does not pan the 100vh shell, which would move the
// channel header and composer instead of the messages.
func (a *app) handleWheel(ctx context.Context, x, y, dx, dy float64) (bool, error) {
	if dx == 0 && dy == 0 {
		return false, nil
	}
	if a.settingsOpen {
		return true, nil
	}
	if a.page == nil {
		return false, nil
	}
	sx, sy := a.page.Scroll().ScrollOffset()
	x -= float64(sx)
	y -= float64(sy)
	if !overBoxID(a.page.Boxes(), "dc-scroll-viewport", x, y) {
		return false, nil
	}
	// Measure first so the wheel acts on a fresh range; otherwise a stale
	// max lets one event render past the latest message.
	a.updateScrollCache(a.currentMessages())
	if a.chatScroll.Wheel(dy, chatWheelStep) {
		return true, a.refresh(ctx)
	}
	return true, nil
}

func overBoxID(boxes []ownframe.Box, id string, x, y float64) bool {
	for _, b := range boxes {
		if b.ID == id {
			return x >= b.X && x <= b.X+b.W && y >= b.Y && y <= b.Y+b.H
		}
	}
	return false
}

func (a *app) startScrollDrag(_ context.Context, box ownframe.Box) (bool, error) {
	if box.ID == "dc-scroll-thumb" {
		return a.chatScroll.BeginDrag(a.chatViewH), nil
	}
	return false, nil
}

func (a *app) moveScrollDrag(ctx context.Context, _, dy float64) error {
	a.updateScrollCache(a.currentMessages())
	if a.chatScroll.DragBy(dy, a.chatViewH) {
		return a.refresh(ctx)
	}
	return nil
}

func (a *app) endScrollDrag(context.Context) error {
	a.chatScroll.EndDrag()
	return nil
}

// scrollCacheKey tracks the measured viewport so refresh converges.
func (a *app) scrollCacheKey() [3]int {
	m := a.chatScroll.Metrics()
	return [3]int{a.chatViewH, m.Viewport, m.Content}
}

// updateScrollCache measures the viewport box and sets the scroll range
// from the estimated transcript total plus the content padding. It reports
// whether the stored offset was clamped, so refresh can repaint the
// corrected position instead of leaving a white tail on screen.
func (a *app) updateScrollCache(msgs []dmessage) bool {
	if a.page == nil {
		return false
	}
	for _, box := range a.page.Boxes() {
		if box.ID == "dc-scroll-viewport" {
			a.chatViewH = max(int(box.H), 0)
			break
		}
	}
	m := a.chatScroll.Metrics()
	viewport := a.chatViewH
	if viewport <= 0 {
		viewport = m.Viewport
	}
	changed := a.chatScroll.SetRange(scrollTotal(msgs)+scrollPadV, viewport)
	if a.chatScroll.Offset() > a.chatScroll.Metrics().MaxOffset() {
		a.chatScroll.ScrollTo(a.chatScroll.Metrics().MaxOffset())
		return true
	}
	return changed
}

// Transcript windowing core: host-owned offset, pure arithmetic, zero
// allocations in hot paths. The engine plays no role here: callers keep the
// offset in app.scrollPx, size rows with estimateMsgHeight, size the full
// transcript with scrollTotal, and slice the render list with visibleRange.
// No func below appends, maps, or otherwise heap-allocates; all walks are
// index math over msgs plus len() and flag reads.
const (
	scrollCharsPerLine = 90
	scrollBaseH        = 64
	scrollLineH        = 18
	scrollReplyH       = 20
	scrollReactH       = 30
	scrollDividerH     = 28
	scrollOverscan     = 300
	scrollHistCap      = 500
	scrollWheelStep    = 3
	// scrollPadV is the vertical content padding (top 16 + bottom 8 in
	// styles.go #dc-scroll-content). The estimate covers rows and
	// dividers only, so the range adds the padding explicitly.
	scrollPadV = 24
)

// estimateMsgHeight sizes one transcript row in CSS pixels without touching
// layout: one line holds scrollCharsPerLine bytes, the first line costs
// scrollBaseH, each extra line costs scrollLineH, a reply context line costs
// scrollReplyH, and a reaction row costs scrollReactH. It reads only len()
// and flags, so it allocates nothing.
func estimateMsgHeight(m dmessage) int {
	lines := (len(m.Text) + scrollCharsPerLine - 1) / scrollCharsPerLine
	if lines < 1 {
		lines = 1
	}
	h := scrollBaseH + scrollLineH*(lines-1)
	if m.ReplyTo != "" {
		h += scrollReplyH
	}
	if len(m.Reactions) > 0 {
		h += scrollReactH
	}
	return h
}

// scrollTotal sizes the whole transcript in CSS pixels: the sum of
// estimateMsgHeight over msgs plus scrollDividerH per day divider. Dividers
// follow the messages.go transcript loop: walking in order, a divider
// precedes message i when messageDay(msgs[i].Time) is non-empty and differs
// from the previous divider label (equivalently, the last non-empty day
// seen). Index math only; allocates nothing.
func scrollTotal(msgs []dmessage) int {
	total := 0
	prevDay := ""
	for i := range msgs {
		if day := messageDay(msgs[i].Time); day != "" && day != prevDay {
			total += scrollDividerH
			prevDay = day
		}
		total += estimateMsgHeight(msgs[i])
	}
	return total
}

// visibleRange slices the transcript for windowed rendering.
//
// Semantics: the transcript is a sequence of blocks, one per message, where
// block i is (divider? scrollDividerH : 0) + estimateMsgHeight(msgs[i]), with
// the divider rule matching scrollTotal. Block 0 starts at y=0; block i
// starts where block i-1 ends. total is scrollTotal(msgs).
//
// The offset is clamped to [0, max(0, total-viewH)] before use. The render
// window is [offset, offset+viewH) expanded by scrollOverscan above and below
// (clamped to [0, total]). lo is the first message index whose block end
// lies past the expanded top, hi is one past the last message index whose
// block start lies before the expanded bottom, so callers render
// msgs[lo:hi]. topPad is the stacked height of all blocks before lo
// (dividers plus clipped messages, in px) and botPad is total minus the
// stacked height through the end of block hi-1, so topPad + rendered +
// botPad preserves total height. The divider that would precede msgs[lo] in
// the full walk is folded into topPad and must not be re-emitted; for i in
// (lo, hi), emit a divider when messageDay(msgs[i].Time) is non-empty and
// differs from the running label seeded with the day state preceding lo.
// Empty input returns (0, 0, 0, 0); non-positive viewH returns (0, 0, 0,
// total). Index math only; allocates nothing.
func visibleRange(msgs []dmessage, offset, viewH int) (lo, hi, topPad, botPad int) {
	total := scrollTotal(msgs)
	if len(msgs) == 0 {
		return 0, 0, 0, 0
	}
	if viewH <= 0 {
		return 0, 0, 0, total
	}
	maxOffset := total - viewH
	if maxOffset < 0 {
		maxOffset = 0
	}
	if offset < 0 {
		offset = 0
	}
	if offset > maxOffset {
		offset = maxOffset
	}
	loBound := offset - scrollOverscan
	if loBound < 0 {
		loBound = 0
	}
	hiBound := offset + viewH + scrollOverscan
	if hiBound > total {
		hiBound = total
	}
	lo = len(msgs)
	hi = len(msgs)
	topPad = total
	endHi := 0
	y := 0
	prevDay := ""
	for i := range msgs {
		day := messageDay(msgs[i].Time)
		blockStart := y
		if day != "" && day != prevDay {
			y += scrollDividerH
			prevDay = day
		}
		y += estimateMsgHeight(msgs[i])
		blockEnd := y
		if blockEnd > loBound && lo == len(msgs) {
			lo = i
			topPad = blockStart
		}
		if blockStart < hiBound {
			hi = i + 1
			endHi = blockEnd
		}
	}
	if lo == len(msgs) {
		return len(msgs), len(msgs), total, 0
	}
	if hi < lo {
		hi = lo
		endHi = topPad
	}
	botPad = total - endHi
	if botPad < 0 {
		botPad = 0
	}
	return lo, hi, topPad, botPad
}
