package main

import (
	"bytes"
	"context"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/chinmay-sawant/ownframe"
)

func TestSidebarHoverPreservesPaintContext(t *testing.T) {
	a, err := newApp()
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	a.page.Window().SetTransitions(false)
	if err := a.page.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	for _, action := range []string{"nav:overview", "nav:logs", "nav:overview"} {
		var x, y float64
		found := false
		for _, box := range a.page.Boxes() {
			if box.Action == action {
				x, y, found = box.X+box.W/2, box.Y+box.H/2, true
				break
			}
		}
		if !found {
			t.Fatalf("sidebar control %q disappeared from the hit boxes", action)
		}
		if err := a.page.Edit().Hover(ctx, x, y); err != nil {
			t.Fatal(err)
		}
		patched := sidebarPaintContext(a)
		patchedOrder := sidebarPaintOrder(a)
		patchedPNG := a.page.Probe().PNG()
		if err := a.page.Redraw(ctx); err != nil {
			t.Fatal(err)
		}
		full := sidebarPaintContext(a)
		if !slices.Equal(patchedOrder, sidebarPaintOrder(a)) {
			t.Fatalf("sidebar paint order after %s differs:\npatched=%v\nfull=%v", action, patchedOrder, sidebarPaintOrder(a))
		}
		if !bytes.Equal(patchedPNG, a.page.Probe().PNG()) {
			t.Fatalf("hover %s differs from a full render", action)
		}
		for text, want := range full {
			if got, ok := patched[text]; !ok || got != want {
				t.Fatalf("%s after hover %s: patched=%+v full=%+v", text, action, got, want)
			}
		}
		if len(patched) != 2 || len(full) != 2 {
			t.Fatalf("missing sidebar text: patched=%v full=%v", patched, full)
		}
	}
}

func TestHoverSurfacesStayOpenAcrossTheirContent(t *testing.T) {
	a, err := newApp()
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if err := a.page.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	initialHeader := hoverHeaderPositions(a.page)
	before := a.page.Probe().Stats()
	for _, target := range []struct {
		triggerID string
		contentID string
	}{
		{triggerID: "tooltip-trigger", contentID: "live-tooltip-content"},
		{triggerID: "hover-trigger", contentID: "oncall-hover-content"},
	} {
		trigger, found := findHoverTestBox(a.page, func(box ownframe.Box) bool { return box.ID == target.triggerID })
		if !found {
			t.Fatalf("trigger %q missing from boxes: %+v", target.triggerID, a.page.Boxes())
		}
		if err := a.page.Edit().Hover(ctx, trigger.X+trigger.W/2, trigger.Y+trigger.H/2); err != nil {
			t.Fatal(err)
		}
		if positions := hoverHeaderPositions(a.page); !slices.EqualFunc(initialHeader, positions, func(left, right hoverHeaderBox) bool {
			return left.ID == right.ID && left.X == right.X && left.Y == right.Y && left.W == right.W && left.H == right.H
		}) {
			t.Fatalf("header controls shifted when %s opened: before=%+v after=%+v", target.triggerID, initialHeader, positions)
		}
		content, found := findHoverTestBox(a.page, func(box ownframe.Box) bool { return box.ID == target.contentID })
		if !found {
			t.Fatalf("content %q missing from boxes", target.contentID)
		}
		if content.W <= 0 || content.H <= 0 {
			t.Fatalf("content %q has no visible bounds: %+v", target.contentID, content)
		}
		if err := a.page.Edit().Hover(ctx, content.X+content.W/2, content.Y+content.H/2); err != nil {
			t.Fatal(err)
		}
		if _, found := findHoverTestBox(a.page, func(box ownframe.Box) bool { return box.ID == target.contentID }); !found {
			t.Fatalf("content %q closed while the pointer was over its popup", target.contentID)
		}
		if err := a.page.Edit().Hover(ctx, 900, 500); err != nil {
			t.Fatal(err)
		}
		if _, found := findHoverTestBox(a.page, func(box ownframe.Box) bool { return box.ID == target.contentID }); found {
			t.Fatalf("content %q stayed open after the pointer left its popup", target.contentID)
		}
	}
	after := a.page.Probe().Stats()
	if after.Parses != before.Parses || after.Cascades != before.Cascades {
		t.Fatalf("hover rebuilt the page: parses %d -> %d, cascades %d -> %d", before.Parses, after.Parses, before.Cascades, after.Cascades)
	}
}

func TestTopbarMenusAlignAndAnchorPanels(t *testing.T) {
	a, err := newApp()
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if err := a.page.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	platform := loglensLayoutBox(t, a.page, func(box ownframe.Box) bool { return box.ID == "product-platform-trigger" })
	file := loglensLayoutBox(t, a.page, func(box ownframe.Box) bool { return box.ID == "menu-file-trigger" })
	if platform.Y != file.Y {
		t.Fatalf("menu rows do not align vertically: Platform y=%v File y=%v", platform.Y, file.Y)
	}
	resources := loglensLayoutBox(t, a.page, func(box ownframe.Box) bool { return box.ID == "product-resources-trigger" })
	if err := a.page.Click(ctx, resources.X+resources.W/2, resources.Y+resources.H/2); err != nil {
		t.Fatal(err)
	}
	panel := loglensLayoutBox(t, a.page, func(box ownframe.Box) bool { return box.ID == "product-resources-content" })
	nav := loglensLayoutBox(t, a.page, func(box ownframe.Box) bool { return box.ID == "top-nav" })
	if panel.X != nav.X {
		t.Fatalf("Resources panel left=%v, want its menu left=%v", panel.X, nav.X)
	}
}

func TestLiveTooltipShowsCurrentUpdateAge(t *testing.T) {
	a, err := newApp()
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if err := a.page.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	// Stamp the update age after the redraw: under -race the redraw above
	// can take over a second, which used to push the hover label into the
	// next second bucket and flake the exact-match assertion below.
	a.lastUpdatedAt = time.Now().Add(-13 * time.Second)
	trigger := loglensLayoutBox(t, a.page, func(box ownframe.Box) bool { return box.ID == "tooltip-trigger" })
	before := a.page.Probe().Stats()
	if err := a.page.Edit().Hover(ctx, trigger.X+trigger.W/2, trigger.Y+trigger.H/2); err != nil {
		t.Fatal(err)
	}
	tip := loglensLayoutBox(t, a.page, func(box ownframe.Box) bool { return box.ID == "live-tooltip-content" })
	// The hover handler labels the tooltip with its own time.Now(), so a
	// second boundary between the hover and this assertion can age the
	// label by one second. Accept either side of that boundary.
	age := int(time.Since(a.lastUpdatedAt).Seconds())
	want := "Updated " + strconv.Itoa(age) + " seconds ago"
	fallback := "Updated " + strconv.Itoa(age-1) + " seconds ago"
	if tip.Text != want && tip.Text != fallback {
		t.Fatalf("live tooltip says %q, want %q (or %q across a second boundary)", tip.Text, want, fallback)
	}
	after := a.page.Probe().Stats()
	if after.Parses != before.Parses || after.Cascades != before.Cascades {
		t.Fatalf("updating the tooltip rebuilt the page: parses %d -> %d, cascades %d -> %d", before.Parses, after.Parses, before.Cascades, after.Cascades)
	}
}

func TestInboxPopoverDoesNotMoveRightHeaderControls(t *testing.T) {
	a, err := newApp()
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if err := a.page.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	before := hoverHeaderPositions(a.page)
	trigger := loglensLayoutBox(t, a.page, func(box ownframe.Box) bool { return box.ID == "notifications-popover" })
	if err := a.page.Click(ctx, trigger.X+trigger.W/2, trigger.Y+trigger.H/2); err != nil {
		t.Fatal(err)
	}
	if !a.notifOpen {
		t.Fatal("inbox popover did not open")
	}
	after := hoverHeaderPositions(a.page)
	if !slices.EqualFunc(before, after, func(left, right hoverHeaderBox) bool {
		return left.ID == right.ID && left.X == right.X && left.Y == right.Y && left.W == right.W && left.H == right.H
	}) {
		t.Fatalf("header controls shifted when inbox opened: before=%+v after=%+v", before, after)
	}
}

type hoverHeaderBox struct {
	ID   string
	X, Y float64
	W, H float64
}

func hoverHeaderPositions(page *ownframe.Page) []hoverHeaderBox {
	wanted := map[string]bool{
		"tooltip-trigger": true, "hover-trigger": true, "notifications-popover": true,
		"user-menu": true,
	}
	positions := make([]hoverHeaderBox, 0, len(wanted))
	for _, box := range page.Boxes() {
		if wanted[box.ID] {
			positions = append(positions, hoverHeaderBox{ID: box.ID, X: box.X, Y: box.Y, W: box.W, H: box.H})
		}
	}
	slices.SortFunc(positions, func(a, b hoverHeaderBox) int {
		return strings.Compare(a.ID, b.ID)
	})
	return positions
}

func findHoverTestBox(page *ownframe.Page, match func(ownframe.Box) bool) (ownframe.Box, bool) {
	for _, box := range page.Boxes() {
		if match(box) {
			return box, true
		}
	}
	return ownframe.Box{}, false
}

func sidebarPaintOrder(a *app) []string {
	order := []string{}
	d := a.page.Probe().Display()
	for _, i := range d.Order {
		op := d.Ops[i]
		if op.Fixed && op.X < 180 && op.Y >= 45 && op.Y < 270 {
			order = append(order, fmt.Sprintf("%v %q %.2f,%.2f %.2fx%.2f %.2f,%.2f,%.2f,%.2f", op.Kind, op.Text, op.X, op.Y, op.W, op.H, op.R, op.G, op.B, op.Alpha))
		}
	}
	return order
}

type sidebarContext struct {
	fixed bool
	zSet  bool
	z     int
}

func sidebarPaintContext(a *app) map[string]sidebarContext {
	contexts := map[string]sidebarContext{}
	for _, op := range a.page.Probe().Display().Ops {
		if op.X < 180 && (op.Text == "Overview" || op.Text == "Logs") {
			contexts[op.Text] = sidebarContext{fixed: op.Fixed, zSet: op.ZIndexSet, z: op.ZIndex}
		}
	}
	return contexts
}
