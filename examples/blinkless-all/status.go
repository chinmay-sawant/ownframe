package main

// painterDeclines lists the properties the engine parses and lays out but the
// window painter does not draw, so the page shows the row as partial: the
// catalog says implemented for the engine, and the picture cannot show it. A
// demo whose picture is identical to its control belongs here, not under an
// implemented badge.
var painterDeclines = map[string]bool{
	"visibility": true, "gap": true, "filter": true, "scale": true, "backdrop-filter": true,
	"color-scheme": true, "transform-box": true, "grid-auto-columns": true, "grid-auto-flow": true,
	"writing-mode": true, "position-area": true, "outline-color": true, "clip-rule": true,
	"border-image": true, "border-image-source": true, "border-image-slice": true, "border-image-width": true,
	"mask-border-mode": true, "mask-border-outset": true, "mask-border-repeat": true, "mask-border-slice": true,
	"mask-border-width": true, "grid-row-start": true, "grid-row-end": true, "grid-template-areas": true,
	"marker": true, "marker-start": true, "marker-mid": true, "marker-end": true, "paint-order": true,
	"fill-rule": true, "stroke-linejoin": true, "stroke-miterlimit": true,
	"-webkit-mask-box-image-outset": true, "-webkit-mask-box-image-repeat": true,
	"-webkit-mask-box-image-slice": true, "-webkit-mask-box-image-width": true,
	"border-bottom-color": true, "border-left-color": true, "border-right-color": true, "border-top-color": true,
	"border-image-outset": true, "border-image-repeat": true,
	"rule-break": true,
	"cx":         true, "cy": true, "r": true, "x": true, "y": true,
	"color-interpolation": true, "color-interpolation-filters": true, "stop-color": true, "stop-opacity": true,
	"font-variant-emoji": true, "tab-size": true, "text-emphasis-style": true, "text-overflow": true,
}

// opOnly lists the properties that change the engine's operation list while
// the window picture stays the same as the control, so the readout says
// "differs" for a picture that does not.
var opOnly = map[string]bool{
	"font-language-override": true, "font-feature-settings": true, "font-synthesis": true,
	"font-synthesis-weight": true, "font-variant-alternates": true, "font-variant-east-asian": true,
	"font-variant-ligatures": true, "font-variant-numeric": true, "text-anchor": true, "text-autospace": true,
	"background-blend-mode": true, "image-rendering": true, "object-fit": true, "object-view-box": true,
	"isolation": true, "rule-overlap": true,
}

// rowStatus is the badge of a row: the catalog status, lowered to partial when
// the window painter does not draw the property.
func (a *App) rowStatus(name string) string {
	st, _ := a.cat.Status(name)
	if st == StatusImplemented && declines(name) {
		return StatusPartial
	}

	return st
}

// declines reports whether the window painter leaves the property out.
func declines(name string) bool {
	return noPaint[name] || painterDeclines[name] || painterDeclines[trimWebkit(name)]
}

func trimWebkit(name string) string {
	const prefix = "-webkit-"
	if len(name) > len(prefix) && name[:len(prefix)] == prefix {
		return name[len(prefix):]
	}

	return name
}
