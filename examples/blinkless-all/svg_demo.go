package main

import (
	"context"
	"fmt"
	"html"

	"github.com/chinmay-sawant/blinkless/layout"
)

// svgShow draws an HTML box that carries the property, then an inline svg
// that carries it on a real SVG element, then the value the engine exported
// for the box. The rasterizer ignores all five properties, so the readout
// says what reaches the serialized SVG instead of claiming a pixel change.
func svgShow(id, prop, boxStyle, svgInner, caption string) string {
	markup := fmt.Sprintf(`<div class="cell"><div id="%s" style="%s;font-size:13px;padding:4px 8px;background:#f1efe7">%s</div></div>`+
		`<div class="cell"><svg width="260" height="60" viewBox="0 0 260 60" xmlns="http://www.w3.org/2000/svg">%s</svg>`+
		`<div class="cap">%s</div></div>`, id, boxStyle, html.EscapeString(boxStyle), svgInner, html.EscapeString(caption))

	d, err := shapesLayout(context.Background(), markup)
	if err != nil {
		return markup + "<p>layout failed: " + html.EscapeString(err.Error()) + "</p>"
	}

	return markup + shapesRes("%s: %q on the box; baked onto the svg element, ignored by the rasterizer", prop, declaredByID(d)[id][prop])
}

func demoSvgTextRendering() string {
	return svgShow("sv-tr", "text-rendering", "text-rendering:optimizeLegibility",
		`<text x="6" y="30" font-size="16" style="text-rendering:optimizeSpeed">SVG text</text>`, "svg text with optimizeSpeed (the rasterizer draws no SVG text)")
}

func demoSvgLighting() string {
	return svgShow("sv-lc", "lighting-color", "lighting-color:#336699",
		`<filter id="lf"><feDiffuseLighting style="lighting-color:#336699"/></filter><rect x="10" y="10" width="60" height="36" fill="#9db4d6" filter="url(#lf)"/>`,
		"rect with a lighting filter, drawn without the filter")
}

func demoSvgGlyph() string {
	return svgShow("sv-go", "glyph-orientation-vertical", "glyph-orientation-vertical:90",
		`<text x="6" y="30" font-size="16" style="glyph-orientation-vertical:90">SVG text</text>`, "svg text with 90deg (the rasterizer draws no SVG text)")
}

func demoSvgBaseline() string {
	return svgShow("sv-bs", "baseline-shift", "baseline-shift:super",
		`<text x="6" y="30" font-size="16">base<tspan style="baseline-shift:super">raised</tspan></text>`, "tspan with baseline-shift: super (no SVG text is drawn)")
}

func demoSvgMarker() string {
	return svgShow("sv-ms", "marker-side", "marker-side:match-parent",
		`<path d="M10 30 L200 30" stroke="#c0392b" stroke-width="3" fill="none" style="marker-side:match-parent"/>`, "path with match-parent")
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
