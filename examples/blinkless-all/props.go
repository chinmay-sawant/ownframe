package main

import (
	"context"
	"fmt"
	"html"
	"reflect"
	"runtime"
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
// under the bitmap limit even when a tab holds nested blend/isolation groups.
const genPageRows = 16

// genGroups lists the generated tabs in order. Their ids start with "all-".
//
//nolint:gochecknoglobals // static tab group table
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
			hint := grp.Hint
			props := make([]Prop, 0, len(part))

			for _, r := range part {
				props = append(props, Prop{
					Name:        r.Name,
					Note:        fmt.Sprintf("The right box sets %s: %s.", r.Name, r.Sample),
					Description: genDescription(r),
					ValidateFn:  func() string { set.once.Do(set.compute); return genValidate(r, hint) },
					Demo:        func() string { return set.demo(r) },
				})
			}

			id, title := "all-"+grp.ID, grp.Title
			if pages > 1 {
				id = fmt.Sprintf("%s-%d", id, pg+1)
				title = fmt.Sprintf("%s %d/%d", title, pg+1, pages)
			}

			out = append(out, Tab{
				ID:      id,
				Title:   title,
				Props:   props,
				Head:    genHead,
				Prepare: func() { set.once.Do(set.compute) },
			})
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

	badge := "The status badge must read implemented."
	if declines(r.Name) {
		badge = "The status badge reads partial: the window painter does not draw this property fully, or the demo shows no visible effect, so the picture alone does not prove it."
	}

	return fmt.Sprintf("The left demo is the control. The right demo adds `%s: %s`. %s The readout lists each measured element in the control and with the property, and says whether the layout or the paint operations differ. %s", r.Name, r.Sample, look, badge)
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
		base := "font-size:30px;font-weight:bold;color:#2c3e50;padding:4px 8px;margin:8px 0;"
		look := "Look for an outline drawn around each glyph of the word."
		visible := false
		if r.Name == "-webkit-text-stroke-color" {
			base += "-webkit-text-stroke-width:2px;"
			look = "The 2 px glyph outline turns red instead of inheriting the dark slate text colour."
			visible = true
		}

		return DemoSpec{
			HTML:    `<div id="{{ID}}" style="` + base + `{{STYLE}}">Stroke</div>`,
			Look:    look,
			Visible: visible,
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
}

// opSig is the whole paint of one operation without its id: every field,
// including paths, colours and pointed-to values, so two operations that
// paint differently never compare equal.
func opSig(op layout.DisplayOp) string {
	op.ID = 0

	v := reflect.New(reflect.TypeOf(op)).Elem()
	v.Set(reflect.ValueOf(op))

	return fmt.Sprintf("%+v", deepValue(v))
}

// deepValue follows pointers and slices so a printed operation shows what they
// point at, not an address.
func deepValue(v reflect.Value) any {
	switch v.Kind() {
	case reflect.Pointer, reflect.Interface:
		if v.IsNil() {
			return nil
		}

		return deepValue(v.Elem())
	case reflect.Struct:
		out := make([]any, 0, v.NumField())

		for i := range v.NumField() {
			f := v.Field(i)
			if v.Type().Field(i).Name == "Font" && f.Kind() == reflect.Pointer {
				// A font face is a large shared object; two operations use the
				// same face when they hold the same pointer.
				out = append(out, f.Pointer())

				continue
			}

			out = append(out, deepValue(f))
		}

		return out
	case reflect.Slice, reflect.Array:
		out := make([]any, 0, v.Len())
		for i := range v.Len() {
			out = append(out, deepValue(v.Index(i)))
		}

		return out
	case reflect.Map, reflect.Func, reflect.Chan, reflect.UnsafePointer:
		return nil
	}

	if v.CanInterface() {
		return v.Interface()
	}

	switch v.Kind() {
	case reflect.Bool:
		return v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int()
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint()
	case reflect.Float32, reflect.Float64:
		return v.Float()
	case reflect.String:
		return v.String()
	case reflect.Complex64, reflect.Complex128:
		return v.Complex()
	}

	return nil
}

// genCellW is the width of one demo cell on the page, so the readout lays the
// demo out at the size the picture does.
const genCellW = 300

// genMeasure lays one demo out alone in a cell and returns its boxes. Each
// layout starts at the origin, so the control and the demo compare by absolute
// geometry and by every field of every paint operation, paths and paint
// included.
func genMeasure(r genProp, prefix string, withStyle bool) (boxes []genBox, root layout.Box, err error) {
	spec := specFor(r)
	id := genID(prefix, r.Name)

	d, err := shapesLayout(context.Background(), fmt.Sprintf(`<div style="width:%dpx">%s</div>`, genCellW, spec.expand(r, id, withStyle)))
	if err != nil {
		return nil, root, err
	}

	opByID := map[uint64]layout.DisplayOp{}
	for _, op := range d.Ops {
		opByID[op.ID] = op
	}

	// SVG shapes and other paint that no box owns still belong to the demo:
	// they are added to the root element's operations, so a changed fill or
	// stroke inside an <svg> shows in the readout.
	owned := map[uint64]bool{}

	for _, box := range d.Boxes {
		for _, opID := range box.OperationIDs {
			owned[opID] = true
		}
	}

	var loose []string

	for _, op := range d.Ops {
		if !owned[op.ID] {
			loose = append(loose, opSig(op))
		}
	}

	for _, box := range d.Boxes {
		if box.ID == id {
			root = box
		}

		if box.ID != id && !strings.HasPrefix(box.ID, id+"-") {
			continue
		}

		gb := genBox{x: box.X, y: box.Y, w: box.W, h: box.H}
		for _, opID := range box.OperationIDs {
			gb.ops = append(gb.ops, opSig(opByID[opID]))
		}

		if box.ID == id {
			gb.ops = append(gb.ops, loose...)
		}

		boxes = append(boxes, gb)
	}

	return boxes, root, nil
}

// readoutCache keeps the readout of each property for the life of the
// process: a row's readout depends only on the row, and each generated tab
// computes its rows lazily on first render, so each row is measured at most once.
//
//nolint:gochecknoglobals // process-lifetime memoization cache
var readoutCache sync.Map

// compute fills the readout of every row of the tab. The rows are independent
// layouts, so they run in parallel.
func (g *genSet) compute() {
	g.res = make(map[string]string, len(g.rows))

	var (
		mu sync.Mutex
		wg sync.WaitGroup
	)

	limit := make(chan struct{}, runtime.NumCPU())

	for _, r := range g.rows {
		if v, ok := readoutCache.Load(r.Name); ok {
			g.res[r.Name], _ = v.(string)

			continue
		}

		wg.Add(1)

		limit <- struct{}{}

		go func() {
			defer func() { <-limit; wg.Done() }()

			out := measureReadout(r)
			readoutCache.Store(r.Name, out)

			mu.Lock()
			g.res[r.Name] = out
			mu.Unlock()
		}()
	}

	wg.Wait()
}

// measureReadout lays the control and the demo of one row out and compares them.
func measureReadout(r genProp) string {
	ctl, _, err := genMeasure(r, "c_", false)
	if err != nil {
		return "layout failed: " + err.Error()
	}

	tst, root, err := genMeasure(r, "t_", true)
	if err != nil {
		return "layout failed: " + err.Error()
	}

	return genReadout(r, ctl, tst, root)
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

		if opOnly[r.Name] {
			verdict = "the engine's paint operation list differs from the control, but the window picture matches it"
		}
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

	out := g.res[r.Name]
	if note != "" && !strings.HasSuffix(out, ".") {
		out += "."
	}

	return g.pageMarkup(r) + shapesRes("%s%s", out, note)
}

// noPaint lists the properties the window painter cannot draw. A page that
// carries one falls back to a bitmap the example refuses, so the visible demo
// leaves the declaration out and the readout still comes from the layout.
//
//nolint:gochecknoglobals // read-only property lookup table
var noPaint = map[string]bool{}
