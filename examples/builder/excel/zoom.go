package main

import (
	"math"
	"regexp"
	"strconv"
	"strings"
)

// The engine has no CSS zoom property, so the UI scale is real layout: every
// px length in the theme sheet is multiplied by the zoom percentage, and the
// Go geometry (bands, grid cells, popovers) goes through zs with the same
// rounding, so painted boxes and hit testing agree at every level.
const (
	zoomStepPct = 10 // one chord or wheel notch

	// Default scale when the user has not chosen one, by window width.
	zoomWideWidth  = 2400
	zoomWidePct    = 150
	zoomLargeWidth = 1800
	zoomLargePct   = 125

	cssDecimals = 100 // fractional px lengths keep two decimals
)

// autoZoomPct is the default zoom for a window of the given width: a wide
// monitor gets a larger UI so the chrome stays readable.
func autoZoomPct(width int) int {
	switch {
	case width >= zoomWideWidth:
		return zoomWidePct
	case width >= zoomLargeWidth:
		return zoomLargePct
	}
	return zoomDefaultPct
}

func clampZoom(pct int) int {
	return min(max(pct, zoomMinPct), zoomMaxPct)
}

// scaleBy scales n by pct percent, rounding half up and keeping at least 1.
func scaleBy(n, pct int) int {
	if pct == zoomBase || n <= 0 {
		return n
	}
	return max((n*pct+zoomBase/2)/zoomBase, 1)
}

// setZoom stores an explicit zoom, which ends the width based default. It
// reports whether the level changed.
func (a *app) setZoom(pct int) bool {
	pct = clampZoom(pct)
	if pct == a.zoomPct() && a.x.zoom != 0 {
		return false
	}
	a.x.zoom = pct
	a.redraw()
	return true
}

// stepZoom moves the zoom by delta percent.
func (a *app) stepZoom(delta int) {
	a.setZoom(a.zoomPct() + delta)
}

// zoomKey runs a Ctrl zoom chord. Plus is usually Shift+Equal, so both forms
// and the keypad keys count.
func (a *app) zoomKey(key string) bool {
	switch key {
	case "ctrl+=", "ctrl++", "ctrl+equal", "ctrl+plus", "ctrl+kpadd",
		"ctrl+shift+=", "ctrl+shift+equal", "ctrl+shift+plus":
		a.stepZoom(zoomStepPct)
	case "ctrl+-", "ctrl+minus", "ctrl+kpsubtract", "ctrl+shift+minus":
		a.stepZoom(-zoomStepPct)
	case "ctrl+0", "ctrl+kp0":
		a.setZoom(zoomDefaultPct)
	default:
		return false
	}
	return true
}

// pxLength matches a CSS px length, with an optional fraction.
var pxLength = regexp.MustCompile(`(\d+\.\d+|\d+|\.\d+)px`)

// scaleCSS multiplies every px length in a stylesheet by pct percent. A 1px
// hairline stays 1px, and @media lines keep their breakpoints.
func scaleCSS(css string, pct int) string {
	if pct == zoomBase {
		return css
	}
	lines := strings.Split(css, "\n")
	for i, ln := range lines {
		if strings.Contains(ln, "@media") || !strings.Contains(ln, "px") {
			continue
		}
		lines[i] = pxLength.ReplaceAllStringFunc(ln, func(m string) string {
			v, err := strconv.ParseFloat(strings.TrimSuffix(m, "px"), 64)
			if err != nil {
				return m
			}
			return scaleLength(v, pct) + "px"
		})
	}
	return strings.Join(lines, "\n")
}

func scaleLength(v float64, pct int) string {
	if v == 1 {
		return "1"
	}
	if v == math.Trunc(v) {
		return strconv.Itoa(scaleBy(int(v), pct))
	}
	scaled := math.Round(v*float64(pct)/zoomBase*cssDecimals) / cssDecimals
	return strconv.FormatFloat(scaled, 'f', -1, 64)
}

// themeSheet is the active theme at the active zoom. It records the zoom it
// was built for, so syncTheme can tell when the page needs a new sheet.
func (a *app) themeSheet() string {
	pct := a.zoomPct()
	a.x.themePct = pct
	if a.x.themeCSS != "" && a.x.themeDark == a.dark && a.x.themeCachePct == pct {
		return a.x.themeCSS
	}
	base := ExcelLight
	if a.dark {
		base = ExcelDark
	}
	css := scaleCSS(base, pct)
	a.x.themeCSS, a.x.themeDark, a.x.themeCachePct = css, a.dark, pct
	return css
}

// syncTheme reloads the page theme when the zoom, set by the user or chosen
// from the window width, differs from the one the page was built with.
func (a *app) syncTheme() {
	if a.page == nil || a.zoomPct() == a.x.themePct {
		return
	}
	if err := a.page.SetTheme(a.themeSheet()); err != nil {
		a.lastErr = err
	}
}
