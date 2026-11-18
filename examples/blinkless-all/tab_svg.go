package main

// svgTab is the "SVG leftovers" tab: 5 properties in 5 examples.
func svgTab() Tab {
	return Tab{ID: "svg", Title: "SVG leftovers", Props: []Prop{
		{Name: "text-rendering",
			Note: "SVG text now draws as outlines from the bundled fonts. The three modes draw the same unhinted glyphs, so the property changes no pixels.",
			Demo: demoSvgTextRendering},
		{Name: "lighting-color",
			Note: "A diffuse light with a distant light shades the shape in lighting-color: a flat interior, with a darker edge band on the side away from the light.",
			Demo: demoSvgLighting},
		{Name: "glyph-orientation-vertical",
			Note: "Accepted and exported. It has no rendering effect on horizontal text, so both lines draw the same.",
			Demo: demoSvgGlyph},
		{Name: "baseline-shift",
			Note: "A tspan moves up or down: super and sub by half the font's height, a length as written, a percentage of the font size.",
			Demo: demoSvgBaseline},
		{Name: "marker-side",
			Note: "Accepted and exported. Markers draw the same with either keyword.",
			Demo: demoSvgMarker},
	}, Head: shapesHead}
}
