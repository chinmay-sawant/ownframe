package main

import "strings"

// vizPNG names the page image: a 20x20 green square, opaque left half and
// faint right half, which the example answers for the name "mask".
const vizPNG = "mask"

// vizStripes is a hard-stop backdrop for blend, filter and mask demos.
const vizStripes = "linear-gradient(90deg,#e74c3c 0 25%,#f1c40f 25% 50%,#2ecc71 50% 75%,#3498db 75%)"

const vizGrad = "linear-gradient(#c0392b,#2980b9)"

// vizBox is the 140x60 coloured box with a baseline style and the property.
func vizBox(base, inner string) string {
	return `<div id="{{ID}}" style="width:140px;height:60px;margin:18px 24px;background:#dfe9f3;` + base + `{{STYLE}}">` + inner + `</div>`
}

// vizPlain is a box with no background of its own, for properties that draw it.
func vizPlain(base, inner string) string {
	return `<div id="{{ID}}" style="width:140px;height:60px;margin:4px 10px;` + base + `{{STYLE}}">` + inner + `</div>`
}

// vizWrap is a parent that the property's element sits inside; the child is
// measured as {{ID}}-1 and carries the declaration.
func vizWrap(parent, child, inner string) string {
	return `<div id="{{ID}}" style="width:200px;height:90px;` + parent + `"><div id="{{ID}}-1" style="` + child + `{{STYLE}}">` + inner + `</div></div>`
}

func vizOver(parent, child string) string {
	return vizWrap("position:relative;background:"+vizStripes+";", "position:absolute;left:30px;top:15px;width:100px;height:50px;"+child, "")
}

// vizRules is a two by two grid with 16 px gaps, where column and row rules draw.
func vizRules(base string) string {
	cell := func(n string) string {
		return `<div style="background:#dfe9f3;height:36px;padding:8px 12px;box-sizing:border-box">` + n + `</div>`
	}

	return `<div id="{{ID}}" style="display:grid;grid-template-columns:1fr 1fr;gap:16px;width:280px;margin:14px 0 16px;font-size:11px;` + base + `{{STYLE}}">` +
		cell("1") + cell("2") + cell("3") + cell("4") + `</div>`
}

func vizList(base string) string {
	return `<ul id="{{ID}}" style="width:140px;margin:0;padding-left:28px;font-size:13px;` + base + `{{STYLE}}"><li style="padding-left:12px;text-indent:16px">first</li><li style="padding-left:12px;text-indent:16px">second</li><li style="padding-left:12px;text-indent:16px">third</li></ul>`
}

func vizImg(base string) string {
	return `<div id="{{ID}}" style="width:160px"><img id="{{ID}}-1" src="` + vizPNG + `" style="width:140px;height:60px;background:#dfe9f3;` + base + `{{STYLE}}"></div>`
}

// vizDisc paints the float as a disc, so text that wraps into its corners does
// not run over a painted box.
const vizDisc = "radial-gradient(circle closest-side,#c0392b,#2980b9 99%,transparent 100%)"

func vizFloat(base string) string { return vizFloatBg(vizDisc, base) }

func vizFloatBg(bg, base string) string {
	return `<div id="{{ID}}" style="width:260px;font-size:12px;line-height:16px;height:100px;overflow:hidden"><div id="{{ID}}-1" style="float:left;width:80px;height:80px;margin-right:4px;background:` + bg + `;` + base + `{{STYLE}}"></div>` +
		`<p id="{{ID}}-2" style="margin:0">The quick brown fox jumps over the lazy dog and keeps running across the field until the text wraps around the float on its left side.</p></div>`
}

func vizSVG(base, shape string) string {
	return `<svg id="{{ID}}" width="140" height="70" viewBox="0 0 140 70" style="margin:6px;background:#dfe9f3">` + shape + `</svg>`
}

// visualSpec returns the demo of a visual property, or false when another group
// file answers for it.
func visualSpec(r genProp) (DemoSpec, bool) {
	var (
		d  DemoSpec
		ok bool
	)

	switch r.Group {
	case "border":
		d, ok = borderSpec(r)
	case "paint":
		d, ok = paintSpec(r)
	case "fx":
		d, ok = fxSpec(r)
	}

	if !ok {
		return DemoSpec{}, false
	}

	if look, found := vizLook[r.Name]; found {
		d.Look = look
	}

	if vizReadoutBlind[r.Name] || vizLook[r.Name] != "" {
		d.Visible = false
	}

	return d, true
}

// vizReadoutBlind lists the properties whose effect shows in the picture but
// not in the readout, which compares box geometry and operation signatures
// without gradients, clips, masks or matrices.
var vizReadoutBlind = map[string]bool{
	"clip-path": true, "mask-composite": true, "mask-repeat": true, "object-position": true,
	"transform-origin": true, "border-image": true, "border-image-repeat": true, "border-image-width": true,
	"border-image-slice": true, "border-bottom-color": true, "border-left-color": true, "border-right-color": true,
	"border-top-color": true, "outline-color": true, "rule-break": true, "rule-inset-junction": true,
	"column-rule-inset-junction": true, "column-rule-inset-junction-end": true, "column-rule-inset-junction-start": true,
	"row-rule-inset-junction": true, "row-rule-inset-junction-end": true, "row-rule-inset-junction-start": true,
}

// vizLook replaces the sentence for a property the painter draws no differently
// from the control.
var vizLook = map[string]string{
	"backdrop-filter":       "The translucent box looks the same as the control: the painter does not blur what is behind it.",
	"filter":                "The white box and its text stay sharp: the painter does not apply blur().",
	"scale":                 "The box keeps its size: the painter does not apply the percentage scale.",
	"transform-box":         "The rect looks the same as the control; transform-box is accepted but changes nothing on this rect.",
	"background-blend-mode": "The picture matches the control (the painter does not blend the background layers); the engine's operation list differs.",
	"background-color":      "The inner box looks the same as the control: the parent's amber shows through the transparent control and the inherited amber alike.",
	"image-rendering":       "The picture matches the control (the painter scales the image with one filter either way); the engine's operation list differs.",
	"mask-mode":             "The masked box looks the same as the control: alpha is already how a gradient mask is read.",
	"object-fit":            "The picture matches the control (the painter stretches the image to the box whichever fit is set); the engine's operation list differs.",
	"object-view-box":       "The picture matches the control (the painter ignores the view box crop); the engine's operation list differs.",
	"clip-rule":             "The red rect is not clipped to the star in the control or the demo: the painter ignores the clipPath, so clip-rule has no shape to apply to.",
}

func spec(html, look string) (DemoSpec, bool) {
	return DemoSpec{HTML: html, Look: look, Visible: true}, true
}

func specNo(html, look string) (DemoSpec, bool) {
	return DemoSpec{HTML: html, Look: look}, true
}

// ruleSpec covers column-rule*, row-rule* and rule*.
func ruleSpec(r genProp) (DemoSpec, bool) {
	n := r.Name
	p := n
	suffix := ""

	if i := strings.Index(n, "rule-"); i >= 0 {
		p, suffix = n[:i+4], n[i+5:]
	}

	look := "Rule lines between the grid cells (the grid has 16 px gaps)."
	base := "rule:2px solid #c0392b;"

	switch {
	case suffix == "" || n == p:
		return spec(vizRules(""), "Rule lines appear in the gaps between the grid cells.")
	case suffix == "color":
		return spec(vizRules(p+":3px solid #2c3e50;"), "The rule lines change colour to "+r.Sample+".")
	case suffix == "style":
		return spec(vizRules(p+"-width:3px;"+p+"-color:#2c3e50;"), "The rules go from absent (style none) to "+r.Sample+" lines.")
	case suffix == "width":
		return spec(vizRules(p+":1px solid #2c3e50;"), "The rule lines thicken to "+r.Sample+".")
	}

	switch {
	case strings.HasSuffix(n, "rule-break"):
		return specNo(vizRules(base), "Rules in the grid gaps. In the control the vertical rule runs straight through the crossing. With "+n+": "+r.Sample+" it is cut into two pieces that stop at the junction, and the picture differs from the control.")
	case n == "rule-overlap":
		return specNo(vizRules(base), "Rules in the grid gaps. Both rules have the same colour and width, so which one paints over the other at the crossing cannot be seen: the picture looks the same as the control while the engine's operation order differs.")
	case strings.Contains(n, "junction"):
		return specNo(vizRules(base), look+" "+n+" ("+r.Sample+") only trims rules where they meet at junctions; this 2 by 2 grid has one crossing, and the picture is the same as the control.")
	case strings.HasSuffix(n, "rule-visibility-items"):
		return spec(vizRules(base), "Rules in the grid gaps. With "+n+": "+r.Sample+" the rule set changes: "+visibilityLook(n, r.Sample)+" The control draws only the rules between cells.")
	}

	return spec(vizRules(base), look+" The property changes how the "+r.Sample+" value trims or limits them.")
}

func visibilityLook(n, v string) string {
	switch v {
	case "around", "all":
		if strings.HasPrefix(n, "column-") {
			return "extra vertical rules also appear along the left and right outer edges of the grid, beside the one between the columns."
		}

		return "extra horizontal rules also appear along the top and bottom outer edges of the grid, beside the one between the rows."
	case "between":
		return "rules draw only between the cells."
	}

	return "rules follow the normal rule."
}

func borderSpec(r genProp) (DemoSpec, bool) {
	n := r.Name

	if r.Sample == "inherit" && (strings.HasSuffix(n, "-color") || n == "outline-color") {

		return specNo(`<div id="{{ID}}" style="width:170px;margin:4px 10px;padding:10px;color:#2c3e50;border:2px solid #c0392b;border-color:#c0392b;outline:2px solid #c0392b;background:#f1efe7"><div id="{{ID}}-1" style="width:140px;height:50px;background:#dfe9f3;`+vizEdge(n)+`{{STYLE}}"></div></div>`,
			"The inner box keeps its grey edge or outline in the picture: the parent's border is red, but the engine does not resolve inherit for this edge colour, so demo and control paint the same.")
	}

	switch {
	case strings.Contains(n, "rule"):
		return ruleSpec(r)
	case strings.HasPrefix(n, "box-shadow"):
		return shadowSpec(r)
	case strings.HasPrefix(n, "outline"):
		return outlineSpec(r)
	case strings.HasPrefix(n, "border-image"):
		return borderImageSpec(r)
	case strings.HasPrefix(n, "corner"):
		return cornerSpec(r)
	}

	switch n {
	case "border-collapse":
		return specNo(`<table id="{{ID}}" style="border-collapse:collapse;border-spacing:6px;margin:10px 0;{{STYLE}}"><tr><td id="{{ID}}-1" style="border:3px solid #2c3e50;padding:12px 22px">a</td><td id="{{ID}}-2" style="border:3px solid #2c3e50;padding:12px 22px">b</td></tr></table>`,
			"The control collapses the cell borders into single shared lines. Here the cells keep the separated model: each cell shows its own full 3 px border with a 6 px gap, so the borders read as doubled and the table is wider and taller.")
	case "border-spacing":
		return spec(`<table id="{{ID}}" style="border-collapse:separate;{{STYLE}}"><tr><td id="{{ID}}-1" style="border:2px solid #2c3e50;padding:6px">a</td><td id="{{ID}}-2" style="border:2px solid #2c3e50;padding:6px">b</td></tr></table>`,
			"The gap between the two bordered cells grows to 12 px and the table gets larger.")
	case "border-boundary":
		return specNo(vizBox("border:3px solid #2c3e50;", ""), "The box keeps its border; border-boundary is a hint that decides which parent edge clips the border, so nothing moves here.")
	case "border-limit":
		return specNo(vizBox("border:3px solid #2c3e50;border-radius:30px;", ""), "The box keeps its rounded border; border-limit caps how far a border draws, so nothing moves here.")
	case "border-shape":
		return specNo(`<div style="margin:16px 0">`+vizBox("border:3px solid #2c3e50;", "")+`</div>`, "The box keeps its rectangular border; the circle() shape is accepted but not painted.")
	case "border":
		return spec(vizBox("", ""), "A 3 px dashed red border appears around the box.")
	}

	look := func(what string) string { return what }

	switch {
	case strings.HasSuffix(n, "-radius"):
		return spec(vizBox("border:2px solid #2c3e50;", ""), look("The matching corner (or corners) of the box turn round with a 12 px radius."))
	case n == "border-radius":
		return spec(vizBox("border:2px solid #2c3e50;", ""), "All four corners turn round with a 12 px radius.")
	case strings.HasSuffix(n, "-style"):
		return spec(vizBox("border-width:4px;border-color:#2c3e50;", ""), "The named sides change from a solid line to a dashed one.")
	case strings.HasSuffix(n, "-width"):
		return spec(vizBox("border:1px solid #2c3e50;", ""), "The border on the named sides thickens from 1 px to 3 px and the box grows.")
	case strings.HasSuffix(n, "-color"):
		return spec(vizBox("border:4px solid #2c3e50;", ""), "The border on the named sides turns from dark slate to "+r.Sample+".")
	case strings.HasSuffix(n, "-clip"):
		return spec(vizBox("border:5px solid #2c3e50;", ""), "The border on the named sides is clipped, leaving a gap in the line.")
	case n == "border-color":
		return spec(vizBox("border:4px solid #2c3e50;", ""), "The whole border turns red.")
	case n == "border-style":
		return spec(vizBox("border-width:4px;border-color:#2c3e50;", ""), "The border goes from absent to a dashed line on all four sides.")
	case n == "border-width":
		return spec(vizBox("border:1px solid #2c3e50;", ""), "The border thickens from 1 px to 3 px on all four sides.")
	case n == "border-top" || n == "border-bottom" || n == "border-left" || n == "border-right" || n == "border-block" || n == "border-inline":
		return spec(vizBox("", ""), "A 3 px dashed red border appears on the named side or sides only.")
	}

	// border-block-start etc.: the sample is a bare style.
	return spec(vizBox("", ""), "A dashed border (default width, current colour) appears on the named side.")
}

func vizEdge(n string) string {
	if n == "outline-color" {
		return "outline:3px solid #888888;"
	}

	return "border:4px solid #888888;"
}

func shadowSpec(r genProp) (DemoSpec, bool) {
	base := ""

	look := ""

	switch r.Name {
	case "box-shadow":
		return spec(vizBox("", ""), "A hard 4 px offset slate shadow appears below and to the right of the box.")
	case "box-shadow-color":
		base = "box-shadow-offset:8px;"
		look = "The 8 px offset shadow below and to the right of the box turns red; the control's shadow uses the default colour."
	case "box-shadow-blur":
		base = "box-shadow-offset:8px;box-shadow-color:#c0392b;"
		look = "The red 8 px offset shadow softens: its edge spreads out by 12 px instead of staying hard."
	case "box-shadow-offset":
		base = "box-shadow-color:#c0392b;"
		look = "The red shadow, which sits directly behind the box in the control, moves 12 px right and 12 px down."
	case "box-shadow-spread":
		base = "box-shadow-offset:8px;box-shadow-color:#c0392b;"
		look = "The red shadow grows 12 px on every side, so the red band is much thicker than in the control."
	case "box-shadow-inset", "box-shadow-position":
		base = "box-shadow-spread:6px;box-shadow-color:#2c3e50;"
		look = "In the control the 6 px slate spread draws as a thick frame outside the box. Here the outside frame is gone and the spread shows only as a thin pale-grey band along the inside edge of the box."
	}

	return spec(vizBox(base, ""), look)
}

func outlineSpec(r genProp) (DemoSpec, bool) {
	switch r.Name {
	case "outline":
		return spec(vizBox("", ""), "A 3 px dashed red outline appears around the box, outside its border edge.")
	case "outline-color":
		return spec(vizBox("outline:3px solid #2c3e50;", ""), "The outline changes from slate to the inherited border colour of the wrapper (red).")
	case "outline-offset":
		return spec(vizBox("outline:3px solid #2c3e50;", ""), "The outline moves 12 px away from the box edge.")
	case "outline-style":
		return spec(vizBox("outline-width:3px;outline-color:#2c3e50;", ""), "The outline changes from a solid line to a dashed one.")
	}

	return spec(vizBox("outline:1px solid #2c3e50;", ""), "The outline thickens from 1 px to 3 px.")
}

func borderImageSpec(r genProp) (DemoSpec, bool) {
	img := "border:12px solid #cccccc;border-image-source:url(" + vizPNG + ");border-image-slice:8;"

	switch r.Name {
	case "border-image":
		return specNo(vizBox("border:12px solid #cccccc;", ""), "The grey border stays as drawn: the painter does not draw the gradient border image, so the picture matches the control and only the engine's operations may differ.")
	case "border-image-source":
		return spec(vizBox("border:12px solid #cccccc;border-image-slice:10;", ""), "The grey border disappears: the gradient border image replaces it and the painter draws that gradient as nothing.")
	case "border-image-outset":
		return spec(vizBox(img, ""), "The green sample border image is drawn 12 px further out, so it surrounds a larger area than in the control.")
	case "border-image-repeat":
		return spec(vizBox(img, ""), "The green sample border image tiles round along each edge as small repeated cells instead of stretching. The pale column on the right edge is the faint right half of the 20 px sample image, not a gap.")
	}

	return specNo(vizBox(img, ""), "The green border image looks the same as the control: the painter draws "+r.Name+" ("+r.Sample+") with the same result as the default.")
}

func cornerSpec(r genProp) (DemoSpec, bool) {
	if strings.HasSuffix(r.Name, "shape") {
		if r.Sample == "bevel" {
			return spec(vizBox("border:2px solid #2c3e50;border-radius:20px;", ""), "All four rounded corners become straight diagonal cuts (bevel) instead of curves.")
		}

		return specNo(vizBox("border:2px solid #2c3e50;border-radius:20px;", ""), "The corner looks the same as the control: a superellipse exponent of "+r.Sample+" is the ordinary round curve the box already has.")
	}

	return spec(vizBox("border:2px solid #2c3e50;", ""), "The named corner or corners get a 12 px round corner.")
}

func paintSpec(r genProp) (DemoSpec, bool) {
	n := r.Name

	switch {
	case strings.HasPrefix(n, "mask"):
		return maskSpec(r)
	case strings.HasPrefix(n, "background"):
		return backgroundSpec(r)
	case strings.HasPrefix(n, "shape-") && n != "shape-rendering":
		return shapeSpec(r)
	case strings.HasPrefix(n, "list-style"):
		return listSpec(r)
	case strings.HasPrefix(n, "object-"):
		return objectSpec(r)
	}

	switch n {
	case "clip":
		return specNo(`<div id="{{ID}}" style="width:200px;height:60px;background:#dfe9f3;"><div id="{{ID}}-1" style="position:relative;width:140px;height:50px;background:#e8b04a;{{STYLE}}"></div></div>`,
			"The child stays the same: clip: inherit takes the parent's initial auto, so nothing is clipped.")
	case "clip-path":
		return spec(vizBox("background:"+vizGrad+";", ""), "The gradient is cut to a circle covering 40% of the box; the pale box background behind it stays square.")
	case "clip-rule":
		return specNo(vizSVG("", `<clipPath id="{{ID}}-cp"><path style="{{STYLE}}" d="M70 5 L90 62 L40 25 L100 25 L50 62 Z"/></clipPath><rect x="10" y="5" width="120" height="60" fill="#c0392b" clip-path="url(#{{ID}}-cp)"/>`),
			"A red rect is clipped to a five-point star path; evenodd would leave the centre of the star empty where the painter supports clip-rule.")
	case "image-animation":
		return specNo(vizImg(""), "The image shows the same still frame; a static PNG has no animation to pause.")
	case "image-orientation":
		return specNo(vizImg(""), "The image shows the same picture; image-orientation none keeps the stored orientation, which a PNG already has.")
	case "image-rendering":
		return spec(`<div id="{{ID}}" style="width:160px"><img id="{{ID}}-1" src="`+vizPNG+`" style="width:140px;height:60px;{{STYLE}}"></div>`, "The 20x20 sample image scales up with hard pixel edges instead of a smooth blend.")
	case "image-resolution":
		return spec(`<div id="{{ID}}" style="width:160px"><img id="{{ID}}-1" src="`+vizPNG+`" style="{{STYLE}}"></div>`, "The unsized 20x20 image gets smaller: at 2 dppx its natural size halves from 20 px to 10 px.")
	case "shape-rendering":
		return specNo(vizSVG("", `<circle cx="70" cy="35" r="26" fill="#c0392b" style="{{STYLE}}"/>`), "A red circle with crispEdges rendering; the painter anti-aliases both, so the picture is the same.")
	}

	return specNo(vizBox("", ""), "A plain box: "+n+" has no demo here.")
}

func backgroundSpec(r genProp) (DemoSpec, bool) {
	n := r.Name
	img := "background-image:linear-gradient(#c0392b,#c0392b);background-repeat:no-repeat;background-size:40px 30px;"

	switch {
	case n == "background":
		return spec(vizPlain("border:1px solid #888;", ""), "A solid red background fills the box (the missing x.png adds nothing).")
	case n == "background-color":
		return spec(`<div id="{{ID}}" style="width:200px;padding:8px;background:#e8b04a"><div id="{{ID}}-1" style="width:140px;height:60px;border:1px solid #888;{{STYLE}}"></div></div>`,
			"The inner box takes the parent's amber background through inherit; the control is transparent.")
	case n == "background-image":
		return spec(vizPlain("border:1px solid #888;", ""), "A red to blue gradient fills the box.")
	case n == "background-attachment":
		return specNo(vizPlain(img+"border:1px solid #888;", ""), "The small red square sits in the same place; fixed only changes how it moves on scroll.")
	case n == "background-blend-mode":
		return spec(vizPlain("background-image:"+vizGrad+";background-color:#f1c40f;border:1px solid #888;", ""), "The gradient multiplies with the yellow background colour, so the colours darken.")
	case n == "background-clip":
		return spec(vizPlain("padding:12px;border:6px solid rgba(44,62,80,.4);background:#e8b04a;", ""), "The amber fill now stops at the content edge, inside the padding and the border, instead of reaching the outer border edge.")
	case n == "background-origin":
		return spec(vizPlain("padding:16px;border:2px solid #888;background-image:linear-gradient(#c0392b,#c0392b);background-repeat:no-repeat;background-size:40px 30px;", ""), "The small red square moves in by the 16 px padding to the content corner.")
	case n == "background-size":
		return spec(vizPlain(img+"border:1px solid #888;", ""), "The red image grows from a 40x30 square to cover the whole box.")
	case n == "background-repeat" || strings.HasPrefix(n, "background-repeat"):
		return spec(vizPlain(img+"border:1px solid #888;background-size:30px 20px;", ""), "The small red tile repeats along the axis the property names instead of sitting once at the top left.")
	case strings.HasPrefix(n, "background-position"):
		return spec(vizPlain(img+"border:1px solid #888;", ""), "The small red square moves 12 px from the top left along the named axis.")
	}

	return specNo(vizPlain("border:1px solid #888;", ""), "The box looks the same: "+n+" is accepted but draws nothing new.")
}

func maskSpec(r genProp) (DemoSpec, bool) {
	n := r.Name
	maskBase := "mask-image:linear-gradient(#000,transparent);"
	body := func(base string) string {
		return `<div id="{{ID}}" style="width:140px;height:60px;margin:4px 10px;background:` + vizStripes + `;` + base + `{{STYLE}}"></div>`
	}

	switch n {
	case "mask", "mask-image", "mask-border", "mask-border-source":
		return specNo(body(""), "The striped box is unmasked: the sample is none, which is the initial value.")
	case "mask-type":
		return specNo(body(""), "The striped box is unchanged; mask-type applies to SVG mask elements, not to a box.")
	case "mask-composite":
		look := "Subtracting the fading layer from the solid layer changes the box's opacity along its width; the picture differs from the control's fully opaque stripes."
		if strings.Contains(r.Sample, "source-out") {
			look = "The control's stripes are fully opaque. With source-out the box fades: the red left band is nearly transparent, the yellow and green middle bands stay the most opaque, and the blue right band is faint again."
		}

		return spec(body("mask-image:linear-gradient(#000,#000),linear-gradient(90deg,#000,transparent);"), look)
	case "mask-mode":
		return spec(body(maskBase), "The mask is read as alpha, so the box fades toward the bottom.")
	case "mask-clip", "mask-origin":
		return specNo(body("padding:12px;border:4px solid #2c3e50;"), "The painter declines this mask value, so the page draws the striped box unmasked; the readout still comes from the engine.")
	case "mask-position":
		return specNo(body(""), "The painter declines the four-value position, so the page draws the striped box unmasked; the readout still comes from the engine.")
	case "mask-repeat":
		return spec(body("mask-image:linear-gradient(#000,#000);mask-size:30px 20px;"), "The mask tile repeats along the x axis only, so just a 20 px band along the top of the box stays visible and the rest is masked away.")
	case "mask-size":
		return specNo(body(""), "The painter declines mask-size cover, so the page draws the striped box unmasked; the readout still comes from the engine.")
	}

	if strings.HasPrefix(n, "mask-border") {
		return specNo(`<div style="margin:22px 14px">`+body("border:10px solid #2c3e50;")+`</div>`, "The striped box with a masked border: "+n+" changes the border mask if the painter supports it.")
	}

	return specNo(body(maskBase), "The masked box: "+n+" is accepted but changes nothing visible.")
}

func shapeSpec(r genProp) (DemoSpec, bool) {
	switch r.Name {
	case "shape-outside":
		return spec(vizFloatBg("radial-gradient(circle 32px at 40px 40px,#c0392b,#2980b9 31px,transparent 32px)", ""), "The control's text keeps clear of the whole 80 px square around the disc (here a 64 px disc, matching the circle's radius). With the circle(40%) shape the text moves in toward the disc, so its first lines start further left.")
	case "shape-margin":
		return spec(vizFloat("shape-outside:circle(50%);border-radius:50%;"), "The control's text hugs the disc's circle. With shape-margin the text keeps 12 px more distance from it, so each line starts further right.")
	case "shape-image-threshold":
		return spec(vizFloatBg(vizGrad, "shape-outside:"+vizGrad+";"), "The gradient used as the shape is fully opaque, so at a 0.5 threshold the shape is the whole 80 px square: the text stays beside the box in both pictures.")
	}

	return specNo(vizFloat("shape-outside:circle(50%);"), "The text wraps around the disc float the same way as in the control; "+r.Name+" is accepted but not used for layout.")
}

func listSpec(r genProp) (DemoSpec, bool) {
	switch r.Name {
	case "list-style", "list-style-type":
		return spec(vizList(""), "The round discs of the control become small squares. The list also sets inside, so each marker sits at the start of its line, and the text is indented 16 px so the marker stays clear of the first letter.")
	case "list-style-position":
		return specNo(vizList(""), "The markers stay outside the text: outside is the initial value.")
	}

	return specNo(`<div style="list-style-image:url(`+vizPNG+`)">`+vizList("list-style-image:none;")+`</div>`, "The markers look the same as the control: both show plain discs, because the engine does not apply an explicit list-style-image: inherit from the wrapper's sample image.")
}

func objectSpec(r genProp) (DemoSpec, bool) {
	switch r.Name {
	case "object-fit":
		return spec(vizImg(""), "The 20x20 image stops stretching to the box and instead covers it, cropping to keep its square shape.")
	case "object-position":
		return spec(vizImg("object-fit:none;"), "The unscaled 20 px image moves to the top left corner of the box; in the control it sits at the centre.")
	}

	return spec(vizImg(""), "The image shows only a 10% inset crop of the source, so it scales up.")
}

func fxSpec(r genProp) (DemoSpec, bool) {
	n := r.Name

	switch n {
	case "accent-color":
		return spec(`<div id="{{ID}}" style="width:140px;margin:10px"><input id="{{ID}}-1" type="checkbox" checked style="{{STYLE}}"> <input id="{{ID}}-2" type="radio" checked style="{{STYLE}}"></div>`, "The checked checkbox and radio use a red accent.")
	case "backdrop-filter":
		return spec(vizOver("", "background:rgba(255,255,255,.3);"), "The stripes behind the translucent box blur.")
	case "caret", "caret-animation", "caret-shape":
		return specNo(`<div id="{{ID}}" style="width:160px;margin:10px"><input id="{{ID}}-1" value="type here" style="width:140px;{{STYLE}}"></div>`, "The text field looks the same until it has focus; the caret change shows only in a focused field.")
	case "color":
		return spec(vizBox("font-size:20px;", "Sample text"), "The text turns red.")
	case "color-scheme":
		return specNo(vizBox("", "Sample text"), "The box looks the same: color-scheme only changes the default colours of form controls and the root.")
	case "filter":
		return spec(vizWrap("background:"+vizStripes+";", "width:100px;height:50px;margin:20px 30px;background:#fff;color:#c0392b;font-size:18px;"+"", "Sharp"), "The white box and its text blur by 2 px.")
	case "isolation":
		return specNo(`<div id="{{ID}}" style="width:200px;height:90px;background:`+vizStripes+`"><div id="{{ID}}-1" style="margin:0 30px;padding:15px 0;{{STYLE}}"><div id="{{ID}}-2" style="width:100px;height:50px;background:#c0392b"></div></div></div>`,
			"The picture matches the control (isolate only matters to a blending descendant, and there is none); the engine's operation list differs.")
	case "mix-blend-mode":
		return spec(vizWrap("position:relative;background:"+vizStripes+";", "position:absolute;left:60px;top:15px;width:120px;height:50px;background:#c0392b;", ""), "The red box multiplies with the stripes behind it, so it turns dark over the yellow, green and blue bands.")
	case "opacity":
		return spec(vizOver("", "background:#c0392b;"), "The red box turns half transparent and the stripes show through.")
	case "perspective", "perspective-origin":
		return spec(`<div id="{{ID}}" style="width:200px;height:90px;background:#dfe9f3;`+persp(n)+`{{STYLE}}"><div id="{{ID}}-1" style="margin:20px 50px;width:100px;height:50px;background:#c0392b;transform:rotateY(40deg)"></div></div>`,
			persLook(n, r.Sample))
	case "rotate":
		return spec(vizOver("", "background:#c0392b;"), "The red box turns about 2 degrees clockwise (the bare 2 is read as degrees).")
	case "scale":
		return spec(vizOver("", "background:#c0392b;"), "The box shrinks to 40% of its size around its centre.")
	case "translate":
		return spec(vizOver("", "background:#c0392b;"), "The box moves 12 px to the right.")
	case "transform":
		return spec(vizOver("", "background:#c0392b;"), "The box rotates 10 degrees clockwise.")
	case "transform-origin":
		return spec(`<div style="margin:24px 0 36px">`+vizOver("", "background:#c0392b;transform:rotate(30deg);")+`</div>`, "The 30 degree rotation pivots about a point 12 px from the box's left edge (still vertically centred) instead of its centre, so the box swings to a different place.")
	case "transform-style":
		return spec(`<div id="{{ID}}" style="width:200px;height:90px;background:#dfe9f3;perspective:300px"><div id="{{ID}}-1" style="margin:20px 50px;width:100px;height:50px;transform:rotateX(30deg);{{STYLE}}"><div id="{{ID}}-2" style="width:100px;height:50px;background:#c0392b;transform:rotateY(50deg)"></div></div></div>`,
			"In the control the parent flattens the child into one plane, so the red box shows as a wide, short rectangle. With preserve-3d the two rotations combine and the box comes out narrower and taller. It reads as a flat shape, not as an obvious 3D tilt.")
	case "transform-box":
		return spec(`<svg id="{{ID}}" width="200" height="90" viewBox="0 0 200 90" style="background:#dfe9f3"><rect id="{{ID}}-1" x="60" y="20" width="80" height="50" fill="#c0392b" style="transform:rotate(25deg);{{STYLE}}"/></svg>`,
			"The rect turns about its own box instead of the SVG origin, so it spins in place.")
	}

	return specNo(vizBox("", ""), "A plain box: "+n+" has no demo here.")
}

func persLook(n, sample string) string {
	if n == "perspective" {
		return "The control shows the red box turned 40 degrees about its vertical axis with no perspective. With perspective " + sample + " (a very short distance) it is foreshortened into a narrow sliver."
	}

	return "The perspective is 200px in both. Moving the vanishing point to " + sample + " (the container's top-left corner, not its centre) makes the red box narrower and taller and shifts it right, compared with the control's wider box."
}

func persp(n string) string {
	if n == "perspective-origin" {
		return "perspective:200px;"
	}

	return ""
}
