package main

import (
	"context"
	"testing"
)

func resizeWin(t *testing.T, a *app, w, h int) {
	t.Helper()
	a.page.SetSize(w, h)
	a.redraw()
	if a.lastErr != nil {
		t.Fatal(a.lastErr)
	}
}

func setZoomTo(t *testing.T, a *app, pct int) {
	t.Helper()
	a.x.zoom = pct
	a.redraw()
	if a.lastErr != nil {
		t.Fatal(a.lastErr)
	}
}

func TestIZoomCtrlWheelAndChords(t *testing.T) {
	a, m := newTestApp(t)
	ctx := context.Background()
	if a.zoomPct() != 100 {
		t.Fatalf("default zoom %d", a.zoomPct())
	}
	m.ctrl = true
	if handled, err := a.handleWheel(ctx, 0, 0, 0, 1); !handled || err != nil {
		t.Fatalf("ctrl+wheel handled=%v err=%v", handled, err)
	}
	if a.zoomPct() != 110 {
		t.Fatalf("wheel up gave %d", a.zoomPct())
	}
	if _, err := a.handleWheel(ctx, 0, 0, 0, -1); err != nil {
		t.Fatal(err)
	}
	if a.zoomPct() != 100 || a.scrollRow != 0 {
		t.Fatalf("wheel down gave %d, scrollRow %d", a.zoomPct(), a.scrollRow)
	}
	m.ctrl = false
	for _, k := range []string{"ctrl+=", "ctrl+equal", "ctrl+shift+equal", "ctrl+kpadd"} {
		before := a.zoomPct()
		if err := a.handleKey(ctx, k); err != nil {
			t.Fatal(err)
		}
		if a.zoomPct() != before+zoomStepPct {
			t.Fatalf("%s: %d -> %d", k, before, a.zoomPct())
		}
		a.x.zoom = 100
	}
	if err := a.handleKey(ctx, "ctrl+-"); err != nil {
		t.Fatal(err)
	}
	if a.zoomPct() != 90 {
		t.Fatalf("ctrl+- gave %d", a.zoomPct())
	}
	if err := a.handleKey(ctx, "ctrl+0"); err != nil {
		t.Fatal(err)
	}
	if a.zoomPct() != 100 {
		t.Fatalf("ctrl+0 gave %d", a.zoomPct())
	}
	for range 30 {
		a.zoomKey("ctrl+=")
	}
	if a.zoomPct() != zoomMaxPct {
		t.Fatalf("no clamp at max: %d", a.zoomPct())
	}
	for range 30 {
		a.zoomKey("ctrl+-")
	}
	if a.zoomPct() != zoomMinPct {
		t.Fatalf("no clamp at min: %d", a.zoomPct())
	}
	if got := mustBox(t, a, "xl-status-zoom"); got.Text != "50%" {
		t.Fatalf("status readout %q", got.Text)
	}
}

func TestIZoomChromeScales(t *testing.T) {
	a, _ := newTestApp(t)
	heights := func() (title, grid, status, tabs, fbar float64) {
		return mustBox(t, a, "xl-titlebar").H, mustBox(t, a, "xl-grid").Y,
			mustBox(t, a, "xl-status").H, mustBox(t, a, "xl-tabs").H, mustBox(t, a, "xl-fbar").H
	}
	title, gridY, status, tabs, fbar := heights()
	if int(title) != titleBarHeight || int(status) != statusBarHeight || int(tabs) != sheetBarHeight || int(fbar) != formulaBarHeight {
		t.Fatalf("100%%: title %v status %v tabs %v fbar %v", title, status, tabs, fbar)
	}
	setZoomTo(t, a, 150)
	title2, gridY2, status2, tabs2, fbar2 := heights()
	if int(title2) != a.zs(titleBarHeight) || int(status2) != a.zs(statusBarHeight) ||
		int(tabs2) != a.zs(sheetBarHeight) || int(fbar2) != a.zs(formulaBarHeight) {
		t.Fatalf("150%%: title %v status %v tabs %v fbar %v", title2, status2, tabs2, fbar2)
	}
	if int(gridY2) != int(float64(gridY)*1.5) {
		t.Fatalf("grid origin %v at 150%%, %v at 100%%", gridY2, gridY)
	}
	if ox, oy := a.gridOrigin(); int(oy) != a.zs(titleBarHeight)+a.ribbonHeight()+a.fbarH() || ox != 0 {
		t.Fatalf("gridOrigin %v,%v", ox, oy)
	}
}

func TestIZoomGridFitsWindow(t *testing.T) {
	a, _ := newTestApp(t)
	for _, z := range []int{50, 100, 150, 200} {
		setZoomTo(t, a, z)
		winW, winH := a.viewportSize()
		grid := mustBox(t, a, "xl-grid")
		status := mustBox(t, a, "xl-status")
		tabs := mustBox(t, a, "xl-tabs")
		if grid.Y+grid.H > tabs.Y+0.5 || tabs.Y+tabs.H > status.Y+0.5 || status.Y+status.H > float64(winH)+0.5 {
			t.Fatalf("%d%%: bands overflow: grid %v..%v tabs %v status %v..%v win %d", z, grid.Y, grid.Y+grid.H, tabs.Y, status.Y, status.Y+status.H, winH)
		}
		if grid.X+grid.W > float64(winW)+0.5 {
			t.Fatalf("%d%%: grid wider than window", z)
		}
		vw, vh := a.gridViewportSize()
		if int(grid.W) != vw || int(grid.H+0.5) != vh {
			t.Fatalf("%d%%: viewport %dx%d vs grid box %vx%v", z, vw, vh, grid.W, grid.H)
		}
		cw, ch := a.canvasSize()
		for _, s := range a.layoutCols(cw) {
			if s.pos+s.size > cw {
				t.Fatalf("%d%%: column %d overflows canvas", z, s.idx)
			}
		}
		for _, s := range a.layoutRows(ch) {
			if s.pos+s.size > ch {
				t.Fatalf("%d%%: row %d overflows canvas", z, s.idx)
			}
		}
	}
}

func TestIZoomClickSelectsCellAt150(t *testing.T) {
	a, _ := newTestApp(t)
	setZoomTo(t, a, 150)
	for _, c := range []CellRef{{Row: 3, Col: 2}, {Row: 7, Col: 5}} {
		clickCell(t, a, c)
		if a.active != c {
			t.Fatalf("click on %v selected %v", c, a.active)
		}
	}
}

func TestIZoomDefaultByWindowWidth(t *testing.T) {
	cases := []struct{ w, want int }{{1440, 100}, {1799, 100}, {1800, 125}, {2399, 125}, {2400, 150}, {2560, 150}}
	for _, c := range cases {
		if got := autoZoomPct(c.w); got != c.want {
			t.Errorf("autoZoomPct(%d) = %d, want %d", c.w, got, c.want)
		}
	}
	a, _ := newTestApp(t)
	resizeWin(t, a, 2560, 1440)
	if a.zoomPct() != 150 {
		t.Fatalf("2560 wide window zoom %d", a.zoomPct())
	}
	if h := mustBox(t, a, "xl-titlebar").H; int(h) != a.zs(titleBarHeight) {
		t.Fatalf("title bar %v did not follow the default scale", h)
	}
	if got := mustBox(t, a, "xl-status-zoom"); got.Text != "150%" {
		t.Fatalf("status readout %q", got.Text)
	}
	// A choice by the user wins over the width default.
	a.zoomKey("ctrl+0")
	resizeWin(t, a, 2600, 1440)
	if a.zoomPct() != 100 {
		t.Fatalf("explicit 100%% overridden: %d", a.zoomPct())
	}
}

// TestIZoomChordWhileEditing: Ctrl+= zooms even while a cell is in edit mode,
// and the edit buffer is left alone.
func TestIZoomChordWhileEditing(t *testing.T) {
	a, m := newTestApp(t)
	a.setZoom(100)
	a.beginEdit("abc")
	m.ctrl = true
	if err := a.handleKey(context.Background(), "="); err != nil {
		t.Fatal(err)
	}
	m.ctrl = false
	if got := a.zoomPct(); got <= 100 {
		t.Fatalf("zoom %d%% after Ctrl+= while editing, want above 100", got)
	}
	if !a.editing || a.editBuf != "abc" {
		t.Fatalf("edit state changed: editing=%v buf=%q", a.editing, a.editBuf)
	}
}

// TestTickRescalesWhenWindowGrows: the window reaches its real size after the
// first build, so the frame tick must apply the width based default zoom.
func TestTickRescalesWhenWindowGrows(t *testing.T) {
	a, _ := newTestApp(t)
	ctx := context.Background()
	_ = a.onTick(ctx)
	a.page.SetSize(2560, 1440)
	if err := a.onTick(ctx); err != nil {
		t.Fatal(err)
	}
	if a.x.themePct != 150 {
		t.Fatalf("theme built at %d%%, want 150%% for a 2560 px window", a.x.themePct)
	}
}
