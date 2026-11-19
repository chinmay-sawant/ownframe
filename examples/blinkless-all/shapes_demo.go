package main

import (
	"bytes"
	"context"
	"fmt"
	"html"
	"image"
	"image/color"
	"image/png"
	"strings"

	"github.com/chinmay-sawant/blinkless/css"
	xhtml "github.com/chinmay-sawant/blinkless/html"
	"github.com/chinmay-sawant/blinkless/layout"
)

// shapesHead is the CSS shared by the rows of this tab.
func shapesHead() string {
	return `<style>.res{font-family:monospace;font-size:12px;color:#0f5132;margin:2px 0}` +
		`.cap{font-family:monospace;font-size:11px;color:#555;margin-top:2px}` +
		`.cell{display:inline-block;vertical-align:top;margin:0 14px 8px 0}</style>`
}

// shapesBox is one bordered demo box, 120 by 48 px with a 4px red border, and
// a caption under it.
func shapesBox(id, style, caption string) string {
	return fmt.Sprintf(`<div class="cell"><div id="%s" style="box-sizing:border-box;width:120px;height:48px;`+
		`border:4px solid #c0392b;%s"></div><div class="cap">%s</div></div>`, id, style, html.EscapeString(caption))
}

// shapesShow draws markup first, then one readout line per id: how many
// border line pieces the engine emitted for that box, then the plain extras.
func shapesShow(markup string, ids []string, extra ...string) string {
	var b strings.Builder

	b.WriteString(markup)

	d, err := shapesLayout(context.Background(), markup)
	if err != nil {
		return markup + "<p>layout failed: " + html.EscapeString(err.Error()) + "</p>"
	}

	lines := shapesLines(d)
	for _, id := range ids {
		b.WriteString(shapesRes("%s: %d border line pieces", id, len(lines[id])))
	}

	for _, line := range extra {
		b.WriteString(shapesRes("%s", line))
	}

	return b.String()
}

const (
	shapesPlainCap = "no clip"
	shapesCut      = "20px 10px"
)

func demoBorderClip() string {
	return shapesShow(
		shapesBox("bc-plain", "", shapesPlainCap)+shapesBox("bc-all", "border-clip:24px 12px", "border-clip: 24px 12px"),
		[]string{"bc-plain", "bc-all"})
}

func demoSideClips() string {
	return shapesShow(
		shapesBox("bc-top", "border-top-clip:"+shapesCut, "border-top-clip")+
			shapesBox("bc-right", "border-right-clip:"+shapesCut, "border-right-clip")+
			shapesBox("bc-bottom", "border-bottom-clip:"+shapesCut, "border-bottom-clip")+
			shapesBox("bc-left", "border-left-clip:"+shapesCut, "border-left-clip"),
		[]string{"bc-top", "bc-right", "bc-bottom", "bc-left"})
}

func demoBlockClip() string {
	return shapesShow(shapesBox("bc-block", "border-block-clip:24px 12px", "border-block-clip: 24px 12px"),
		[]string{"bc-block"})
}

func demoBlockStartEnd() string {
	return shapesShow(
		shapesBox("bc-bstart", "border-block-start-clip:"+shapesCut, "border-block-start-clip")+
			shapesBox("bc-bend", "border-block-end-clip:"+shapesCut, "border-block-end-clip"),
		[]string{"bc-bstart", "bc-bend"})
}

func demoInlineClip() string {
	return shapesShow(shapesBox("bc-inline", "border-inline-clip:24px 12px", "border-inline-clip: 24px 12px"),
		[]string{"bc-inline"})
}

func demoInlineStartEnd() string {
	return shapesShow(
		shapesBox("bc-istart", "border-inline-start-clip:"+shapesCut, "border-inline-start-clip")+
			shapesBox("bc-iend", "border-inline-end-clip:"+shapesCut, "border-inline-end-clip"),
		[]string{"bc-istart", "bc-iend"})
}

func demoBorderLimit() string {
	markup := shapesBox("bl-plain", "", shapesPlainCap) + shapesBox("bl-left", "border-limit:left 60px", "border-limit: left 60px")

	d, err := shapesLayout(context.Background(), markup)
	if err != nil {
		return markup + "<p>layout failed</p>"
	}

	var top float64

	for _, op := range shapesLines(d)["bl-left"] {
		if op.H == 0 {
			top = op.W / shapesPtPerPx
		}
	}

	return shapesShow(markup, []string{"bl-plain", "bl-left"}, fmt.Sprintf("bl-left horizontal edge: %.0fpx of 120px", top))
}

func demoBoundary() string {
	cell := func(id, boundary string) string {
		return `<div class="cell"><div style="width:150px;height:60px;border-radius:30px;background:#eee"><div id="` + id +
			`" style="position:relative;left:0;top:0;width:50px;height:20px;border:4px solid #c0392b;` + boundary +
			`"></div></div><div class="cap">` + boundary + `</div></div>`
	}

	markup := cell("bb-none", "border-boundary:none") + cell("bb-parent", "border-boundary:parent")

	d, err := shapesLayout(context.Background(), markup)
	if err != nil {
		return markup + "<p>layout failed</p>"
	}

	var b strings.Builder

	b.WriteString(markup)

	for _, id := range []string{"bb-none", "bb-parent"} {
		_, strokes := shapesPaths(d, id)
		b.WriteString(shapesRes("%s: %d border path along the parent curve", id, strokes))
	}

	return b.String()
}

func demoBorderShape() string {
	markup := shapesBox("bs-plain", "background:#f6e3df;border-radius:24px", "border-radius: 24px, a rounded rectangle") +
		shapesBox("bs-box", "background:#f6e3df;border-shape:ellipse(50% 50%)", "border-shape: ellipse(50% 50%)")

	d, err := shapesLayout(context.Background(), markup)
	if err != nil {
		return markup + "<p>layout failed</p>"
	}

	var b strings.Builder

	b.WriteString(markup)

	for _, id := range []string{"bs-plain", "bs-box"} {
		fills, strokes := shapesPaths(d, id)
		b.WriteString(shapesRes("%s: %d fill path, %d border path", id, fills, strokes))
	}

	return b.String()
}

func demoTextFit() string {
	row := func(id, decl string) string {
		return fmt.Sprintf(`<div class="cell"><div id="%s" style="width:200px;font-size:12px;background:#f1efe7;%s">`+
			`Fit this line</div><div class="cap">%s</div></div>`, id, decl, html.EscapeString(decl+" "))
	}

	markup := row("tf-plain", "") + row("tf-grow", "text-fit:grow") + row("tf-cap", "text-fit:grow consistent 150%") +
		`<div class="cell"><div id="tf-shrink" style="width:90px;font-size:12px;background:#f1efe7;text-fit:shrink">` +
		`Shrink this longer line</div><div class="cap">text-fit:shrink, 90px box</div></div>`

	d, err := shapesLayout(context.Background(), markup)
	if err != nil {
		return markup + "<p>layout failed</p>"
	}

	var b strings.Builder

	b.WriteString(markup)

	for _, id := range []string{"tf-plain", "tf-grow", "tf-cap", "tf-shrink"} {
		b.WriteString(shapesRes("%s: font %.1fpx", id, shapesTextSize(d, id)/shapesPtPerPx))
	}

	return b.String()
}

func demoFloatDefer() string {
	cell := func(id, decl string) string {
		return `<div class="cell"><div style="column-count:3;column-gap:10px;width:320px;background:#eef2f7"><div id="` + id +
			`" style="float:left;width:20px;height:20px;margin-right:8px;background:#9db4d6;` + decl + `"></div><p style="margin:0">one two three</p></div>` +
			`<div class="cap">` + decl + `</div></div>`
	}

	markup := cell("fd-none", "float-defer:none") + cell("fd-one", "float-defer:1") + cell("fd-last", "float-defer:last")

	d, err := shapesLayout(context.Background(), markup)
	if err != nil {
		return markup + "<p>layout failed</p>"
	}

	var b strings.Builder

	b.WriteString(markup)

	for _, box := range d.Boxes {
		if strings.HasPrefix(box.ID, "fd-") {
			b.WriteString(shapesRes("%s: %.0fpx from the left edge of its box", box.ID, box.X))
		}
	}

	return b.String()
}

func demoImageThreshold() string {
	cell := func(id, threshold string) string {
		return `<div class="cell"><div style="width:260px;background:#eef7ee;overflow:hidden"><div style="float:left;width:80px;` +
			`height:80px;shape-outside:url(mask);shape-image-threshold:` + threshold + `;background:linear-gradient(90deg,#8fc79b 50%,#e6f2e8 50%)"></div>` +
			`<div id="` + id + `">Text beside an image shape</div></div>` +
			`<div class="cap">shape-image-threshold: ` + threshold + `</div></div>`
	}

	markup := cell("si-low", "0") + cell("si-high", "0.5")

	d, err := shapesLayout(context.Background(), markup)
	if err != nil {
		return markup + "<p>layout failed</p>"
	}

	var b strings.Builder

	b.WriteString(markup)

	for _, id := range []string{"si-low", "si-high"} {
		b.WriteString(shapesRes("%s: text starts %.0fpx from the float's left edge", id, shapesTextX(d, id)/shapesPtPerPx-shapesCellLeft(d, id)))
	}

	return b.String()
}

func shapesColumns(vis string) string {
	return fmt.Sprintf(`<div class="cell"><div id="cr-%s" style="width:230px;column-count:2;column-gap:20px;`+
		`column-rule:3px solid #c0392b;column-rule-visibility-items:%s;font-size:11px;text-align:center;padding:0 14px">`+
		`one two three four five six seven eight nine ten eleven</div><div class="cap">%s</div></div>`, vis, vis, vis)
}

func demoColumnRule() string {
	markup := shapesColumns("between") + shapesColumns("around") + shapesColumns("all") + shapesColumns("normal")

	d, err := shapesLayout(context.Background(), markup)
	if err != nil {
		return markup + "<p>layout failed</p>"
	}

	var b strings.Builder

	b.WriteString(markup)

	lines := shapesLines(d)

	for _, v := range []string{"between", "around", "all", "normal"} {
		n := 0

		for _, op := range lines["cr-"+v] {
			if op.W == 0 && op.H > 0 {
				n++
			}
		}

		b.WriteString(shapesRes("%s: %d vertical rules", v, n))
	}

	return b.String()
}

// shapesLayout lays out the demo markup and returns the display list.
func shapesLayout(ctx context.Context, markup string) (*layout.Display, error) {
	doc, err := xhtml.Parse([]byte(`<!DOCTYPE html><html><body style="margin:0;font-size:14px">` + markup + `</body></html>`))
	if err != nil {
		return nil, err
	}

	styled, err := css.Apply(ctx, doc, css.Options{WidthPx: shapesPageW, HeightPx: shapesPageH, Media: "screen"})
	if err != nil {
		return nil, err
	}

	return layout.DisplayListOptions(ctx, styled, layout.Options{Images: shapesMask})
}

// shapesMask answers the url(mask) source with a 20 by 20 PNG: the left half
// is opaque and the right half has alpha 0.3.
func shapesMask(string) ([]byte, error) {
	img := image.NewNRGBA(image.Rect(0, 0, maskSizePx, maskSizePx))

	for y := range 20 {
		for x := range 20 {
			alpha := uint8(opaqueAlpha)
			if x >= maskHalf {
				alpha = 77
			}

			img.SetNRGBA(x, y, color.NRGBA{R: 0, G: 128, B: 0, A: alpha}) //nolint:mnd // mask color, RGBA table
		}
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// shapesPaths counts the fill and border path ops a box owns.
func shapesPaths(d *layout.Display, id string) (int, int) {
	fills, strokes := 0, 0

	for i, op := range d.Ops {
		if op.Kind != layout.DisplayOpPath || i >= len(d.OpOwners) || d.OpOwners[i] < 0 || d.Boxes[d.OpOwners[i]].ID != id {
			continue
		}

		if path, ok := op.CornerPathValue(); ok && path.Stroke {
			strokes++
		} else {
			fills++
		}
	}

	return fills, strokes
}

// shapesTextOp returns the first text op a box owns.
func shapesTextOp(d *layout.Display, id string) (layout.DisplayOp, bool) {
	for i, op := range d.Ops {
		if op.Kind == layout.DisplayOpText && i < len(d.OpOwners) && d.OpOwners[i] >= 0 && d.Boxes[d.OpOwners[i]].ID == id {
			return op, true
		}
	}

	return layout.DisplayOp{}, false
}

// shapesTextSize is the font size, in points, of a box's first text op.
func shapesTextSize(d *layout.Display, id string) float64 {
	op, _ := shapesTextOp(d, id)

	return op.Size
}

// shapesTextX is the x, in points, of a box's first text op.
func shapesTextX(d *layout.Display, id string) float64 {
	op, _ := shapesTextOp(d, id)

	return op.X
}

// shapesCellLeft is the left edge, in CSS pixels, of the float's row: the
// x of the text box's own left edge when no float narrows it.
func shapesCellLeft(d *layout.Display, id string) float64 {
	for _, box := range d.Boxes {
		if box.ID == id {
			return box.X
		}
	}

	return 0
}

// shapesLines returns the stroked line ops each box owns, keyed by element id.
// Lines come back in points as the engine stores them.
func shapesLines(d *layout.Display) map[string][]layout.DisplayOp {
	out := map[string][]layout.DisplayOp{}

	for i, op := range d.Ops {
		if op.Kind != layout.DisplayOpLine || i >= len(d.OpOwners) || d.OpOwners[i] < 0 {
			continue
		}

		id := d.Boxes[d.OpOwners[i]].ID
		out[id] = append(out[id], op)
	}

	return out
}

// shapesRes is one readout line under a showcase.
func shapesRes(format string, args ...any) string {
	return `<div class="res">` + html.EscapeString(fmt.Sprintf(format, args...)) + `</div>`
}

// Numbers of the shapes demo: points per CSS pixel and the page width.
const (
	shapesPtPerPx = 0.75
	shapesPageW   = 900
	shapesPageH   = 600
)

// Mask picture for the url(mask) source: its side in pixels, the column where
// the alpha drops, and the opaque alpha value.
const (
	maskSizePx  = 20
	maskHalf    = 10
	opaqueAlpha = 255
)
