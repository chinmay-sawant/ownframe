package main

// readTab is the "Reading order and misc" tab: 5 properties in 5 examples.
func readTab() Tab {
	return Tab{ID: "read", Title: "Reading order and misc", Props: []Prop{
		{Name: "reading-flow", Demo: readDemoFlow,
			Note: "The host reads the children of a container in visual, row or column order instead of source order."},
		{Name: "reading-order", Demo: readDemoOrder,
			Note: "Inside a reordered container the host reads the children by their number, lowest first."},
		{Name: "overlay", Demo: readDemoOverlay,
			Note: "A closing element with overlay auto stays in the top layer until its exit ends; none removes it at once."},
		{Name: "background-attachment", Demo: readDemoAttach,
			Note: "The engine paints every layer as scroll; the host computes the shift a local or fixed layer would need."},
		{Name: "background-tbd", Demo: readDemoTbd,
			Note: "Not a CSS property with a defined meaning: the engine keeps the text and the host shows it."},
	}, Head: func() string { return readHeadCSS }}
}
