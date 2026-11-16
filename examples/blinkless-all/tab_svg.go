package main

// svgTab is the "SVG leftovers" tab: 5 properties in 5 examples.
func svgTab() Tab {
	return Tab{ID: "svg", Title: "SVG leftovers", Props: []Prop{
		{Name: "text-rendering",
			Note: "The engine writes the value onto the SVG text element; the rasterizer drops SVG text, so no pixel changes.",
			Demo: demoSvgTextRendering},
		{Name: "lighting-color",
			Note: "The engine writes the color onto the lighting filter primitive; no filter graph runs, so no pixel changes.",
			Demo: demoSvgLighting},
		{Name: "glyph-orientation-vertical",
			Note: "The engine writes 90deg or 0deg onto the SVG text element; the rasterizer ignores it.",
			Demo: demoSvgGlyph},
		{Name: "baseline-shift",
			Note: "The engine writes the shift onto the tspan; the rasterizer ignores it.",
			Demo: demoSvgBaseline},
		{Name: "marker-side",
			Note: "The engine writes the keyword onto the shape; the rasterizer ignores it.",
			Demo: demoSvgMarker},
	}, Head: shapesHead}
}
