package main

import (
	"context"
	"strings"
	"testing"
)

func TestZoomKeysScaleLayout(t *testing.T) {
	app := newTestApp(t)
	ctx := context.Background()

	if got := app.zoomLabel(); got != "100%" {
		t.Fatalf("zoom label = %q, want 100%%", got)
	}

	rowsBefore := len(app.View().Rows)

	if err := app.onKeyDown(ctx, "control"); err != nil {
		t.Fatal(err)
	}

	if err := app.onKeyDown(ctx, "equal"); err != nil {
		t.Fatal(err)
	}

	if app.uiZoom() <= 1 {
		t.Fatalf("zoom after Ctrl+= = %v, want above 1", app.uiZoom())
	}

	if got := app.zoomLabel(); got == "100%" {
		t.Fatalf("zoom label stayed %q after zooming in", got)
	}

	if err := app.onKeyDown(ctx, "0"); err != nil {
		t.Fatal(err)
	}

	if app.uiZoom() != 1 {
		t.Fatalf("zoom after Ctrl+0 = %v, want 1", app.uiZoom())
	}

	if err := app.onKeyDown(ctx, "minus"); err != nil {
		t.Fatal(err)
	}

	if app.uiZoom() >= 1 {
		t.Fatalf("zoom after Ctrl+- = %v, want below 1", app.uiZoom())
	}

	if err := app.onKeyDown(ctx, "0"); err != nil {
		t.Fatal(err)
	}

	if len(app.View().Rows) != rowsBefore {
		t.Fatal("reset did not restore the row window")
	}
}

func TestZoomKeysNeedControl(t *testing.T) {
	app := newTestApp(t)
	ctx := context.Background()

	for _, key := range []string{"equal", "minus", "0", "kpadd", "kpsubtract", "kp0"} {
		if err := app.onKeyDown(ctx, key); err != nil {
			t.Fatal(err)
		}
	}

	if app.uiZoom() != 1 {
		t.Fatalf("zoom without control = %v, want 1", app.uiZoom())
	}
}

func TestZoomCopyUnaffected(t *testing.T) {
	app := newTestApp(t)
	ctx := context.Background()

	if err := app.onKeyDown(ctx, "control"); err != nil {
		t.Fatal(err)
	}

	if err := app.onKeyDown(ctx, "c"); err != nil {
		t.Fatal(err)
	}

	if app.uiZoom() != 1 {
		t.Fatalf("zoom after Ctrl+C = %v, want 1", app.uiZoom())
	}
}

func TestZoomWheelNeedsControl(t *testing.T) {
	app := newTestApp(t)
	ctx := context.Background()

	if _, err := app.onWheel(ctx, 0, 0, 0, 1); err != nil {
		t.Fatal(err)
	}

	if app.uiZoom() != 1 {
		t.Fatalf("zoom after plain wheel = %v, want 1", app.uiZoom())
	}

	app.ctrl = true

	if _, err := app.onWheel(ctx, 0, 0, 0, 1); err != nil {
		t.Fatal(err)
	}

	if app.uiZoom() <= 1 {
		t.Fatalf("zoom after Ctrl+wheel = %v, want above 1", app.uiZoom())
	}
}

func TestZoomPersistsInSession(t *testing.T) {
	app := newTestApp(t)
	ctx := context.Background()
	app.sessionReady = true

	if err := app.onKeyDown(ctx, "control"); err != nil {
		t.Fatal(err)
	}

	if err := app.onKeyDown(ctx, "equal"); err != nil {
		t.Fatal(err)
	}

	state := app.sessionState()

	if state.Zoom <= 1 {
		t.Fatalf("session zoom = %v, want above 1", state.Zoom)
	}

	restored := state.sanitize(app.ws.Root)

	if restored.Zoom != state.Zoom {
		t.Fatalf("restored zoom = %v, want %v", restored.Zoom, state.Zoom)
	}
}

func TestZoomOverrideTracksLevel(t *testing.T) {
	if got := zoomOverrideCSS(1); got != "" {
		t.Fatalf("100 percent override = %q, want empty", got)
	}

	css := zoomOverrideCSS(1.5)

	for _, want := range []string{
		"--row-h:30px",
		"--titlebar-h:63px",
		"--side-w:402px",
		".trow",
		".cline",
		// Font sizes stay integral: the replay shapes on an integer ppem
		// grid, so a fractional size paints wider than the layout frame.
		"font-size:18px",
		// Percent and negative values must render literally, not as verbs.
		"left:calc(50% - 420px)",
		"margin-right:-15px",
		".hist-graph svg",
		// Scaled chrome the reports flagged.
		".changes-title",
		".side-title",
		".win-btn",
		".search-row",
		".hist-row",
		// Activity bar, compact history rail, and crumb bar.
		".activity",
		".act svg",
		".act-badge",
		".side-tools",
		"#side-refresh svg",
		".side-history .hist-hash",
		".crumbs",
		".crumb-btn",
		".tab-tool svg",
		".settings-nav-active",
		".settings-theme",
		".file-nav-current",
		"@media (max-width: 900px)",
	} {
		if !strings.Contains(css, want) {
			t.Fatalf("override missing %q", want)
		}
	}

	// The diagram row is a var() row now, so no fixed height may survive.
	if strings.Contains(css, ".pdiag { height") {
		t.Fatal("override still pins a fixed .pdiag height")
	}
}
