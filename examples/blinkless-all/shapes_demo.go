package main

import (
	"context"
	"fmt"
	"html"
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

// shapesStored draws markup and reports that the named boxes only store the
// property: the Declared value is read back and nothing is drawn for it.
func shapesStored(markup string, ids ...string) string {
	d, err := shapesLayout(context.Background(), markup)
	if err != nil {
		return markup + "<p>layout failed: " + html.EscapeString(err.Error()) + "</p>"
	}

	var b strings.Builder

	b.WriteString(markup)

	for _, box := range d.Boxes {
		for _, id := range ids {
			if box.ID != id {
				continue
			}

			for name, v := range box.Declared {
				b.WriteString(shapesRes("%s: %s stored as %q, stored only, not drawn", id, name, v))
			}
		}
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
	return shapesStored(`<div class="cell"><div id="bb-parent" style="width:150px;height:60px;border:4px solid #888;`+
		`border-radius:50%"><div id="bb-child" style="width:60px;height:30px;border:4px solid #c0392b;`+
		`border-boundary:parent"></div></div><div class="cap">child with border-boundary: parent</div></div>`, "bb-child")
}

func demoBorderShape() string {
	return shapesStored(shapesBox("bs-box", "border-shape:ellipse(50% 50%);border-radius:24px",
		"border-shape: ellipse(50% 50%), still a rounded rectangle"), "bs-box")
}

func demoTextFit() string {
	return shapesStored(`<div class="cell"><div id="tf-box" style="width:200px;font-size:12px;background:#f1efe7;`+
		`text-fit:grow consistent 150%">Fit this line</div><div class="cap">text-fit: grow consistent 150%, text stays 12px</div></div>`, "tf-box")
}

func demoFloatDefer() string {
	return shapesStored(`<div style="width:260px;background:#eef2f7;overflow:hidden"><div id="fd-float" style="float:left;width:60px;`+
		`height:30px;background:#9db4d6;float-defer:1">defer 1</div><div id="fd-text">Text beside a float with float-defer 1.</div></div>`, "fd-float")
}

func demoImageThreshold() string {
	return shapesStored(`<div style="width:260px;background:#eef7ee;overflow:hidden"><div id="si-float" style="float:left;width:60px;`+
		`height:60px;shape-outside:circle(50%);shape-image-threshold:0.5;background:#9dd6a8">0.5</div>`+
		`<div id="si-text">Text beside a circle float with an image threshold.</div></div>`, "si-float")
}

func shapesColumns(vis string) string {
	return fmt.Sprintf(`<div class="cell"><div id="cr-%s" style="width:200px;column-count:2;column-gap:20px;`+
		`column-rule:3px solid #c0392b;column-rule-visibility-items:%s;font-size:11px;padding:0 4px">`+
		`one two three four five six seven eight nine ten eleven twelve</div><div class="cap">%s</div></div>`, vis, vis, vis)
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

	return layout.DisplayList(ctx, styled)
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
