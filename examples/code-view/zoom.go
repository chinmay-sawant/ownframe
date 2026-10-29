package main

import (
	"context"
	"fmt"
	"math"
	"strings"
)

// Layout zoom bounds for the code view. The window's own paint zoom covers
// 0.25 to 4; the layout zoom stays inside it and matches the editor feel of
// staying readable without exploding the chrome.
const (
	zoomMin  = 0.5
	zoomMax  = 3.0
	zoomStep = 1.1
	// zoomEpsilon treats a level within a rounding error of 1 as exactly
	// 100 percent, so the override stays byte-stable at the default.
	zoomEpsilon = 1e-9
	// zoomPercent scales a level to the status bar label.
	zoomPercent = 100
)

// uiZoom is the active layout multiplier, defaulting to 1 for a fresh app
// and clamping anything the session hands back into range.
func (a *App) uiZoom() float64 {
	if a.zoom <= 0 {
		return 1
	}

	if a.zoom < zoomMin {
		return zoomMin
	}

	if a.zoom > zoomMax {
		return zoomMax
	}

	return a.zoom
}

// zx scales an integer pixel value by the layout zoom, keeping at least one
// pixel. Pass every layout constant that feeds scroll math, hit testing, or
// a CSS var through here so Go and paint stay on the same grid.
func (a *App) zx(v int) int {
	if v <= 0 {
		return v
	}

	return max(1, int(math.Round(float64(v)*a.uiZoom())))
}

// zf scales a fractional pixel value such as a monospace cell advance.
func (a *App) zf(v float64) float64 {
	return v * a.uiZoom()
}

// zoomLabel renders the status bar chip, e.g. "110%".
func (a *App) zoomLabel() string {
	return fmt.Sprintf("%.0f%%", math.Round(a.uiZoom()*zoomPercent))
}

// zoomIn steps the layout zoom up and keeps the scroll position.
func (a *App) zoomIn(ctx context.Context) error {
	return a.setZoom(ctx, a.uiZoom()*zoomStep)
}

// zoomOut steps the layout zoom down and keeps the scroll position.
func (a *App) zoomOut(ctx context.Context) error {
	return a.setZoom(ctx, a.uiZoom()/zoomStep)
}

// resetZoom returns the layout to 100 percent.
func (a *App) resetZoom(ctx context.Context) error {
	return a.setZoom(ctx, 1)
}

// setZoom clamps the level, rescales the scroll offsets so the same first
// row stays on screen, persists the session, and repaints with the zoom
// stylesheet layered over the theme.
func (a *App) setZoom(ctx context.Context, level float64) error {
	old := a.uiZoom()

	a.zoom = level
	now := a.uiZoom()

	if old > 0 && now != old {
		a.scrollY = int(math.Round(float64(a.scrollY) * now / old))
		a.scrollX = int(math.Round(float64(a.scrollX) * now / old))
	}

	a.persistSession()
	a.setScrollX(a.scrollX)
	a.scrollTo(a.scrollY)

	if err := a.page.SetTheme(a.activeThemeCSS()); err != nil {
		return err
	}

	return a.page.Redraw(ctx)
}

// activeThemeCSS layers the zoom override after the theme so it wins the
// cascade without touching the theme catalog.
func (a *App) activeThemeCSS() string {
	return themeCSS(a.theme) + zoomOverrideCSS(a.uiZoom())
}

// zoomFont is one fixed font size the zoom override rescales.
type zoomFont struct {
	sel  string
	base float64
}

// zoomFontTable covers the fixed text the layout zoom rescales. Everything
// else inherits the scaled body size.
func zoomFontTable() []zoomFont {
	return []zoomFont{
		{"body", 12},
		{".cline", 12},
		{".ln", 10},
		{".ln.cont", 9},
		{".tab-name", 9.5},
		{".crumb-path, .diff-path, .file-path", 9.5},
		{".tname", 11},
		{".hist-subject", 9.5},
		{".hist-branch", 9},
		{".history-current", 9.5},
		{".history-current-count", 8.5},
		{".hist-kicker, .hist-pages", 8.5},
		{".hist-sample, .hist-file-stat", 8},
		{".hist-more, .hist-error", 8.5},
		{".history-current-arrow", 15},
		{".drow", 12},
		{".palette-close", 14},
		{".dnum", 9.5},
		{".sb-branch", 9.5},
		{".sb-path", 9},
		{".act-badge", 8.5},
		// Titlebar.
		{".file-btn", 12},
		{".brand-name", 12.5},
		{".brand-sub", 9.5},
		{".cmdbar", 10.5},
		{".cmd-key", 9},
		{".file-item, .project-input", 11},
		{".locations-label", 9},
		{".browse-path", 10},
		{".file-alert", 11},
		{".file-alert-x", 10},
		// Explorer tree and filter.
		{".filter-input", 10.5},
		{".filter-x", 13},
		{".tree-note", 7},
		{".fstat, .cstat", 9},
		{".tree-tag, .tree-more, .side-foot", 9.5},
		{".side-count", 9},
		// Changes list.
		{".changes-title", 9.5},
		{".side-title", 9.5},
		{".change-name", 10.5},
		{".change-dir", 9},
		{".change-num", 8.5},
		{".sc-stats", 9},
		{".changes-count", 8.5},
		{".side-tag", 7.5},
		// Editor chrome.
		{".tab-x", 13},
		{".tab-tool", 11},
		{".seg-btn", 9},
		{".crumb-btn", 9},
		{".crumb-meta", 9.5},
		{".crumbs", 10},
		{".lang-pill", 8.5},
		{".edit-flag", 8},
		{".find-input", 12},
		{".find-count", 8.5},
		{".edit-hint", 9},
		{".expand-btn", 10},
		{".keys em", 8.5},
		{".welcome-card h1", 21},
		{".welcome-card p", 11},
		{".empty h2", 15},
		{".empty p", 10.5},
		{".empty .button", 10},
		{".ctx-item", 11},
		{".keys", 9},
		// Diff.
		{".dhunk, .dmeta", 9.5},
		// History rail.
		{".hist-hash", 9},
		{".hist-when, .hist-stat", 8.5},
		{".hist-meta", 8.5},
		{".hist-file-name", 9.5},
		{".hist-file-dir", 8.5},
		{".hist-chev", 9},
		{".branch-name", 9},
		{".branch-input", 9.5},
		{".branch-mark", 8},
		{".branch-track", 8},
		{".branch-tag", 7.5},
		{".branch-more", 8.5},
		{".hist-btn", 8.5},
		{".commit-tip-float b", 10},
		{".commit-tip-float span", 9.5},
		{".commit-tip-float em", 8},
		// Search palette.
		{".palette-text b", 10.5},
		{".palette-text span", 9},
		{".palette-go", 10},
		{".search-meta", 9},
		{".search-loc", 8.5},
		{".search-line, .search-text", 9},
		{".palette-foot", 9},
		{".palette-empty", 10},
		// Details panel.
		{".settings-header", 14},
		{".settings-title", 14},
		{".settings-theme-text span", 10},
		{".settings-check", 16},
		{".det-kicker", 9.5},
		{".det-x", 14},
		{".det-file b", 11},
		{".det-file span", 9},
		{".det-grid", 9.5},
		{".okind", 8},
		{".oname", 9},
		{".olno", 8.5},
		{".det-empty", 9.5},
		{".det-sub", 8.5},
		{".ficon", 8.5},
		{".ficon.lg", 11},
		// Status bar.
		{".statusbar", 10},
		{".scan-badge", 7.5},
		// Preview.
		{".prow", 12},
		{".pmark", 10},
		{".prow-h1 .ptext", 14},
		{".prow-h2 .ptext", 13},
		{".prow-h3 .ptext", 12},
		{".prow-code .ptext, .prow-table .ptext, .prow-diagram .ptext", 11},
		// Toast and bottom floats.
		{".toast-text b", 11.5},
		{".toast-text span", 10.5},
		{".toast-x", 14},
		{".file-nav-arrow", 17},
		{".file-nav-change", 14},
		{".file-nav-pos, .file-nav-stats", 9.5},
		{".file-nav-current", 11},
		{".file-nav-dir", 9},
	}
}

// zoomOverrideCSS rescales the layout vars, the fixed row heights, and the
// reader fonts. An empty string at 100 percent keeps the theme byte-stable.
func zoomOverrideCSS(z float64) string {
	if z <= 0 {
		z = 1
	}

	if math.Abs(z-1) < zoomEpsilon {
		return ""
	}

	px := func(v int) string {
		return fmt.Sprintf("%dpx", max(1, int(math.Round(float64(v)*z))))
	}

	var b strings.Builder
	b.WriteString(":root {")
	fmt.Fprintf(&b, " --row-h:%s !important;", px(rowH))
	fmt.Fprintf(&b, " --titlebar-h:%s !important;", px(chromeTop))
	fmt.Fprintf(&b, " --tab-h:%s !important;", px(tabStripH))
	fmt.Fprintf(&b, " --crumb-h:%s !important;", px(crumbH))
	fmt.Fprintf(&b, " --side-w:%s !important;", px(sideWidth))
	fmt.Fprintf(&b, " --activity-w:%s !important;", px(activityW))
	fmt.Fprintf(&b, " --status-h:%s !important;", px(chromeBottom))
	b.WriteString(" }")

	// Core rows.
	fmt.Fprintf(&b, ".trow { height:%s !important; }", px(treeRowH))

	histH := px(histRowH)
	fmt.Fprintf(&b, ".hist-row { height:%s !important; min-height:%s !important; max-height:%s !important; }", histH, histH, histH)
	fmt.Fprintf(&b, ".change { min-height:%s !important; }", px(changeRowH))
	fmt.Fprintf(&b, ".orow { height:%s !important; }", px(treeRowH))
	fmt.Fprintf(&b, ".edit-selection { height:%s !important; }", px(rowH))
	fmt.Fprintf(&b, ".hsplit { height:%s !important; flex-basis:%s !important; }", px(histSplitH), px(histSplitH))

	for _, f := range zoomFontTable() {
		fmt.Fprintf(&b, "%s { font-size:%.1fpx !important; }", f.sel, f.base*z)
	}

	zoomChromeCSS(&b, px)
	zoomTreeCSS(&b, px)
	zoomEditorCSS(&b, px)
	zoomRailsCSS(&b, px)
	zoomHistoryCSS(&b, px)
	zoomDiffCSS(&b, px)
	zoomFloatsCSS(&b, px)
	zoomOverlayCSS(&b, px)
	zoomDetailsCSS(&b, px)
	zoomPreviewCSS(&b, px)
	zoomActivityCSS(&b, px)

	return b.String()
}
