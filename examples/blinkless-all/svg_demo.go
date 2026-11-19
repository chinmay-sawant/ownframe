package main

import (
	"context"
	"fmt"
	"html"

	"github.com/chinmay-sawant/blinkless/layout"
)

// svgShow draws an HTML box that carries the property, then an inline svg
// that carries it on a real SVG element, then the value the engine exported
// for the box.
func svgShow(id, prop, boxStyle, svgInner, caption string) string {
	markup := fmt.Sprintf(`<div class="cell"><div id="%s" style="%s;font-size:13px;padding:4px 8px;background:#f1efe7">%s</div></div>`+
		`<div class="cell"><svg width="300" height="90" viewBox="0 0 300 90" xmlns="http://www.w3.org/2000/svg">%s</svg>`+
		`<div class="cap">%s</div></div>`, id, boxStyle, html.EscapeString(boxStyle), svgInner, html.EscapeString(caption))

	d, err := shapesLayout(context.Background(), markup)
	if err != nil {
		return markup + "<p>layout failed: " + html.EscapeString(err.Error()) + "</p>"
	}

	return markup + shapesRes("%s: %q on the box, exported to the host", prop, declaredByID(d)[id][prop])
}

func demoSvgTextRendering() string {
	return svgShow("sv-tr", "text-rendering", "text-rendering:optimizeLegibility",
		`<text x="6" y="28" font-size="22">Wave auto</text>`+
			`<text x="6" y="56" font-size="22" style="text-rendering:optimizeSpeed">Wave speed</text>`+
			`<text x="6" y="84" font-size="22" style="text-rendering:geometricPrecision">Wave geometric</text>`,
		"three lines: auto, optimizeSpeed, geometricPrecision. All draw the same unhinted outlines")
}

func demoSvgLighting() string {
	filter := func(id, color string) string {
		return `<filter id="` + id + `"><feDiffuseLighting in="SourceGraphic" lighting-color="` + color +
			`" surfaceScale="5"><feDistantLight azimuth="45" elevation="55"/></feDiffuseLighting></filter>`
	}

	return svgShow("sv-lc", "lighting-color", "lighting-color:#336699",
		filter("lfw", "white")+filter("lfb", "#336699")+
			`<rect x="10" y="10" width="60" height="60" fill="#aaa" filter="url(#lfw)"/>`+
			`<rect x="130" y="10" width="60" height="60" fill="#aaa" filter="url(#lfb)"/>`,
		"the same rect lit by a distant light (55 deg): white light (paints light grey), then lighting-color #336699")
}

func demoSvgGlyph() string {
	return svgShow("sv-go", "glyph-orientation-vertical", "glyph-orientation-vertical:90",
		`<text x="6" y="30" font-size="22">Ab12 plain</text>`+
			`<text x="6" y="64" font-size="22" style="glyph-orientation-vertical:90">Ab12 90deg</text>`,
		"horizontal text with and without the property: it draws the same, the property has no effect")
}

func demoSvgBaseline() string {
	return svgShow("sv-bs", "baseline-shift", "baseline-shift:super",
		`<text x="6" y="60" font-size="26">x<tspan>N</tspan> x<tspan style="baseline-shift:super">sup</tspan> `+
			`x<tspan style="baseline-shift:sub">sub</tspan> x<tspan style="baseline-shift:15px">15</tspan></text>`,
		"tspans raised and lowered: none, super, sub, 15px")
}

func demoSvgMarker() string {
	marker := `<marker id="mk" markerWidth="12" markerHeight="12" refX="6" refY="6" orient="auto">` +
		`<path d="M0 0 L12 6 L0 12 Z" fill="#2563eb"/></marker>`

	return svgShow("sv-ms", "marker-side", "marker-side:match-parent",
		`<defs>`+marker+`</defs>`+
			`<path d="M10 25 L120 25 L150 60" stroke="#c0392b" stroke-width="3" fill="none" marker-end="url(#mk)"/>`+
			`<path d="M160 25 L270 25 L290 60" stroke="#c0392b" stroke-width="3" fill="none" marker-end="url(#mk)" style="marker-side:match-parent"/>`,
		"the same path without and with match-parent: the markers draw the same")
}

// declaredByID keys the Declared maps of a display by element id.
func declaredByID(d *layout.Display) map[string]map[string]string {
	out := map[string]map[string]string{}

	for _, box := range d.Boxes {
		if box.ID != "" && box.Declared != nil {
			out[box.ID] = box.Declared
		}
	}

	return out
}
