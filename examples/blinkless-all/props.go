package main

import (
	"context"
	"fmt"
	"html"
	"reflect"
	"sort"
	"strings"
	"sync"

	"github.com/chinmay-sawant/blinkless/layout"
)

// genProp is one row of the generated property table: the property, the
// accepted values from the compatibility matrix, the sample value the demo
// sets, and the tab group it belongs to.
type genProp struct {
	Name, Accepted, Sample, Group string
}

// genPageRows caps the rows on one generated tab, so a tab's picture stays
// under the bitmap limit.
const genPageRows = 20

// genGroups lists the generated tabs in order. Their ids start with "all-".
var genGroups = []struct{ ID, Title, Hint string }{
	{"layout", "Layout and box", "Look for a change in the size, position or spacing of the right box."},
	{"border", "Borders, corners, rules", "Look for a changed edge, corner or rule line on the right box."},
	{"text", "Text and fonts", "Look for changed glyph shape, spacing, wrapping or decoration in the right box's text."},
	{"anim", "Animation and timelines", "These declare motion. The right box carries the declaration; motion itself is shown on the Motion tab."},
	{"paint", "Background, mask, shape", "Look for a changed fill, clip or mask on the right box."},
	{"fx", "Effects and transforms", "Look for a changed blend, filter, opacity, transform or colour on the right box."},
	{"svg", "SVG paint", "These apply to SVG shapes; the small svg beside the boxes is control rect, then a rect that carries the declaration."},
	{"webkit", "-webkit- aliases", "Each alias must behave like its unprefixed property."},
	{"misc", "Other", "Look for the effect the description names; some only export a value to the host."},
}

// genTabs builds one tab per group from the generated table, leaving out the
// properties the hand-built tabs already cover.
func genTabs(taken map[string]bool) []Tab {
	byGroup := map[string][]genProp{}

	for _, g := range genProps() {
		if !taken[g.Name] {
			byGroup[g.Group] = append(byGroup[g.Group], g)
		}
	}

	var out []Tab

	for _, grp := range genGroups {
		rows := byGroup[grp.ID]
		if len(rows) == 0 {
			continue
		}

		sort.Slice(rows, func(i, j int) bool { return rows[i].Name < rows[j].Name })

		pages := (len(rows) + genPageRows - 1) / genPageRows

		for pg := range pages {
			part := rows[pg*len(rows)/pages : (pg+1)*len(rows)/pages]
			set := newGenSet(part)
			props := make([]Prop, 0, len(part))

			for _, r := range part {
				props = append(props, Prop{
					Name:        r.Name,
					Note:        fmt.Sprintf("The right box sets %s: %s.", r.Name, r.Sample),
					Description: genDescription(r),
					Validate:    genValidate(r, grp.Hint),
					Demo:        func() string { return set.demo(r) },
				})
			}

			id, title := "all-"+grp.ID, grp.Title
			if pages > 1 {
				id = fmt.Sprintf("%s-%d", id, pg+1)
				title = fmt.Sprintf("%s %d/%d", title, pg+1, pages)
			}

			out = append(out, Tab{ID: id, Title: title, Props: props, Head: genHead})
		}
	}

	return out
}

func genDescription(r genProp) string {
	return fmt.Sprintf("%s. Accepted values: %s.", r.Name, r.Accepted)
}

func genValidate(r genProp, hint string) string {
	spec := specFor(r)

	look := spec.Look
	if look == "" {
		look = hint
	}

	return fmt.Sprintf("The left demo is the control. The right demo adds `%s: %s`. %s The readout lists each measured element in the control and with the property, and says whether the layout or the paint operations differ. The status badge must read implemented.", r.Name, r.Sample, look)
}

func genHead() string {
	return `<style>.gcell{display:inline-block;vertical-align:top;margin:0 12px 6px 0;width:300px}` +
		`.gcap{font-size:11px;color:#6b7280;margin:0 0 3px 0}` +
		`.gbox{padding:6px 8px;background:#f1efe7;border:1px solid #bbb;font-size:12px;min-height:20px}</style>`
}

// DemoSpec is the demo of one property. HTML is a fragment that holds the
// measured elements and the place the property goes. It uses these markers:
//
//	{{ID}}     the id prefix. The root measured element has id="{{ID}}" and any
//	           other measured element has an id that starts with "{{ID}}-".
//	{{STYLE}}  replaced by "" in the control and by "name:value;" in the demo,
//	           so put it inside a style attribute, after the fixed styles.
//	{{NAME}}   the property name. {{VALUE}} the sample value.
//
// The framework lays the control and the demo out, measures every element whose
// id starts with the prefix, and fills the readout. Look says in one sentence
// what to see. Visible means the property must change the layout or the paint
// operations of the measured elements; a test checks it.
type DemoSpec struct {
	HTML    string
	Look    string
	Visible bool
}

// specFor returns the property's demo. Each group file answers for its own
// properties; anything unclaimed gets a plain box.
func specFor(r genProp) DemoSpec {
	if strings.HasPrefix(r.Name, "-webkit-text-stroke") {
		return DemoSpec{
			HTML:    `<div id="{{ID}}" style="font-size:30px;font-weight:bold;color:#2c3e50;padding:4px 8px;{{STYLE}}">Stroke</div>`,
			Look:    "Look for an outline drawn around each glyph of the word.",
			Visible: false,
		}
	}

	for _, f := range []func(genProp) (DemoSpec, bool){layoutSpec, textSpec, visualSpec} {
		if spec, ok := f(r); ok {
			return spec
		}
	}

	// A -webkit- alias behaves like its unprefixed property, so it reuses that
	// property's demo; the demo still sets the prefixed name.
	if base := strings.TrimPrefix(r.Name, "-webkit-"); base != r.Name {
		base = strings.Replace(base, "mask-box-image", "mask-border", 1)
		alias := r
		alias.Name = base

		for _, g := range genProps() {
			if g.Name == base {
				alias.Group = g.Group
				alias.Accepted = g.Accepted

				break
			}
		}

		for _, f := range []func(genProp) (DemoSpec, bool){layoutSpec, textSpec, visualSpec} {
			if spec, ok := f(alias); ok {
				return spec
			}
		}
	}

	return DemoSpec{HTML: defaultSpecHTML}
}

const defaultSpecHTML = `<div id="{{ID}}" class="gbox" style="{{STYLE}}">{{NAME}}: {{VALUE}}</div>`

func (s DemoSpec) expand(r genProp, id string, withStyle bool) string {
	style := ""
	if withStyle {
		style = r.Name + ":" + r.Sample + ";"
	}

	return strings.NewReplacer("{{ID}}", id, "{{STYLE}}", html.EscapeString(style),
		"{{NAME}}", html.EscapeString(r.Name), "{{VALUE}}", html.EscapeString(r.Sample)).Replace(s.HTML)
}

// genSet lays out every demo of one tab in one pass, so the page reads the
// control and the demo geometry of all its rows from a single layout.
type genSet struct {
	rows []genProp
	once sync.Once
	res  map[string]string
}

func newGenSet(rows []genProp) *genSet { return &genSet{rows: rows} }

func genID(prefix, name string) string {
	return prefix + strings.NewReplacer("-", "_", ".", "_").Replace(name)
}

// pageMarkup is what the page draws for the row: the control and the demo.
func (g *genSet) pageMarkup(r genProp) string {
	spec := specFor(r)
	with := !noPaint[r.Name]

	return `<div class="gcell"><div class="gcap">control</div>` + spec.expand(r, genID("pc_", r.Name), false) + `</div>` +
		`<div class="gcell"><div class="gcap">` + html.EscapeString(r.Name+": "+r.Sample) + `</div>` + spec.expand(r, genID("pt_", r.Name), with) + `</div>`
}

type genBox struct {
	x, y, w, h float64
	ops        []string
	declared   string
}

// opSig is the paint of one operation without its id or pointers.
func opSig(op layout.DisplayOp, ox, oy float64) string {
	op.ID = 0
	op.X -= ox
	op.Y -= oy
	v := reflect.ValueOf(op)

	var b strings.Builder

	for i := range v.NumField() {
		f := v.Field(i)
		if !f.CanInterface() {
			continue
		}

		switch f.Kind() {
		case reflect.Pointer, reflect.Interface, reflect.Map, reflect.Slice, reflect.Func, reflect.Chan:
			continue
		}

		fmt.Fprintf(&b, "%v,", f.Interface())
	}

	return b.String()
}

func (g *genSet) compute() {
	g.res = map[string]string{}

	var b strings.Builder

	for _, r := range g.rows {
		spec := specFor(r)
		b.WriteString(`<div>` + spec.expand(r, genID("c_", r.Name), false) + spec.expand(r, genID("t_", r.Name), true) + `</div>`)
	}

	d, err := shapesLayout(context.Background(), b.String())
	if err != nil {
		for _, r := range g.rows {
			g.res[r.Name] = "layout failed: " + err.Error()
		}

		return
	}

	opByID := map[uint64]layout.DisplayOp{}
	for _, op := range d.Ops {
		opByID[op.ID] = op
	}

	// boxes keyed by element id, with the geometry made relative to the root
	// measured element of the same prefix.
	byID := map[string]layout.Box{}
	for _, box := range d.Boxes {
		if box.ID != "" {
			byID[box.ID] = box
		}
	}

	collect := func(prefix string) []genBox {
		root, ok := byID[prefix]
		if !ok {
			return nil
		}

		var out []genBox

		// Operation positions are made relative to the top-left of the
		// prefix's operations, so a control and a demo that sit at different
		// places in the page still compare equal.
		ox, oy, seen := 0.0, 0.0, false

		for _, box := range d.Boxes {
			if box.ID != prefix && !strings.HasPrefix(box.ID, prefix+"-") {
				continue
			}

			for _, id := range box.OperationIDs {
				op := opByID[id]
				if !seen || op.X < ox {
					ox = op.X
				}

				if !seen || op.Y < oy {
					oy = op.Y
				}

				seen = true
			}
		}

		for _, box := range d.Boxes {
			if box.ID != prefix && !strings.HasPrefix(box.ID, prefix+"-") {
				continue
			}

			gb := genBox{x: box.X - root.X, y: box.Y - root.Y, w: box.W, h: box.H}
			for _, id := range box.OperationIDs {
				gb.ops = append(gb.ops, opSig(opByID[id], ox, oy))
			}

			if v := box.Declared[""]; v != "" {
				gb.declared = v
			}

			out = append(out, gb)
		}

		return out
	}

	for _, r := range g.rows {
		g.res[r.Name] = genReadout(r, collect(genID("c_", r.Name)), collect(genID("t_", r.Name)), byID[genID("t_", r.Name)])
	}
}

func genReadout(r genProp, ctl, tst []genBox, root layout.Box) string {
	if len(tst) == 0 {
		return "the demo element is not in the display list"
	}

	differs := len(ctl) != len(tst)

	var lines []string

	for i, t := range tst {
		var c genBox
		if i < len(ctl) {
			c = ctl[i]
		}

		same := c.x == t.x && c.y == t.y && c.w == t.w && c.h == t.h && strings.Join(c.ops, "|") == strings.Join(t.ops, "|")
		if !same {
			differs = true
		}

		mark := "="
		if !same {
			mark = "!="
		}

		lines = append(lines, fmt.Sprintf("element %d: control %.0fx%.0f at %.0f,%.0f with %d ops %s demo %.0fx%.0f at %.0f,%.0f with %d ops",
			i, c.w, c.h, c.x, c.y, len(c.ops), mark, t.w, t.h, t.x, t.y, len(t.ops)))
	}

	verdict := "same layout and paint as the control"
	if differs {
		verdict = "layout or paint differs from the control"
	}

	if v, ok := root.Declared[r.Name]; ok {
		verdict += fmt.Sprintf("; exported to the host as %q", v)
	}

	return strings.Join(lines, " | ") + " => " + verdict
}

func (g *genSet) demo(r genProp) string {
	g.once.Do(g.compute)

	note := ""
	if noPaint[r.Name] {
		note = " The window painter declines this property (it falls back to a bitmap or reports unsupported text shaping options), so the demo is drawn without the declaration; the readout comes from the engine's layout."
	}

	return g.pageMarkup(r) + shapesRes("%s%s", g.res[r.Name], note)
}

// noPaint lists the properties the window painter cannot draw. A page that
// carries one falls back to a bitmap the example refuses, so the visible demo
// leaves the declaration out and the readout still comes from the layout.
var noPaint = map[string]bool{
	"font-feature-settings": true, "font-variant": true, "font-variant-alternates": true, "font-variant-caps": true,
	"font-variant-east-asian": true, "font-variant-ligatures": true, "text-autospace": true, "writing-mode": true,
	"font-variant-numeric": true, "font-variant-position": true,
	"mask-clip": true, "mask-origin": true, "mask-position": true, "mask-size": true,
	"-webkit-mask-clip": true, "-webkit-mask-origin": true, "-webkit-mask-position": true, "-webkit-mask-size": true,
}
