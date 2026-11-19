package main

import "strings"

// layoutSpec returns the demo of a layout or misc property, or false when
// another group file answers for it.
func layoutSpec(r genProp) (DemoSpec, bool) {
	if r.Group != "layout" && r.Group != "misc" {
		return DemoSpec{}, false
	}

	n := r.Name

	switch {
	case in(n, "align-content", "place-content"):
		return lay(flexWrapDemo("flex-wrap:wrap;{{STYLE}}height:100px;"), "The two wrapped lines of boxes move from the top of the container to its centre.", true), true
	case n == "align-items":
		return lay(flexDemo("{{STYLE}}", "", "", ""), "The three boxes of different heights centre on the cross axis instead of stretching or sitting at the top.", true), true
	case n == "justify-content":
		return lay(flexDemo("{{STYLE}}", "", "", ""), "The three boxes bunch in the middle of the row instead of the left edge.", true), true
	case n == "align-self":
		return lay(flexDemo("", "{{STYLE}}", "", ""), "The first box drops to the bottom of the container while the other two stay at the top.", true), true
	case n == "order":
		return lay(flexDemo("", "{{STYLE}}", "", ""), "The first (orange) box moves from the left end to after the other two.", true), true
	case n == "flex" || n == "flex-grow":
		return lay(flexDemo("", "{{STYLE}}", "", ""), "The first box stretches to take the free width of the row.", true), true
	case n == "flex-basis":
		return lay(flexDemo("", "width:auto;min-width:24px;{{STYLE}}", "", ""), "The first box starts from half the container width instead of its content width.", true), true
	case n == "flex-shrink":
		return lay(shrinkDemo(), "The first box gives up more of its width when the three boxes overflow the row.", true), true
	case n == "flex-direction":
		return lay(flexDemo("{{STYLE}}", "", "", ""), "The row runs right to left, packed against the right edge: the orange box is first on the right and the blue box last on the left of the group, leaving free space on the left.", true), true
	case n == "flex-wrap":
		return lay(flexWrapDemo("{{STYLE}}"), "The third box wraps onto a second line instead of overflowing the row.", true), true
	case n == "flex-flow":
		return lay(`<div id="{{ID}}" style="display:flex;width:280px;height:60px;background:#e8e4d8;{{STYLE}}">`+
			kid(1, "width:50px;height:30px;background:#c96;")+kid(2, "width:50px;height:30px;background:#6a9;")+kid(3, "width:50px;height:30px;background:#69c;")+`</div>`,
			"The boxes stack in a column and wrap into a second column when the container height runs out.", true), true
	case n == "gap":
		return lay(gridDemo("grid-template-columns:1fr 1fr;{{STYLE}}", "", ""), "Accepted; the engine does not apply the two-value gap shorthand, so the grid keeps its spacing (row-gap and column-gap are shown on their own rows).", false), true
	case n == "row-gap":
		return lay(flexWrapDemo("flex-wrap:wrap;{{STYLE}}"), "Space opens between the two wrapped lines of boxes.", true), true
	case in(n, "justify-items", "justify-self", "place-items", "place-self"):
		cont, kidStyle, look := "", "", "The 40px boxes centre inside their wider grid cells instead of sitting at the cell's start edge."
		if strings.HasSuffix(n, "self") {
			kidStyle = "{{STYLE}}"
			look = "Only the first 40px box (orange, top left) centres inside its wider grid cell; the other boxes keep the cell's start edge."
		} else {
			cont = "{{STYLE}}"
		}

		return lay(gridDemo("grid-template-columns:1fr 1fr;"+cont, kidStyle, "width:40px;"), look, true), true

	case n == "grid" || n == "grid-template-columns":
		return lay(gridDemo("{{STYLE}}", "", ""), "The six boxes lay out in columns instead of one stacked column.", true), true
	case n == "grid-template-rows":
		return lay(gridDemo("{{STYLE}}", "", ""), "The first grid row becomes 30px and the second 60px instead of the 16px of their boxes; the boxes keep their 16px and sit at the top of their rows, and later rows stay 16px.", true), true
	case n == "grid-template":
		return lay(gridDemo("{{STYLE}}", "", ""), "The first grid row becomes the declared 40px tall instead of the 16px of its box; the box keeps its 16px and sits at the top of the row, and the later rows stay 16px.", true), true
	case n == "grid-auto-rows":
		return lay(gridDemo("{{STYLE}}", "", ""), "Every implicit row is 40px tall.", true), true
	case n == "grid-auto-columns":
		return lay(gridDemo("grid-auto-flow:column;{{STYLE}}", "", ""), "Accepted; the engine lays implicit columns out the same, so the widths do not change.", false), true
	case n == "grid-auto-flow":
		return lay(gridDemo("grid-template-rows:20px 20px;{{STYLE}}", "", ""), "Accepted; the engine keeps filling row by row, so the boxes do not move.", false), true
	case n == "grid-template-areas":
		return lay(gridDemo("{{STYLE}}", "", ""), "The container declares a one-cell named area, but the boxes name no area, so they keep the grid's layout; nothing visible moves.", false), true
	case n == "grid-row-start" || n == "grid-row-end":
		return lay(gridDemo("grid-template-columns:1fr 1fr 1fr;", "{{STYLE}}", ""), "Accepted; the engine places the first box in the same cell, so nothing moves (grid-row shows the placement).", false), true
	case n == "grid-row":
		return lay(gridDemo("grid-template-columns:1fr 1fr 1fr;", "{{STYLE}}", ""), "The first box occupies rows 1 and 2 of the first column, so its 16px box leaves an empty cell under it, box 4 moves over and box 6 drops to a third row.", true), true
	case n == "grid-column-end":
		return lay(gridDemo("grid-template-columns:1fr 1fr 1fr;", "{{STYLE}}", ""), "The first box ends at column line 3 with no start line, so it takes column 2 only and column 1 of the first row stays empty; box 2 moves to column 3 and the later boxes follow on the next rows.", true), true
	case in(n, "grid-area", "grid-column", "grid-column-start"):
		return lay(gridDemo("grid-template-columns:1fr 1fr 1fr;", "{{STYLE}}", ""), "The first box moves to the placed column line or spans columns, and the later boxes shift to make room.", true), true

	case in(n, "margin", "margin-left", "margin-right", "margin-top", "margin-bottom", "margin-block", "margin-block-start", "margin-block-end", "margin-inline", "margin-inline-start", "margin-inline-end"):
		return lay(blockDemo("height:30px;background:#c96;{{STYLE}}", ""), marginLook(n), true), true
	case in(n, "padding", "padding-left", "padding-right", "padding-top", "padding-bottom", "padding-block", "padding-block-start", "padding-block-end", "padding-inline", "padding-inline-start", "padding-inline-end"):
		return lay(`<div id="{{ID}}" style="width:200px;background:#c96;{{STYLE}}"><div id="{{ID}}-1" style="height:20px;background:#69c;">content</div></div>`,
			"The orange box grows on the padded side and the blue content moves away from that edge.", true), true
	case in(n, "width", "inline-size"):
		return lay(blockDemo("height:30px;background:#c96;{{STYLE}}", ""), "The orange bar shrinks from the full frame width to 12px.", true), true
	case n == "height":
		return lay(blockDemo("height:40px;width:150px;background:#c96;{{STYLE}}", ""), "The orange box drops from 40px tall to 12px.", true), true
	case n == "block-size":
		return lay(`<div id="{{ID}}" style="height:80px;width:280px;background:#e8e4d8;"><div id="{{ID}}-1" style="width:150px;background:#c96;{{STYLE}}">box</div></div>`,
			"The orange box becomes half the 80px frame height instead of one text line.", true), true
	case n == "min-width":
		return lay(blockDemo("height:30px;width:6px;background:#c96;{{STYLE}}", ""), "The 6px sliver widens to the 12px minimum.", true), true
	case n == "min-inline-size":
		return lay(blockDemo("height:30px;width:6px;background:#c96;{{STYLE}}", ""), "The 6px sliver widens to the 20px minimum.", true), true
	case n == "min-height":
		return lay(blockDemo("width:150px;height:6px;background:#c96;{{STYLE}}", ""), "The 6px strip grows to the 12px minimum height.", true), true
	case n == "min-block-size":
		return lay(blockDemo("width:150px;height:6px;background:#c96;{{STYLE}}", ""), "The 6px strip grows to the 20px minimum height.", true), true
	case n == "min-size":
		return lay(blockDemo("width:6px;height:6px;background:#c96;{{STYLE}}", ""), "The 6px square grows to 20px by 20px, if the engine reads this draft shorthand.", true), true
	case n == "max-width":
		return lay(blockDemo("height:30px;width:200px;background:#c96;{{STYLE}}", ""), "The 200px bar is clamped to 12px.", true), true
	case n == "max-inline-size":
		return lay(blockDemo("height:30px;background:#c96;{{STYLE}}", ""), "The full-width bar is clamped to half the frame.", true), true
	case n == "max-height":
		return lay(blockDemo("width:150px;height:50px;background:#c96;{{STYLE}}", ""), "The 50px box is clamped to 12px tall.", true), true
	case n == "max-block-size":
		return lay(`<div id="{{ID}}" style="height:80px;width:280px;background:#e8e4d8;"><div id="{{ID}}-1" style="width:150px;height:70px;background:#c96;{{STYLE}}">box</div></div>`,
			"The 70px box is clamped to half the 80px frame height.", true), true
	case n == "max-size":
		return lay(`<div id="{{ID}}" style="height:80px;width:280px;background:#e8e4d8;"><div id="{{ID}}-1" style="width:200px;height:50px;background:#c96;{{STYLE}}"></div></div>`, "The 200px by 50px box is clamped to half its 280px by 80px parent in both directions (140px by 40px).", true), true
	case n == "size":
		return lay(blockDemo("background:#c96;{{STYLE}}", "text"), "The box takes a 100px by 50px size instead of full width and one text line, if the engine reads this draft shorthand.", true), true
	case n == "aspect-ratio":
		return lay(blockDemo("width:120px;background:#c96;{{STYLE}}", ""), "The empty 120px-wide box gets a 16:9 height of about 68px instead of collapsing to zero.", true), true
	case n == "box-sizing":
		return lay(blockDemo("width:150px;height:30px;padding:10px;border:3px solid #333;background:#c96;{{STYLE}}", ""), "The box keeps 150px as its outer width, so its content area is narrower, instead of growing to 176px.", true), true
	case n == "margin-trim":
		return lay(`<div id="{{ID}}" style="width:280px;border:1px solid #888;background:#e8e4d8;{{STYLE}}">`+
			kid(1, "height:20px;margin:12px;background:#c96;")+kid(2, "height:20px;margin:12px;background:#69c;")+`</div>`,
			"The container trims the first child's block-start margin and the last child's block-end margin, so the frame shortens.", true), true
	case n == "display":
		return lay(`<div id="{{ID}}" style="width:280px;background:#e8e4d8;"><span id="{{ID}}-1" style="background:#c96;{{STYLE}}">a span</span></div>`,
			"The orange span stretches to the full frame width as a block instead of hugging its text.", true), true
	case n == "visibility":
		return lay(blockDemo("width:150px;height:30px;background:#c96;{{STYLE}}", ""), "Accepted and exported, but the window still paints the orange box; it does not hide it yet.", false), true
	case n == "float":
		return lay(`<div id="{{ID}}" style="width:280px;height:40px;background:#e8e4d8;"><div id="{{ID}}-1" style="width:80px;height:30px;background:#c96;{{STYLE}}"></div></div>`,
			"The orange box moves to the right edge of the frame.", true), true
	case n == "clear":
		return lay(`<div id="{{ID}}" style="width:280px;background:#e8e4d8;overflow:hidden"><div style="float:left;width:80px;height:40px;background:#6a9"></div><div id="{{ID}}-1" style="height:20px;background:#c96;{{STYLE}}">cleared</div></div>`,
			"The orange bar drops below the left float instead of sitting beside it.", true), true
	case in(n, "float-offset", "float-defer", "float-reference"):
		return lay(`<div id="{{ID}}" style="width:280px;height:40px;background:#e8e4d8;"><div id="{{ID}}-1" style="float:right;width:80px;height:30px;background:#c96;{{STYLE}}"></div></div>`,
			floatLook(n), n == "float-offset"), true
	case in(n, "top", "left", "right", "bottom", "inset", "inset-block", "inset-block-start", "inset-block-end", "inset-inline", "inset-inline-start", "inset-inline-end"):
		return lay(absDemo("{{STYLE}}"), "The orange box moves to the declared offset from the frame edge.", true), true
	case n == "position":
		return lay(`<div id="{{ID}}" style="width:280px;height:80px;background:#e8e4d8;"><div id="{{ID}}-1" style="top:20px;left:30px;width:60px;height:24px;background:#c96;{{STYLE}}"></div></div>`,
			"The box shifts 30px right and 20px down once it is relatively positioned; without it the offsets are ignored.", true), true
	case n == "z-index":
		return lay(`<div id="{{ID}}" style="position:relative;width:280px;height:60px;background:#e8e4d8;"><div id="{{ID}}-1" style="position:absolute;left:10px;top:10px;width:80px;height:40px;background:#c96;{{STYLE}}"></div><div id="{{ID}}-2" style="position:absolute;left:50px;top:20px;width:80px;height:30px;background:#69c;"></div></div>`,
			"The inherited value is auto, so the blue box stays on top of the orange one and nothing changes.", false), true
	case in(n, "overflow", "overflow-x", "overflow-y", "overflow-block", "overflow-inline"):
		return lay(overflowDemo("{{STYLE}}"), "The tall orange content is cut at the 40px frame instead of spilling out below it.", true), true
	case strings.HasPrefix(n, "overflow-clip-margin"):
		return lay(overflowAll("overflow:clip;{{STYLE}}"), clipMarginLook(n), true), true
	case n == "overflow-wrap":
		return lay(`<div id="{{ID}}" style="width:100px;background:#e8e4d8;font-size:12px;{{STYLE}}">Zzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzz</div>`,
			"Accepted; the engine already breaks an unbreakable word at the 100px frame edge, so the control and the demo wrap the same way and nothing changes.", false), true
	case n == "column-gap":
		return lay(colKids("column-count:2;{{STYLE}}"), "The two columns sit 3px apart instead of the 12px default gap (normal is 1em, and the font is 12px), so each column grows wider and the blue box starts further left.", true), true
	case n == "column-width":
		return lay(colDemo("{{STYLE}}"), "A 12px column width is narrower than most words, so the paragraph splits into many very narrow columns holding a letter or two each (a real page would use a wider value).", true), true
	case in(n, "columns", "column-count"):
		return lay(colDemo("{{STYLE}}"), "The paragraph splits into two side-by-side columns instead of one full-width paragraph.", true), true
	case n == "column-fill":
		return lay(colKids("column-fill:auto;height:60px;{{STYLE}}"), "Accepted; the engine lays the two boxes out the same with or without balancing, so nothing moves.", false), true
	case n == "column-height":
		return lay(colDemo("column-count:2;{{STYLE}}"), "Draft property: columns get a fixed 12px height, so the box shrinks to one line of the first column and the rest of the text is not shown.", true), true
	case n == "column-wrap":
		return lay(colDemo("column-count:2;{{STYLE}}"), "Draft property: the value is the default, so the columns stay as they are.", false), true
	case n == "column-span":
		return lay(`<div id="{{ID}}" style="column-count:2;width:280px;background:#e8e4d8;font-size:12px;"><p style="margin:0">Alpha beta gamma delta epsilon zeta eta theta iota kappa lambda mu.</p>`+
			`<div id="{{ID}}-1" style="background:#c96;{{STYLE}}">Heading</div><p style="margin:0">Nu xi omicron pi rho sigma tau upsilon phi chi psi omega.</p></div>`,
			"The orange heading stretches across both columns and the text splits around it.", true), true
	case n == "contain":
		return lay(overflowDemo("{{STYLE}}"), "Paint containment clips the tall content at the 40px frame.", true), true
	case strings.HasPrefix(n, "contain-intrinsic"):
		return containIntrinsic(n), true
	case n == "container" || n == "container-type":
		return lay(`<div id="{{ID}}" style="width:280px;height:60px;background:#e8e4d8;"><div id="{{ID}}-1" style="display:inline-block;min-width:100px;background:#c96;font-size:12px;{{STYLE}}">content that sizes the box</div></div>`,
			"Inline-size containment removes the text from the box's width, so the orange box narrows from its text width to its 100px minimum and the text wraps.", true), true
	case n == "container-name" || n == "anchor-scope":
		return lay(`<div id="{{ID}}" style="width:200px;padding:6px;background:#c96;font-size:12px;{{STYLE}}">named, but nothing queries it</div>`,
			"Only a name for later queries or anchors is declared, so nothing visible changes.", false), true
	case n == "anchor-name":
		return lay(anchorDemo("{{STYLE}}", "position-anchor:--a;"), "The engine does not resolve anchor positioning for a name alone, so the blue label stays at the top-left; the name is accepted.", false), true
	case n == "position-anchor":
		return lay(anchorDemo("anchor-name:--a;", "{{STYLE}}"), "The engine does not attach the blue label to the named anchor, so it stays at the top-left; the name is accepted.", false), true
	case n == "position-area":
		return lay(anchorDemo("anchor-name:--{{ID}};", "position-anchor:--{{ID}};{{STYLE}}"), "The engine only moves the blue label down 24px, from the control's y 0 to y 24, still at the left edge and left of the orange anchor. It does not place the label above the anchor. Each row names its own anchor, so the picture and the readout agree.", false), true
	case in(n, "position-try", "position-try-fallbacks", "position-try-order", "position-visibility"):
		return lay(anchorDemo("anchor-name:--{{ID}};", "position-anchor:--{{ID}};position-area:top;{{STYLE}}"), "Declares how a fallback or visibility rule applies when the anchored box overflows. Control and demo both carry position-area:top, which the engine does not resolve, so the blue label stays at the left edge of the frame (x 0, y 24), left of the orange anchor, in both pictures and nothing changes.", false), true
	case n == "table-layout":
		return lay(`<table id="{{ID}}" style="width:200px;border-collapse:collapse;font-size:12px;{{STYLE}}"><tr><td id="{{ID}}-1" style="border:1px solid #888;background:#c96">a very long first cell text</td><td id="{{ID}}-2" style="border:1px solid #888;background:#69c">b</td></tr></table>`,
			"The columns split the table width evenly (fixed) instead of following the long first cell.", true), true
	case n == "caption-side":
		return lay(`<table id="{{ID}}" style="width:200px;border-collapse:collapse;font-size:12px;"><caption id="{{ID}}-1" style="background:#c96;{{STYLE}}">Caption</caption><tr><td id="{{ID}}-2" style="border:1px solid #888;background:#69c">cell</td></tr></table>`,
			"The orange caption moves from above the table to below it.", true), true
	case n == "empty-cells":
		return lay(`<table id="{{ID}}" style="width:200px;border-collapse:separate;font-size:12px;{{STYLE}}"><tr><td id="{{ID}}-1" style="border:2px solid #333;background:#c96;height:24px"></td><td id="{{ID}}-2" style="border:2px solid #333;background:#69c">full</td></tr></table>`,
			"The empty orange cell loses its border and background.", true), true
	case n == "max-lines":
		return lay(`<div id="{{ID}}" style="width:100px;background:#e8e4d8;font-size:12px;{{STYLE}}">one two three four five six seven eight nine ten eleven twelve</div>`,
			"The paragraph is cut from four lines to two, and the engine draws the ellipsis just past the right edge of the 100px box instead of inside it.", true), true
	case n == "zoom":
		return lay(blockDemo("width:200px;height:60px;font-size:12px;background:#c96;{{STYLE}}", "zoomed"), "The orange box scales down to 40% of its size (80px by 24px); its text keeps its full size and still fits inside the shrunken box.", true), true
	case n == "backface-visibility":
		return lay(`<div id="{{ID}}" style="width:120px;height:40px;border:1px dashed #888;"><div id="{{ID}}-1" style="width:120px;height:40px;background:#c96;transform:rotateY(180deg);{{STYLE}}">back</div></div>`,
			"The rotated box sits inside a dashed 120x40 frame; hidden removes the box from the picture, leaving the dashed frame empty.", true), true
	case n == "box-decoration-break":
		return lay(`<div id="{{ID}}" style="width:120px;font-size:12px;"><span id="{{ID}}-1" style="background:#c96;padding:0 8px;border:2px solid #333;{{STYLE}}">a span that wraps over several lines</span></div>`,
			"Both pictures draw borders around each wrapped line fragment; with clone the tan background clips tightly to each individual line fragment instead of filling one rectangular box behind the whole span.", true), true
	case n == "all":
		return lay(`<div id="{{ID}}" style="width:280px;color:#333;background:#e8e4d8;font-size:14px;"><div id="{{ID}}-1" style="width:100px;height:30px;padding:4px;background:#c96;{{STYLE}}">reset</div></div>`,
			"The inner box drops its brown fill and takes its parent's beige background; its 108x38 size and text position stay the same.", true), true
	case n == "field-sizing":
		return lay(`<div id="{{ID}}" style="width:280px;"><input id="{{ID}}-1" value="hi" style="{{STYLE}}"></div>`,
			"Accepted; the engine paints no operations for the bare input here, so nothing visible changes.", false), true
	}

	return lay(`<div id="{{ID}}" class="gbox" style="{{STYLE}}">{{NAME}}: {{VALUE}} (host only)</div>`, lookHost(n), false), true
}

// lookHost says what a host-only or fragmentation property does here.
func lookHost(n string) string {
	return "This draft or host property has no visible effect on a screen box, so the picture and readout match the control."
}

func lay(html, look string, visible bool) DemoSpec {
	return DemoSpec{HTML: html, Look: look, Visible: visible}
}

func in(n string, names ...string) bool {
	for _, s := range names {
		if n == s {
			return true
		}
	}

	return false
}

// kid is a measured child; its id is the demo prefix plus -n.
func kid(n int, style string) string {
	return `<div id="{{ID}}-` + string(rune('0'+n)) + `" style="` + style + `">` + string(rune('0'+n)) + `</div>`
}

// flexDemo is a row of three boxes of different sizes. cont, a, b and c are
// extra styles for the container and the three boxes.
func flexDemo(cont, a, b, c string) string {
	return `<div id="{{ID}}" style="display:flex;width:280px;height:80px;background:#e8e4d8;` + cont + `">` +
		kid(1, "width:50px;height:20px;background:#c96;"+a) + kid(2, "width:70px;height:36px;background:#6a9;"+b) + kid(3, "width:40px;height:28px;background:#69c;"+c) + `</div>`
}

// flexWrapDemo is a flex container with three 120px boxes: two fit on a line.
func flexWrapDemo(cont string) string {
	return `<div id="{{ID}}" style="display:flex;width:280px;background:#e8e4d8;` + cont + `">` +
		kid(1, "width:120px;height:20px;background:#c96;") + kid(2, "width:120px;height:20px;background:#6a9;") + kid(3, "width:120px;height:20px;background:#69c;") + `</div>`
}

func shrinkDemo() string {
	return `<div id="{{ID}}" style="display:flex;width:280px;height:40px;background:#e8e4d8;">` +
		kid(1, "width:150px;height:20px;background:#c96;{{STYLE}}") + kid(2, "width:150px;height:20px;background:#6a9;") + kid(3, "width:150px;height:20px;background:#69c;") + `</div>`
}

// gridDemo is a grid of six boxes. cont and first style the container and the
// first box; all styles every box.
func gridDemo(cont, first, all string) string {
	var b strings.Builder

	b.WriteString(`<div id="{{ID}}" style="display:grid;width:280px;background:#e8e4d8;` + cont + `">`)

	for i, c := range []string{"#c96", "#6a9", "#69c", "#c69", "#9c6", "#96c"} {
		st := "height:16px;background:" + c + ";" + all
		if i == 0 {
			st += first
		}

		b.WriteString(kid(i+1, st))
	}

	b.WriteString(`</div>`)

	return b.String()
}

// blockDemo is a framed block with one measured child.
func blockDemo(child, text string) string {
	return `<div id="{{ID}}" style="width:280px;border:1px solid #888;background:#e8e4d8;"><div id="{{ID}}-1" style="` + child + `">` + text + `</div></div>`
}

func absDemo(style string) string {
	return `<div id="{{ID}}" style="position:relative;width:280px;height:80px;background:#e8e4d8;"><div id="{{ID}}-1" style="position:absolute;width:60px;height:24px;background:#c96;` + style + `"></div></div>`
}

func overflowDemo(style string) string {
	return `<div id="{{ID}}" style="width:200px;height:40px;margin:14px 0 36px;background:#e8e4d8;` + style + `"><div id="{{ID}}-1" style="height:70px;background:#c96;">tall content</div></div>`
}

func colDemo(style string) string {
	return `<div id="{{ID}}" style="width:280px;background:#e8e4d8;font-size:12px;` + style + `">Alpha beta gamma delta epsilon zeta eta theta iota kappa lambda mu nu xi omicron pi rho.</div>`
}

// overflowAll has a checkerboard of 12px cells that spills out on every side
// of a bordered 160px by 40px frame, so a clip margin shows as whole extra
// rows or columns of cells on the side it applies to.
func overflowAll(style string) string {
	var b strings.Builder

	b.WriteString(`<div id="{{ID}}" style="width:160px;height:40px;margin:20px;border:2px solid #333;background:#e8e4d8;` + style + `">` +
		`<div id="{{ID}}-1" style="width:240px;height:108px;margin:-34px 0 0 -34px;">`)

	for row := range 9 {
		b.WriteString(`<div style="display:flex;height:12px;">`)

		for col := range 20 {
			c := "#c96"
			if (row+col)%2 == 1 {
				c = "#69c"
			}

			b.WriteString(`<div style="width:12px;height:12px;background:` + c + `"></div>`)
		}

		b.WriteString(`</div>`)
	}

	b.WriteString(`</div></div>`)

	return b.String()
}

func colKids(style string) string {
	return `<div id="{{ID}}" style="width:280px;background:#e8e4d8;font-size:12px;` + style + `">` +
		kid(1, "height:25px;background:#c96;") + kid(2, "height:25px;background:#69c;") + `</div>`
}

func anchorDemo(target, label string) string {
	return `<div id="{{ID}}" style="position:relative;width:280px;height:80px;background:#e8e4d8;">` +
		`<div id="{{ID}}-1" style="position:absolute;left:100px;top:44px;width:80px;height:24px;background:#c96;` + target + `">anchor</div>` +
		`<div id="{{ID}}-2" style="position:absolute;width:60px;height:20px;background:#69c;` + label + `">label</div></div>`
}

// containIntrinsic is an empty size-contained box. Padding gives the axis
// the property does not size a fixed extent, so the sized axis shows as extra
// orange; without the property that axis is empty.
func containIntrinsic(n string) DemoSpec {
	border, look := "border:3px solid #69c;", "The empty size-contained box gets a 12px intrinsic content size in both directions instead of only its blue border; it renders 12px larger than the control."

	switch n {
	case "contain-intrinsic-width", "contain-intrinsic-inline-size":
		border, look = "padding:12px 0;border:2px solid #69c;", "The size-contained box gets a 12px intrinsic content width, so the orange box is 12px wider between its blue edges; the height comes from the padding."
	case "contain-intrinsic-height", "contain-intrinsic-block-size":
		border, look = "padding:0 40px;border:2px solid #69c;", "The size-contained box gets a 12px intrinsic content height, so the orange box is 12px taller between its blue edges; the width comes from the padding."
	}

	return lay(`<div id="{{ID}}" style="width:280px;height:50px;background:#e8e4d8;"><div id="{{ID}}-1" style="display:inline-block;contain:size;background:#c96;`+border+`{{STYLE}}"></div></div>`, look, true)
}

// clipMarginLook names the side or sides a clip margin reveals.
func clipMarginLook(n string) string {
	side := strings.TrimPrefix(n, "overflow-clip-margin")

	where := map[string]string{
		"":              "on every side",
		"-block":        "on the top and bottom",
		"-inline":       "on the left and right",
		"-block-start":  "on the top",
		"-top":          "on the top",
		"-block-end":    "on the bottom",
		"-bottom":       "on the bottom",
		"-inline-start": "on the left",
		"-left":         "on the left",
		"-inline-end":   "on the right",
		"-right":        "on the right",
	}[side]

	return "The checkerboard spills past the dark-bordered frame and is clipped at its edge in the control; with the clip margin one more 12px row or column of squares shows " + where + "."
}

// marginLook says what one margin property does to the demo frame.
func marginLook(n string) string {
	side := map[string]string{
		"margin-top": "above", "margin-block-start": "above", "margin-bottom": "below", "margin-block-end": "below",
		"margin-left": "left of", "margin-inline-start": "left of", "margin-right": "right of", "margin-inline-end": "right of",
	}[n]
	switch n {
	case "margin-top", "margin-block-start", "margin-bottom", "margin-block-end":
		return "Empty space opens " + side + " the orange box; the box keeps its size and the frame grows on that one side to hold the margin."
	case "margin-left", "margin-right", "margin-inline-start", "margin-inline-end":
		return "Empty space opens " + side + " the orange box, so it moves away from that edge inside the frame (or narrows to leave the space)."
	}

	if n == "margin-inline" {
		return "Empty space opens on the left and right of the orange box, so it narrows and moves in from both edges of the frame; its height does not change."
	}

	if n == "margin-block" {
		return "Empty space opens above and below the orange box; the box keeps its size and the frame grows to hold the vertical margins."
	}

	return "Empty space opens around the orange box, so it narrows or moves inside the frame and the frame grows to hold the vertical margins."
}

// floatLook says what a draft float property does to the demo float.
func floatLook(n string) string {
	if n == "float-offset" {
		return "The orange float stays at the right edge, but the offset shifts its box 12px down."
	}

	return "The orange float sits at the right edge; this draft property does not change that, so nothing moves."
}
