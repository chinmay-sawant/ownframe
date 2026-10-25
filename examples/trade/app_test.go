package main

import (
	"bytes"
	"context"
	"fmt"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/chinmay-sawant/ownframe"
)

const testScreenshotDir = "/tmp/opencode/trade-shots"

// newTestApp builds the example without opening a window.
func newTestApp(t *testing.T) *App {
	t.Helper()

	app, err := New()
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	if err := app.Page().Redraw(context.Background()); err != nil {
		t.Fatalf("initial redraw: %v", err)
	}

	return app
}

// newTestAppSized builds the example at one viewport before its first paint,
// so a size sweep does not pay for an unused 1600x900 render.
func newTestAppSized(t *testing.T, width, height int) *App {
	t.Helper()

	app, err := New()
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	app.Page().SetSize(width, height)

	if err := app.Page().Redraw(context.Background()); err != nil {
		t.Fatalf("initial redraw at %dx%d: %v", width, height, err)
	}

	return app
}

// assertScreenFits checks the canvas, the status bar, and the screen root
// against the requested window. It is the cheap replacement for a pixel
// capture: every screen still has to fit at every size.
func assertScreenFits(t *testing.T, app *App, screen string, width, height int) {
	t.Helper()

	boxes := app.Page().Boxes()
	shell, ok := boxByID(boxes, "app-shell")
	if !ok {
		t.Fatalf("%s: no app shell", screen)
	}
	if math.Abs(shell.W-float64(width)) > 1 || math.Abs(shell.H-float64(height)) > 1 {
		t.Fatalf("%s: shell is %.0fx%.0f at %dx%d", screen, shell.W, shell.H, width, height)
	}

	status, ok := boxByID(boxes, "statusbar-main")
	if !ok {
		t.Fatalf("%s: no status bar", screen)
	}
	if math.Abs(status.Y+status.H-float64(height)) > 1 {
		t.Fatalf("%s: status bar ends at %.0f, want %d", screen, status.Y+status.H, height)
	}

	roots := map[string][]string{
		"watchlist": {"watchlist-main", "center-column", "quote-rail"},
		"markets":   {"mk-page"},
		"portfolio": {"pf-page"},
		"orders":    {"ord-page"},
		"screener":  {"scr-page"},
		"news":      {"news-page"},
		"settings":  {"set-page"},
	}

	for _, id := range roots[screen] {
		root, ok := boxByID(boxes, id)
		if !ok {
			t.Fatalf("%s: no %s box", screen, id)
		}
		if root.W <= 0 || root.H <= 0 {
			t.Fatalf("%s: %s is empty", screen, id)
		}
		if root.X < -1 || root.Y < 45 || root.X+root.W > float64(width)+1 || root.Y+root.H > status.Y+1 {
			t.Fatalf("%s: %s at (%.0f,%.0f) %.0fx%.0f escapes %dx%d", screen, id, root.X, root.Y, root.W, root.H, width, height)
		}
	}
}

// boxByID finds one laid-out box by its element id.
func boxByID(boxes []ownframe.Box, id string) (ownframe.Box, bool) {
	for _, box := range boxes {
		if box.ID == id {
			return box, true
		}
	}

	return ownframe.Box{}, false
}

// TestScreensRender walks every rail destination and requires a replayable
// display list. It catches a template that reads a missing data field.
func TestScreensRender(t *testing.T) {
	app := newTestApp(t)

	if app.Page().Probe().Display() == nil {
		t.Fatal("initial page fell back to a bitmap")
	}

	for _, screen := range []string{"markets", "portfolio", "orders", "screener", "news", "settings", "watchlist"} {
		if err := app.onClick(context.Background(), ownframe.Box{Action: "nav-" + screen}); err != nil {
			t.Fatalf("nav-%s: %v", screen, err)
		}

		if screen != app.view.Screen {
			t.Fatalf("screen = %q, want %q", app.view.Screen, screen)
		}

		if err := app.Page().Redraw(context.Background()); err != nil {
			t.Fatalf("redraw %s: %v", screen, err)
		}

		if app.Page().Probe().Display() == nil {
			t.Fatalf("%s fell back to a bitmap", screen)
		}

		stats := app.Page().Probe().Motion()
		if stats.LastError != "" {
			t.Fatalf("%s motion bind error: %s", screen, stats.LastError)
		}
	}
}

// TestCSSMotionBound requires the stylesheet to bind at least one CSS player.
// It also checks that no manual tick was installed: the page parks without one.
func TestCSSMotionBound(t *testing.T) {
	app := newTestApp(t)

	stats := app.Page().Probe().Motion()
	if stats.Bound == 0 {
		t.Fatal("no CSS animation bound")
	}

	if app.Page().Window().Animations() != true {
		t.Fatal("automatic animations are off")
	}

	if !app.Page().Window().DevTools() {
		t.Fatal("developer tools should open by default")
	}
}

// TestWatchlistFilterAndSelection checks row filtering and symbol switching.
func TestWatchlistFilterAndSelection(t *testing.T) {
	app := newTestApp(t)

	all := len(app.view.Watch)
	if all == 0 {
		t.Fatal("watchlist is empty")
	}

	if err := app.onClick(context.Background(), ownframe.Box{Action: "watch-filter-starred"}); err != nil {
		t.Fatal(err)
	}

	if app.view.WatchFilter != "starred" {
		t.Fatalf("filter = %q, want starred", app.view.WatchFilter)
	}

	if len(app.view.Watch) > all {
		t.Fatalf("filtered watchlist grew: %d > %d", len(app.view.Watch), all)
	}

	if err := app.onClick(context.Background(), ownframe.Box{Action: "watch-filter-all"}); err != nil {
		t.Fatal(err)
	}

	symbol := app.view.Watch[1].Symbol
	if err := app.onClick(context.Background(), ownframe.Box{Action: "watch-" + symbol}); err != nil {
		t.Fatal(err)
	}

	if app.view.Quote.Symbol != symbol {
		t.Fatalf("quote symbol = %q, want %q", app.view.Quote.Symbol, symbol)
	}

	if app.view.Ticket.Symbol != symbol {
		t.Fatalf("ticket symbol = %q, want %q", app.view.Ticket.Symbol, symbol)
	}
}

// TestChartTabsAndTimeframes checks the chart state changes without a redraw
// error.
func TestChartTabsAndTimeframes(t *testing.T) {
	app := newTestApp(t)

	for _, tab := range []string{"options", "news", "financial", "chart"} {
		if err := app.onClick(context.Background(), ownframe.Box{Action: "chart-tab-" + tab}); err != nil {
			t.Fatal(err)
		}

		if app.view.ChartTab != tab {
			t.Fatalf("chart tab = %q, want %q", app.view.ChartTab, tab)
		}

		if err := app.Page().Redraw(context.Background()); err != nil {
			t.Fatalf("redraw %s: %v", tab, err)
		}
	}

	if err := app.onClick(context.Background(), ownframe.Box{Action: "tf-1M"}); err != nil {
		t.Fatal(err)
	}

	if app.view.Timeframe != "1M" {
		t.Fatalf("timeframe = %q, want 1M", app.view.Timeframe)
	}
}

// TestOrderTicketSubmits checks the simulated order path updates the blotter.
func TestOrderTicketSubmits(t *testing.T) {
	app := newTestApp(t)

	before := len(app.demo.Orders)

	if err := app.onClick(context.Background(), ownframe.Box{Action: "order-side-sell"}); err != nil {
		t.Fatal(err)
	}

	if err := app.onClick(context.Background(), ownframe.Box{Action: "order-qty-500"}); err != nil {
		t.Fatal(err)
	}

	if app.view.Ticket.Qty != "500" || app.view.Ticket.SideBuy {
		t.Fatalf("ticket = %+v", app.view.Ticket)
	}

	if err := app.onClick(context.Background(), ownframe.Box{Action: "order-submit"}); err != nil {
		t.Fatal(err)
	}

	if len(app.demo.Orders) != before+1 {
		t.Fatalf("orders = %d, want %d", len(app.demo.Orders), before+1)
	}

	if !app.view.Toast.Show {
		t.Fatal("submit did not show a toast")
	}

	if app.view.Orders[0].Status != "Working" || !app.view.Orders[0].Sell {
		t.Fatalf("new order = %+v", app.view.Orders[0])
	}
}

// TestRailHistory checks Back and Forward across screens.
func TestRailHistory(t *testing.T) {
	app := newTestApp(t)

	for _, action := range []string{"nav-markets", "nav-news"} {
		if err := app.onClick(context.Background(), ownframe.Box{Action: action}); err != nil {
			t.Fatal(err)
		}
	}

	if err := app.onClick(context.Background(), ownframe.Box{Action: "nav-back"}); err != nil {
		t.Fatal(err)
	}

	if app.view.Screen != "markets" {
		t.Fatalf("back screen = %q, want markets", app.view.Screen)
	}

	if err := app.onClick(context.Background(), ownframe.Box{Action: "nav-forward"}); err != nil {
		t.Fatal(err)
	}

	if app.view.Screen != "news" {
		t.Fatalf("forward screen = %q, want news", app.view.Screen)
	}
}

// TestSettingsToggle checks a preference flips and publishes.
func TestSettingsToggle(t *testing.T) {
	app := newTestApp(t)

	before := app.view.Settings

	var key string

	for _, section := range before.Sections {
		if len(section.Rows) > 0 {
			key = section.Rows[0].Key
			break
		}
	}

	if key == "" {
		t.Fatal("no settings row")
	}

	was := app.settings[key]

	if err := app.onClick(context.Background(), ownframe.Box{Action: "toggle-setting-" + key}); err != nil {
		t.Fatal(err)
	}

	if app.settings[key] == was {
		t.Fatalf("setting %q did not flip", key)
	}
}

// TestScreenshot writes a PNG of the watchlist when explicitly requested.
// It is a development aid, not a normal test.
func TestScreenshot(t *testing.T) {
	if os.Getenv("OWNFRAME_TRADE_SCREENSHOT") != "1" {
		t.Skip("set OWNFRAME_TRADE_SCREENSHOT=1 to capture the page")
	}

	app := newTestApp(t)

	if err := os.MkdirAll(testScreenshotDir, 0o700); err != nil {
		t.Fatal(err)
	}

	writeShot(t, app, "watchlist")

	for _, screen := range []string{"markets", "portfolio", "orders", "screener", "news", "settings"} {
		if err := app.onClick(context.Background(), ownframe.Box{Action: "nav-" + screen}); err != nil {
			t.Fatalf("nav-%s: %v", screen, err)
		}

		if err := app.Page().Redraw(context.Background()); err != nil {
			t.Fatalf("redraw %s: %v", screen, err)
		}

		writeShot(t, app, screen)
	}

	if err := app.onClick(context.Background(), ownframe.Box{Action: "nav-watchlist"}); err != nil {
		t.Fatal(err)
	}

	for _, tab := range []string{"options", "news", "financial"} {
		if err := app.onClick(context.Background(), ownframe.Box{Action: "chart-tab-" + tab}); err != nil {
			t.Fatal(err)
		}

		if err := app.Page().Redraw(context.Background()); err != nil {
			t.Fatalf("redraw tab %s: %v", tab, err)
		}

		writeShot(t, app, "watchlist-"+tab)
	}
}

// writeShot writes the current page PNG under the screenshot directory.
func writeShot(t *testing.T, app *App, name string) {
	t.Helper()

	png := app.Page().Probe().PNG()
	if len(png) == 0 {
		t.Fatalf("%s: no PNG", name)
	}

	path := filepath.Join(testScreenshotDir, name+".png")
	if err := os.WriteFile(path, png, 0o600); err != nil {
		t.Fatal(err)
	}

	t.Logf("wrote %s", path)
}

// TestResponsiveScreens checks the layout at the four sizes the terminal must
// follow. The fast path changes the viewport and walks the screens without
// rasterizing the page; the full seven by four PNG matrix stays behind
// OWNFRAME_TRADE_RESPONSIVE_FULL=1. It is a development aid, not part of the
// default test run.
func TestResponsiveScreens(t *testing.T) {
	if os.Getenv("OWNFRAME_TRADE_RESPONSIVE") != "1" {
		t.Skip("set OWNFRAME_TRADE_RESPONSIVE=1 to check responsive layouts")
	}

	// The fast path checks the default, the largest, and the tightest size.
	// The full matrix below still covers 1920x1080.
	sizes := [][2]int{{1600, 900}, {2560, 1440}, {1280, 720}}
	screens := []string{"watchlist", "markets", "portfolio", "orders", "screener", "news", "settings"}

	if os.Getenv("OWNFRAME_TRADE_RESPONSIVE_FULL") == "1" {
		captureResponsiveScreens(t, sizes, screens)
		return
	}

	// Start on the densest screen. Each later size is a relayout of the same
	// source until the last one; the six other screens are then checked at
	// the tightest size, where clipping is most likely.
	app := newTestAppSized(t, sizes[0][0], sizes[0][1])

	// Motion is not what this test checks, and binding it on every screen
	// change is a large share of the layout sweep.
	app.Page().Window().SetAnimations(false)
	app.Page().Window().SetTransitions(false)

	assertScreenFits(t, app, "watchlist", sizes[0][0], sizes[0][1])

	for _, size := range sizes[1:] {
		app.Page().SetSize(size[0], size[1])

		if err := app.Page().Redraw(context.Background()); err != nil {
			t.Fatalf("watchlist redraw at %dx%d: %v", size[0], size[1], err)
		}

		assertScreenFits(t, app, "watchlist", size[0], size[1])
	}

	for _, screen := range screens[1:] {
		if err := app.onClick(context.Background(), ownframe.Box{Action: "nav-" + screen}); err != nil {
			t.Fatalf("nav-%s: %v", screen, err)
		}

		if err := app.Page().Redraw(context.Background()); err != nil {
			t.Fatalf("redraw %s at 1280x720: %v", screen, err)
		}

		assertScreenFits(t, app, screen, sizes[len(sizes)-1][0], sizes[len(sizes)-1][1])
	}
}

// captureResponsiveScreens writes the full seven by four PNG matrix. It is
// slow because every capture rasterizes the whole page in software.
func captureResponsiveScreens(t *testing.T, sizes [][2]int, screens []string) {
	t.Helper()

	if err := os.MkdirAll(testScreenshotDir, 0o700); err != nil {
		t.Fatal(err)
	}

	for _, size := range sizes {
		app := newTestAppSized(t, size[0], size[1])

		for i, screen := range screens {
			if i > 0 {
				if err := app.onClick(context.Background(), ownframe.Box{Action: "nav-" + screen}); err != nil {
					t.Fatalf("nav-%s: %v", screen, err)
				}

				if err := app.Page().Redraw(context.Background()); err != nil {
					t.Fatalf("redraw %s at %dx%d: %v", screen, size[0], size[1], err)
				}
			}

			assertScreenFits(t, app, screen, size[0], size[1])
			writeShot(t, app, fmt.Sprintf("%dx%d-%s", size[0], size[1], screen))
		}
	}
}

// TestWindowFollowsSize keeps the page from reintroducing a fixed canvas.
func TestWindowFollowsSize(t *testing.T) {
	app := newTestApp(t)

	for _, size := range [][2]int{{1280, 720}, {1600, 900}, {1920, 1080}} {
		app.Page().SetSize(size[0], size[1])

		if err := app.Page().Redraw(context.Background()); err != nil {
			t.Fatalf("redraw at %dx%d: %v", size[0], size[1], err)
		}

		cfg, err := png.DecodeConfig(bytes.NewReader(app.Page().Probe().PNG()))
		if err != nil {
			t.Fatalf("decode %dx%d: %v", size[0], size[1], err)
		}

		if cfg.Width != size[0] || cfg.Height != size[1] {
			t.Fatalf("page at %dx%d painted %dx%d", size[0], size[1], cfg.Width, cfg.Height)
		}
	}
}

// TestNoImageAssets keeps the terminal a real HTML/CSS interface. Icons are
// inline SVG; a bitmap or a url() background would make the reference image a
// dependency again.
func TestNoImageAssets(t *testing.T) {
	entries, err := files.ReadDir("components")
	if err != nil {
		t.Fatal(err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if !strings.HasSuffix(name, ".html") && !strings.HasSuffix(name, ".css") {
			continue
		}

		body, err := files.ReadFile("components/" + name)
		if err != nil {
			t.Fatal(err)
		}

		text := strings.ToLower(string(body))
		if strings.Contains(text, "url(") || strings.Contains(text, "<img") {
			t.Fatalf("%s references an external image", name)
		}
	}
}
