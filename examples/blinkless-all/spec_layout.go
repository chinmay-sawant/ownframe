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
		return lay(flexWrapDemo("flex-wrap:wrap;{{STYLE}}height:100px;"), "The two wrapped lines of boxes move from the top of the container to its vertical centre.", true), true
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
		return lay(flexDemo("", "width:auto;{{STYLE}}", "", ""), "The first box starts from half the container width instead of its content width.", true), true
	case n == "flex-shrink":
		return lay(shrinkDemo(), "The first box gives up more of its width when the three boxes overflow the row.", true), true
	case n == "flex-direction":
		return lay(flexDemo("{{STYLE}}", "", "", ""), "The row runs right to left: the blue box is first at the right edge.", true), true
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
		cont, kidStyle := "", ""
		if strings.HasSuffix(n, "self") {
			kidStyle = "{{STYLE}}"
		} else {
			cont = "{{STYLE}}"
		}

		return lay(gridDemo("grid-template-columns:1fr 1fr;"+cont, kidStyle, "width:40px;"), "The 40px boxes centre inside their wider grid cells instead of sitting at the cell's start edge.", true), true

	case n == "grid" || n == "grid-template-columns":
		return lay(gridDemo("{{STYLE}}", "", ""), "The six boxes lay out in columns instead of one stacked column.", true), true
	case n == "grid-template-rows" || n == "grid-template":
		return lay(gridDemo("{{STYLE}}", "", ""), "The grid rows take the declared heights instead of the height of their content.", true), true
	case n == "grid-auto-rows":
		return lay(gridDemo("{{STYLE}}", "", ""), "Every implicit row is 40px tall.", true), true
	case n == "grid-auto-columns":
		return lay(gridDemo("grid-auto-flow:column;{{STYLE}}", "", ""), "Accepted; the engine lays implicit columns out the same, so the widths do not change.", false), true
	case n == "grid-auto-flow":
		return lay(gridDemo("grid-template-rows:20px 20px;{{STYLE}}", "", ""), "Accepted; the engine keeps filling row by row, so the boxes do not move.", false), true
	case n == "grid-template-areas":
		return lay(gridDemo("{{STYLE}}", "", ""), "The container declares a one-cell named area; the boxes keep the grid's layout.", false), true
	case n == "grid-row-start" || n == "grid-row-end":
		return lay(gridDemo("grid-template-columns:1fr 1fr 1fr;", "{{STYLE}}", ""), "Accepted; the engine places the first box in the same cell, so nothing moves (grid-row shows the placement).", false), true
	case in(n, "grid-area", "grid-column", "grid-column-start", "grid-column-end", "grid-row"):
		return lay(gridDemo("grid-template-columns:1fr 1fr 1fr;", "{{STYLE}}", ""), "The first box moves to the placed grid line, or spans cells, instead of the first cell.", true), true

	case in(n, "margin", "margin-left", "margin-right", "margin-top", "margin-bottom", "margin-block", "margin-block-start", "margin-block-end", "margin-inline", "margin-inline-start", "margin-inline-end"):
		return lay(blockDemo("height:30px;background:#c96;{{STYLE}}", ""), "Empty space opens around the orange box, so it moves or shrinks inside the frame.", true), true
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
		return lay(blockDemo("width:200px;height:50px;background:#c96;{{STYLE}}", ""), "The box is clamped to half its parent in both directions, if the engine reads this draft shorthand.", true), true
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
			"The orange float sits at the right edge; this draft float property shifts or keeps it only by what the engine implements.", n == "float-offset"), true
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
		return lay(overflowAll("overflow:clip;{{STYLE}}"), "The clip edge sits 12px outside the frame, so more of the overflowing content shows.", true), true
	case n == "overflow-wrap":
		return lay(`<div id="{{ID}}" style="width:100px;background:#e8e4d8;font-size:12px;{{STYLE}}">Unbreakable_supercalifragilistic</div>`,
			"The long word breaks across lines inside the 100px frame instead of overflowing it.", true), true
	case n == "column-gap":
		return lay(colKids("column-count:2;{{STYLE}}"), "The second box sits 3px, not 16px, from the first column.", true), true
	case in(n, "columns", "column-count", "column-width"):
		return lay(colDemo("{{STYLE}}"), "The paragraph splits into side-by-side columns.", true), true
	case n == "column-fill":
		return lay(colKids("column-fill:auto;height:60px;{{STYLE}}"), "Accepted; the engine lays the two boxes out the same with or without balancing, so nothing moves.", false), true
	case n == "column-height":
		return lay(colDemo("column-count:2;{{STYLE}}"), "Draft property: columns get a fixed height.", true), true
	case n == "column-wrap":
		return lay(colDemo("column-count:2;{{STYLE}}"), "Draft property: the value is the default, so the columns stay as they are.", false), true
	case n == "column-span":
		return lay(`<div id="{{ID}}" style="column-count:2;width:280px;background:#e8e4d8;font-size:12px;"><p style="margin:0">Alpha beta gamma delta epsilon zeta eta theta iota kappa lambda mu.</p>`+
			`<div id="{{ID}}-1" style="background:#c96;{{STYLE}}">Heading</div><p style="margin:0">Nu xi omicron pi rho sigma tau upsilon phi chi psi omega.</p></div>`,
			"The orange heading stretches across both columns and the text splits around it.", true), true
	case n == "contain":
		return lay(overflowDemo("{{STYLE}}"), "Paint containment clips the tall content at the 40px frame.", true), true
	case strings.HasPrefix(n, "contain-intrinsic"):
		return lay(`<div id="{{ID}}" style="display:inline-block;contain:size;background:#c96;font-size:12px;{{STYLE}}">content</div>`,
			"The size-contained box has an intrinsic size of 12px instead of collapsing to nothing.", true), true
	case n == "container" || n == "container-type":
		return lay(`<div id="{{ID}}" style="display:inline-block;background:#c96;font-size:12px;{{STYLE}}">content that sizes the box</div>`,
			"Inline-size containment removes the text from the box's width, so the box shrinks.", true), true
	case n == "container-name" || n == "anchor-scope":
		return lay(`<div id="{{ID}}" style="width:200px;padding:6px;background:#c96;font-size:12px;{{STYLE}}">named, but nothing queries it</div>`,
			"Only a name for later queries or anchors is declared, so nothing visible changes.", false), true
	case n == "anchor-name":
		return lay(anchorDemo("{{STYLE}}", "position-anchor:--a;"), "The engine does not resolve anchor positioning for a name alone, so the blue label stays at the top-left; the name is accepted.", false), true
	case n == "position-anchor":
		return lay(anchorDemo("anchor-name:--a;", "{{STYLE}}"), "The engine does not attach the blue label to the named anchor, so it stays at the top-left; the name is accepted.", false), true
	case n == "position-area":
		return lay(anchorDemo("anchor-name:--a;", "position-anchor:--a;{{STYLE}}"), "The blue label is placed in the area above the orange anchor.", true), true
	case in(n, "position-try", "position-try-fallbacks", "position-try-order", "position-visibility"):
		return lay(anchorDemo("anchor-name:--a;", "position-anchor:--a;position-area:top;{{STYLE}}"), "Declares how a fallback or visibility rule applies when the anchored box overflows; here it fits, so nothing changes.", false), true
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
			"The paragraph is cut to two lines.", true), true
	case n == "zoom":
		return lay(blockDemo("width:200px;height:30px;background:#c96;{{STYLE}}", "zoomed"), "The orange box scales down to 40% of its size.", true), true
	case n == "backface-visibility":
		return lay(`<div id="{{ID}}" style="width:120px;height:40px;background:#c96;transform:rotateY(180deg);{{STYLE}}">back</div>`,
			"The box is turned to face away; hidden removes it from the picture.", true), true
	case n == "box-decoration-break":
		return lay(`<div id="{{ID}}" style="width:120px;font-size:12px;"><span id="{{ID}}-1" style="background:#c96;padding:0 8px;border:2px solid #333;{{STYLE}}">a span that wraps over several lines</span></div>`,
			"Each wrapped line fragment gets its own padding and border instead of one shared set at the ends.", true), true
	case n == "all":
		return lay(`<div id="{{ID}}" style="width:280px;color:#333;background:#e8e4d8;font-size:14px;"><div id="{{ID}}-1" style="width:100px;height:30px;padding:4px;background:#c96;{{STYLE}}">reset</div></div>`,
			"The inner box drops its own width, padding and fill and takes its parent's values.", true), true
	case n == "field-sizing":
		return lay(`<div id="{{ID}}" style="width:280px;"><input id="{{ID}}-1" value="hi" style="{{STYLE}}"></div>`,
			"Accepted; the engine paints no operations for the bare input here, so nothing visible changes.", false), true
	}

	return lay(`<div id="{{ID}}" class="gbox" style="{{STYLE}}">{{NAME}}: {{VALUE}} (host only)</div>`, lookHost(n), false), true
}

// lookHost says what a host-only or fragmentation property does here.
func lookHost(n string) string {
	return "This draft or host property has no visible effect on a screen box; the readout shows the value reaching the page."
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
	return `<div id="{{ID}}" style="width:200px;height:40px;margin:14px 0;background:#e8e4d8;` + style + `"><div id="{{ID}}-1" style="height:70px;background:#c96;">tall content</div></div>`
}

func colDemo(style string) string {
	return `<div id="{{ID}}" style="width:280px;background:#e8e4d8;font-size:12px;` + style + `">Alpha beta gamma delta epsilon zeta eta theta iota kappa lambda mu nu xi omicron pi rho sigma tau upsilon.</div>`
}

// overflowAll has content that spills out on every side of its frame.
func overflowAll(style string) string {
	return `<div id="{{ID}}" style="width:200px;height:40px;margin:20px 20px;background:#e8e4d8;` + style + `"><div id="{{ID}}-1" style="height:70px;margin:-15px;background:#c96;">tall content</div></div>`
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
