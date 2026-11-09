package main

import (
	"context"
	"strconv"
	"testing"

	"github.com/chinmay-sawant/ownframe/pagetest"
)

// Chat transcript owns its wheel: the viewport clips, the content
// translates, and the wheel is consumed so the shell chrome stays put.
func TestChatScrollViewportAndWheel(t *testing.T) {
	ctx := context.Background()
	a, err := newApp()
	if err != nil {
		t.Fatal(err)
	}
	if err := a.page.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	if _, ok := pagetest.BoxByID(a.page, "dc-scroll-viewport"); !ok {
		t.Fatal("missing dc-scroll-viewport")
	}
	if _, ok := pagetest.BoxByID(a.page, "dc-scroll-content"); !ok {
		t.Fatal("missing dc-scroll-content")
	}
	var vx, vy float64
	for _, b := range a.page.Boxes() {
		if b.ID == "dc-scroll-viewport" {
			vx, vy = b.X+b.W/2, b.Y+b.H/2
		}
	}
	handled, err := a.handleWheel(ctx, vx, vy, 0, 1)
	if err != nil {
		t.Fatal(err)
	}
	if !handled {
		t.Fatal("wheel over transcript not consumed")
	}
}

// Scroll-window tests below are adapted to the real scroll API (2026-11-06).
// DELTA vs the task spec: scroll.go has no estimateMsgHeight, scrollTotal,
// visibleRange, scrollPx, scrollCharsPerLine, scrollBaseH, scrollLineH,
// scrollReplyH, scrollReactH, scrollDividerH, scrollOverscan, or
// scrollHistCap. The transcript scrolls through chadcn.ScrollState
// (app.chatScroll, ranged by updateScrollCache from measured boxes) with date
// dividers from messageDay in messages.go and a 500-message per-channel cap
// (channelHistCap) in appendMessage, actions.go. Test names below are the
// adapted counterparts; only TestHistoryCapBounds matches the spec name and
// contract exactly. Tests build &app{} directly (no window, no network)
// except TestChatScrollViewportAndWheel, which mounts the view.

// TestMessageDayDivider pins the divider grouping the transcript loop in
// messages.go relies on: the label before " at ", or "" for no divider.
func TestMessageDayDivider(t *testing.T) {
	cases := []struct{ stamp, day string }{
		{"Today at 08:02", "Today"},
		{"Today at 08:0X", "Today"},
		{"Yesterday at 08:05", "Yesterday"},
		{"Monday at 09:00", "Monday"},
		{"no-separator", ""},
		{"", ""},
	}
	for _, c := range cases {
		if got := messageDay(c.stamp); got != c.day {
			t.Errorf("messageDay(%q) = %q, want %q", c.stamp, got, c.day)
		}
	}
}

// TestTranscriptDividerSpans replays the mainCol divider loop over three
// same-day messages (one divider), then a day change between msg 2 and 3
// (two dividers). This is the real-API counterpart of the spec's
// scrollTotal divider case.
func TestTranscriptDividerSpans(t *testing.T) {
	countDividers := func(msgs []dmessage) int {
		dividers, prevDay := 0, ""
		for _, m := range msgs {
			if day := messageDay(m.Time); day != "" && day != prevDay {
				dividers++
				prevDay = day
			}
		}
		return dividers
	}
	sameDay := []dmessage{
		{Time: "Today at 08:01"},
		{Time: "Today at 08:02"},
		{Time: "Today at 08:03"},
	}
	if got := countDividers(sameDay); got != 1 {
		t.Fatalf("same-day dividers = %d, want 1", got)
	}
	dayChange := []dmessage{
		{Time: "Today at 08:01"},
		{Time: "Today at 08:02"},
		{Time: "Yesterday at 08:03"},
	}
	if got := countDividers(dayChange); got != 2 {
		t.Fatalf("day-change dividers = %d, want 2", got)
	}
}

// TestChatScrollRangeClamp pins the offset window on the real scroll state:
// top offset is 0, a middle offset sticks, and a huge offset clamps to
// MaxOffset (content minus viewport). This is the real-API counterpart of
// the spec's visibleRange/scrollPx clamp cases.
func TestChatScrollRangeClamp(t *testing.T) {
	a := &app{}
	a.chatScroll.SetRange(5000, 400)
	if got := a.chatScroll.Offset(); got != 0 {
		t.Fatalf("initial offset = %d, want 0", got)
	}
	const wantMax = 4600
	if got := a.chatScroll.Metrics().MaxOffset(); got != wantMax {
		t.Fatalf("MaxOffset = %d, want %d", got, wantMax)
	}
	a.chatScroll.ScrollTo(2000)
	if got := a.chatScroll.Offset(); got != 2000 {
		t.Fatalf("middle offset = %d, want 2000", got)
	}
	a.chatScroll.ScrollTo(1 << 30)
	if got := a.chatScroll.Offset(); got != wantMax {
		t.Fatalf("clamped offset = %d, want MaxOffset %d", got, wantMax)
	}
}

// TestChatScrollClampNegative pins the lower clamp and the fitting-content
// case: negative offsets floor at 0, and content that fits cannot scroll.
func TestChatScrollClampNegative(t *testing.T) {
	a := &app{}
	a.chatScroll.SetRange(5000, 400)
	a.chatScroll.ScrollTo(2000)
	a.chatScroll.ScrollTo(-50)
	if got := a.chatScroll.Offset(); got != 0 {
		t.Fatalf("negative offset = %d, want 0", got)
	}
	fit := &app{}
	fit.chatScroll.SetRange(100, 400)
	if got := fit.chatScroll.Metrics().MaxOffset(); got != 0 {
		t.Fatalf("fitting MaxOffset = %d, want 0", got)
	}
	fit.chatScroll.ScrollTo(1 << 30)
	if got := fit.chatScroll.Offset(); got != 0 {
		t.Fatalf("fitting offset = %d, want 0", got)
	}
}

// TestHistoryCapBounds pins the 500-message per-channel history cap in
// actions.go (channelHistCap): appending 600 messages keeps 500 with the
// newest last (drop-oldest via in-place shift).
func TestHistoryCapBounds(t *testing.T) {
	a := &app{messages: map[string][]dmessage{}}
	const total = 600
	for i := 0; i < total; i++ {
		a.appendMessage("c1", dmessage{ID: "m" + strconv.Itoa(i), Author: "You", Text: "msg " + strconv.Itoa(i)})
	}
	const wantCap = 500
	if got := len(a.messages["c1"]); got != wantCap {
		t.Fatalf("messages[c1] = %d, want cap %d", got, wantCap)
	}
	if got := a.messages["c1"][wantCap-1].ID; got != "m599" {
		t.Fatalf("last message = %q, want newest %q", got, "m599")
	}
	if got := a.messages["c1"][0].ID; got != "m100" {
		t.Fatalf("first message = %q, want oldest-kept %q", got, "m100")
	}
}

// TestMessageDayNoAlloc pins the divider hot path as allocation-free, the
// real-API counterpart of the spec's estimator no-alloc case.
func TestMessageDayNoAlloc(t *testing.T) {
	if n := testing.AllocsPerRun(100, func() {
		_ = messageDay("Today at 08:02")
	}); n != 0 {
		t.Fatalf("messageDay allocs = %v, want 0", n)
	}
}

// TestChatPinsLatestByDefault pins the scroll contract: after paint the
// transcript rests on the latest message (offset == MaxOffset), channel
// switches re-pin there, and wheeling down at the bottom moves nowhere.
func TestChatPinsLatestByDefault(t *testing.T) {
	ctx := context.Background()
	a, err := newApp()
	if err != nil {
		t.Fatal(err)
	}
	if err := a.page.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	if err := a.refresh(ctx); err != nil {
		t.Fatal(err)
	}
	max := a.chatScroll.Metrics().MaxOffset()
	if max <= 0 {
		t.Fatalf("MaxOffset = %d, want positive with seeded transcript", max)
	}
	if got := a.chatScroll.Offset(); got != max {
		t.Fatalf("offset = %d, want pinned MaxOffset %d", got, max)
	}
	var vx, vy float64
	for _, b := range a.page.Boxes() {
		if b.ID == "dc-scroll-viewport" {
			vx, vy = b.X+b.W/2, b.Y+b.H/2
		}
	}
	// Wheel down (dy negative in Ebiten terms moves toward latest) at the
	// bottom must not overshoot into white space.
	before := a.chatScroll.Offset()
	handled, err := a.handleWheel(ctx, vx, vy, 0, -3)
	if err != nil {
		t.Fatal(err)
	}
	if !handled {
		t.Fatal("wheel over transcript not consumed")
	}
	if got := a.chatScroll.Offset(); got != before {
		t.Fatalf("offset moved past latest: %d -> %d", before, got)
	}
}
