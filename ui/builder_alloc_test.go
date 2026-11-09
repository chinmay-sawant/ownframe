package ui

import (
	"context"
	"fmt"
	"hash/fnv"
	"html"
	"sort"
	"strconv"
	"strings"
	"testing"
)

const (
	allocSections     = 25
	allocItemsPerList = 10
	allocTableRows    = 10
	allocTableCols    = 4
	allocNavLinks     = 6
	allocDepthChain   = 12
	allocRenderRuns   = 50
	allocIDWindow     = 400

	allocInvoiceRows        = 1000
	allocInvoiceRuns        = 10
	allocInvoiceNumWidth    = 56
	allocInvoiceIDWidth     = 120
	allocInvoiceClientWidth = 160
	allocInvoiceDateWidth   = 96
	allocInvoiceItemWidth   = 180
	allocInvoiceQtyWidth    = 48
	allocInvoiceTotalWidth  = 80
	allocInvoiceStatusWidth = 70
)

// allocBenchTree builds a deterministic multi-hundred-element document that
// exercises every render path: nested sections, sorted attributes,
// synthesized data-action/id/style attributes, escaped text, void elements,
// single-attribute and single-style nodes, a deep chain, and a shared child.
func allocBenchTree() *Element {
	root := Div().ID("alloc-root").Class("page").
		BackgroundColor(Hex("#101418")).Padding(Px(12))
	nav := Nav().Class("topnav").Attr("aria-label", "Sections & Pages <nav>")
	for i := 0; i < allocNavLinks; i++ {
		nav.Children(
			Tag("a").Attr("href", fmt.Sprintf("#section-%d", i)).
				Text(fmt.Sprintf("Item %d & <More> {{.Link}}", i)),
		)
	}
	root.Children(nav)
	for s := 0; s < allocSections; s++ {
		card := Section().Class("card").ID(fmt.Sprintf("section-%d", s)).
			Attr("data-index", fmt.Sprintf("%d", s)).
			BackgroundColor(Hex("#1b2230")).BorderRadius(Px(8))
		card.Children(
			Tag("h2").Text(fmt.Sprintf("Section %d & <Details> {{.Title}}", s)),
			P().Text("Body {{.Body}} with \"quotes\" & 'apostrophes' {braces}"),
		)
		list := Ul().Class("rows")
		for i := 0; i < allocItemsPerList; i++ {
			item := Li()
			if i%2 == 0 {
				item.Class("even")
			} else {
				item.Class("odd").Attr("title", fmt.Sprintf("row %d \"quoted\" & <brackets>", i))
			}
			list.Children(item.Text(fmt.Sprintf("row-%d value <x> & \"y\" {{.Row}}", i)))
		}
		card.Children(list)
		toolbar := Div().Class("toolbar").Display(DisplayFlex).
			FlexDirection(FlexDirectionRow).AlignItems(AlignItemsCenter).Gap(Px(8))
		toolbar.Children(
			Button(fmt.Sprintf("Save %d", s)).ID(fmt.Sprintf("save-%d", s)).
				OnClick(func(context.Context) error { return nil }),
			Button("Cancel").OnClick(func(context.Context) error { return nil }),
			Input().Attr("name", fmt.Sprintf("field-%d", s)).
				Attr("value", "edit \"me\" & you {{.Field}}"),
		)
		card.Children(toolbar)
		card.Children(Span().Width(Px(7)).Text("meter"))
		root.Children(card)
	}
	table := Tag("table").Class("grid")
	head := Tag("tr")
	for c := 0; c < allocTableCols; c++ {
		head.Children(Tag("th").Text(fmt.Sprintf("Col %d & <H> {{.H}}", c)))
	}
	table.Children(Tag("thead").Children(head))
	body := Tag("tbody")
	for r := 0; r < allocTableRows; r++ {
		row := Tag("tr").Class("grid-row")
		for c := 0; c < allocTableCols; c++ {
			row.Children(Tag("td").Text(fmt.Sprintf("r%d-c%d <v> & \"w\" {{.Cell}}", r, c)))
		}
		body.Children(row)
	}
	table.Children(body)
	root.Children(table)
	form := Tag("form").Attr("action", "/submit?x=1&y=2")
	for i := 0; i < 3; i++ {
		form.Children(
			Tag("label").Attr("for", fmt.Sprintf("opt-%d", i)).Text(fmt.Sprintf("Pick %d {{.Opt}}", i)),
			Input().Attr("name", fmt.Sprintf("opt-%d", i)).Attr("value", "a&b"),
		)
	}
	selectBox := Tag("select").Attr("name", "choice")
	for i := 0; i < 3; i++ {
		selectBox.Children(Tag("option").Attr("value", fmt.Sprintf("%d", i)).Text(fmt.Sprintf("Option %d", i)))
	}
	form.Children(selectBox, Tag("textarea").Attr("name", "notes").Text("notes <here> & {{.Notes}}"))
	root.Children(form)
	root.Children(
		Div().Attr("data-action", "jump").Text("Jump {{.Target}} & <away>"),
		Div().Attr("data-action", "stay").ID("stay-put").Text("Stay"),
	)
	deep := Div().Class("depth-0")
	tip := deep
	for d := 1; d < allocDepthChain; d++ {
		next := Div().Class(fmt.Sprintf("depth-%d", d))
		tip.Children(next)
		tip = next
	}
	tip.Children(Span().Text("bottom {{.Deep}} & <end>"))
	root.Children(deep)
	shared := Span().Class("shared").Text("shared <node> & {{.S}}")
	root.Children(shared, shared)
	return root
}

// invoiceRowTree builds the shape the invoice benchmark renders: one flex row
// per record, each holding eight styled cells. Row labels are formatted at
// build time, as an application would format them.
func invoiceRowTree(records int) *Element {
	rows := make([]*Element, 0, records)
	for i := 0; i < records; i++ {
		index := strconv.Itoa(i)
		row := Div().ID("invoice-row-"+index).
			Attr("data-action", "row:"+index).
			Class("inv-row").
			Display(DisplayFlex).
			Gap(Px(12)).
			Padding(Px(8))
		row.Children(
			invoiceCell("N", allocInvoiceNumWidth),
			invoiceCell("INV-0000", allocInvoiceIDWidth),
			invoiceCell("Client", allocInvoiceClientWidth),
			invoiceCell("2026-11-06", allocInvoiceDateWidth),
			invoiceCell("Item", allocInvoiceItemWidth),
			invoiceCell("1", allocInvoiceQtyWidth),
			invoiceCell("$0.00", allocInvoiceTotalWidth),
			Tag("span").Text("paid").Width(Px(allocInvoiceStatusWidth)).FontSize(Px(13)).
				Color(Hex("#15803d")),
		)
		rows = append(rows, row)
	}
	return Div().Display(DisplayFlex).FlexDirection(FlexDirectionColumn).
		Width(Percent(100)).Children(rows...)
}

func invoiceCell(text string, width float64) *Element {
	return Text(text).Width(Px(width)).FontSize(Px(13))
}

// TestAllocBuildThousandRowTree budgets the tree construction the invoice
// benchmark measures. Every row is nine elements with three attributes and
// several numeric styles. The remaining per-style allocation is the interface
// conversion in the generated fluent methods, not the element storage: a
// per-element style map or a formatted length string would add roughly one
// allocation for every element and every length on top of this.
func TestAllocBuildThousandRowTree(t *testing.T) {
	const records = allocInvoiceRows
	var sink *Element
	allocs := testing.AllocsPerRun(allocInvoiceRuns, func() {
		sink = invoiceRowTree(records)
	})
	if sink == nil || len(sink.children) != records {
		t.Fatal("tree construction lost rows")
	}
	t.Logf("build %d-row tree allocs/op: %v", records, allocs)
	if allocs > allocInvoiceTreeBudget {
		t.Fatalf("build %d-row tree allocs/op = %v, budget %v", records, allocs, allocInvoiceTreeBudget)
	}
}

// allocInvoiceTreeBudget covers one element allocation per node, the per-row
// label strings and child slices, and the generated methods' interface
// conversions, while failing if style or attribute storage regresses to maps.
const allocInvoiceTreeBudget = 34000

// TestEstimateSizeCoversThousandRowTree checks the render size hint is not
// capped below a page-sized document, so Render can grow its output buffer
// once instead of doubling it partway through.
func TestEstimateSizeCoversThousandRowTree(t *testing.T) {
	tree := invoiceRowTree(allocInvoiceRows)
	estimate := estimateSize(tree)
	fragment, err := Render(tree)
	if err != nil {
		t.Fatal(err)
	}
	if estimate < len(fragment) {
		t.Fatalf("estimateSize = %d, fragment = %d: the hint would grow the buffer mid-render", estimate, len(fragment))
	}
}

// naiveRenderer mirrors the pre-optimization builder output byte for byte: a
// per-node attribute map copy, html.EscapeString plus brace replacement, and
// fmt-based action and control ID synthesis. It is the frozen oracle for the
// golden parity test below and must not be "optimized".
type naiveRenderer struct {
	output     strings.Builder
	actions    map[string]func(context.Context) error
	active     map[*Element]bool
	nextID     int
	scope      string
	identities map[string]int
}

func naiveEscape(value string) string {
	return strings.ReplaceAll(html.EscapeString(value), "{", "&#123;")
}

func (r *naiveRenderer) naiveControlID(e *Element) string {
	h := fnv.New64a()
	_, _ = fmt.Fprintln(h, r.scope, e.tag, e.attrs.value("data-action"), e.attrs.value("name"), e.attrs.value("aria-label"))
	seen := map[*Element]bool{}
	var text func(*Element)
	text = func(node *Element) {
		if node == nil || seen[node] {
			return
		}
		seen[node] = true
		_, _ = fmt.Fprintln(h, node.tag, node.text)
		for _, child := range node.children {
			text(child)
		}
	}
	text(e)
	base := fmt.Sprintf("ownframe-ui-control-%x", h.Sum64())
	r.identities[base]++
	return fmt.Sprintf("%s-%d", base, r.identities[base])
}

func (r *naiveRenderer) naiveVisit(e *Element, depth int) error {
	if e == nil {
		return nil
	}
	if e.err != nil {
		return e.err
	}
	if depth >= maxDepth {
		return fmt.Errorf("ui: element nesting exceeds %d", maxDepth)
	}
	if !supportedTag(e.tag) {
		return fmt.Errorf("ui: unsupported HTML tag %q", e.tag)
	}
	if r.active[e] {
		return fmt.Errorf("ui: cycle in component tree at <%s>", e.tag)
	}
	r.active[e] = true
	defer delete(r.active, e)

	void := e.tag == "input" || e.tag == "img" || e.tag == "br" || e.tag == "hr"
	if void && (e.text != "" || len(e.children) != 0) {
		return fmt.Errorf("ui: void HTML element <%s> cannot have children or text", e.tag)
	}

	attrs := make(map[string]string, e.attrs.len()+3)
	for i := 0; i < e.attrs.len(); i++ {
		entry := e.attrs.at(i)
		attrs[entry.name] = entry.value
	}
	if e.onClick != nil {
		if _, exists := attrs["data-action"]; exists {
			return fmt.Errorf("ui: cannot combine OnClick with a custom data-action")
		}
		r.nextID++
		action := fmt.Sprintf("ownframe-ui-%d", r.nextID)
		attrs["data-action"] = action
		if attrs["id"] == "" {
			attrs["id"] = r.naiveControlID(e)
		}
		r.actions[action] = e.onClick
	}
	if attrs["data-action"] != "" && attrs["id"] == "" {
		r.nextID++
		attrs["id"] = r.naiveControlID(e)
	}
	previousScope := r.scope
	if attrs["id"] != "" {
		r.scope = attrs["id"]
	}
	defer func() { r.scope = previousScope }()
	if e.styles.len() != 0 {
		keys := make([]string, 0, e.styles.len())
		for i := 0; i < e.styles.len(); i++ {
			keys = append(keys, e.styles.at(i).name)
		}
		sort.Strings(keys)
		var style strings.Builder
		for _, name := range keys {
			entry, _ := e.styles.get(name)
			style.WriteString(name)
			style.WriteByte(':')
			style.WriteString(entry.cssTextValue())
			style.WriteByte(';')
		}
		attrs["style"] = style.String()
	}

	r.output.WriteByte('<')
	r.output.WriteString(e.tag)
	names := make([]string, 0, len(attrs))
	for name := range attrs {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		r.output.WriteByte(' ')
		r.output.WriteString(name)
		r.output.WriteString("=\"")
		r.output.WriteString(naiveEscape(attrs[name]))
		r.output.WriteByte('"')
	}
	r.output.WriteByte('>')
	if void {
		return nil
	}
	r.output.WriteString(naiveEscape(e.text))
	for _, child := range e.children {
		if err := r.naiveVisit(child, depth+1); err != nil {
			return err
		}
	}
	r.output.WriteString("</")
	r.output.WriteString(e.tag)
	r.output.WriteByte('>')
	return nil
}

// naiveRender serializes with the frozen pre-optimization algorithm.
func naiveRender(root *Element) (string, error) {
	r := &naiveRenderer{
		actions:    make(map[string]func(context.Context) error),
		active:     make(map[*Element]bool),
		identities: make(map[string]int),
	}
	if err := r.naiveVisit(root, 0); err != nil {
		return "", err
	}
	return r.output.String(), nil
}

// goldenAllocDocFingerprint pins the pre-optimization render bytes of
// allocBenchTree: FNV-1a 64 hex of the markup plus its length. Any behavioral
// change in Render fails this test alongside the exact oracle compare.
const (
	goldenAllocDocHash   = "f2fb5fc61eec12c1"
	goldenAllocDocLength = 48064
)

func fingerprintDoc(t *testing.T, doc string) {
	t.Helper()
	h := fnv.New64a()
	_, _ = h.Write([]byte(doc))
	t.Logf("golden fingerprint: hash=%x length=%d", h.Sum64(), len(doc))
	if got := fmt.Sprintf("%x", h.Sum64()); got != goldenAllocDocHash {
		t.Fatalf("golden hash changed: got %s want %s", got, goldenAllocDocHash)
	}
	if len(doc) != goldenAllocDocLength {
		t.Fatalf("golden length changed: got %d want %d", len(doc), goldenAllocDocLength)
	}
}

func TestBuilderGoldenParityBigDoc(t *testing.T) {
	t.Parallel()
	want, err := naiveRender(allocBenchTree())
	if err != nil {
		t.Fatal(err)
	}
	got, err := Render(allocBenchTree())
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		for i := 0; i < len(got) && i < len(want); i++ {
			if got[i] != want[i] {
				lo := i - 120
				if lo < 0 {
					lo = 0
				}
				hi := i + 120
				if hi > len(got) {
					hi = len(got)
				}
				wHi := i + 120
				if wHi > len(want) {
					wHi = len(want)
				}
				t.Fatalf("markup diverged at byte %d:\n got: ...%q...\nwant: ...%q...", i, got[lo:hi], want[lo:wHi])
			}
		}
		t.Fatalf("markup length diverged: got %d want %d", len(got), len(want))
	}
	fingerprintDoc(t, got)
}

func TestAllocRenderTree(t *testing.T) {
	tree := allocBenchTree()
	r, err := renderTree(tree)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := len(r.actions), allocSections*2; got != want {
		t.Fatalf("callback registry holds %d actions, want %d", got, want)
	}
	var sink int
	allocs := testing.AllocsPerRun(allocRenderRuns, func() {
		res, err := renderTree(tree)
		if err != nil {
			t.Error(err)
			return
		}
		sink += len(res.output.String())
	})
	if sink == 0 {
		t.Fatal("empty render output")
	}
	t.Logf("renderTree allocs/op: %v", allocs)
	if allocs > allocRenderTreeBudget {
		t.Fatalf("renderTree allocs/op = %v, budget %v", allocs, allocRenderTreeBudget)
	}
}

const allocRenderTreeBudget = 500

func TestAllocDocumentAssembly(t *testing.T) {
	title := "A & {{.Admin}} <tag> \"quoted\" 'single' {brace}"
	fragment, err := Render(allocBenchTree())
	if err != nil {
		t.Fatal(err)
	}
	want := "<!doctype html><html><head><meta charset=\"utf-8\"><title>" +
		escape(title) + "</title><style>html,body{margin:0;padding:0}</style></head><body>" + fragment + "</body></html>"
	got := Document(title, fragment)
	if got != want {
		t.Fatalf("Document assembly diverged:\n got: %q\nwant: %q", got, want)
	}
	var sink int
	allocs := testing.AllocsPerRun(allocRenderRuns, func() {
		sink += len(Document(title, fragment))
	})
	if sink == 0 {
		t.Fatal("empty document output")
	}
	t.Logf("Document allocs/op: %v", allocs)
	if allocs > allocDocumentBudget {
		t.Fatalf("Document allocs/op = %v, budget %v", allocs, allocDocumentBudget)
	}
}

const allocDocumentBudget = 2

// generatedIDForButton extracts the id attribute of the button whose visible
// label matches, or "" when absent.
func generatedIDForButton(t *testing.T, markup, label string) string {
	t.Helper()
	closeTag := ">" + label + "</button>"
	end := strings.Index(markup, closeTag)
	if end < 0 {
		return ""
	}
	start := end - allocIDWindow
	if start < 0 {
		start = 0
	}
	open := strings.LastIndex(markup[start:end], "<button")
	if open < 0 {
		return ""
	}
	tag := markup[start+open : end]
	idPrefix := "id=\""
	at := strings.Index(tag, idPrefix)
	if at < 0 {
		return ""
	}
	rest := tag[at+len(idPrefix):]
	quote := strings.IndexByte(rest, '"')
	if quote < 0 {
		return ""
	}
	return rest[:quote]
}

func TestControlIDStability(t *testing.T) {
	t.Parallel()
	first, err := Render(allocBenchTree())
	if err != nil {
		t.Fatal(err)
	}
	second, err := Render(allocBenchTree())
	if err != nil {
		t.Fatal(err)
	}
	if first != second {
		t.Fatal("identical input rendered different markup")
	}
	export := func(label string) *Element {
		return Button(label).OnClick(func(context.Context) error { return nil })
	}
	without, err := Render(Div().ID("toolbar").Children(export("Export")))
	if err != nil {
		t.Fatal(err)
	}
	with, err := Render(Div().ID("toolbar").Children(
		Button("Transient").OnClick(func(context.Context) error { return nil }),
		export("Export"),
	))
	if err != nil {
		t.Fatal(err)
	}
	before := generatedIDForButton(t, without, "Export")
	after := generatedIDForButton(t, with, "Export")
	if before == "" || after == "" {
		t.Fatalf("missing generated Export id: before=%q after=%q", before, after)
	}
	if before != after {
		t.Fatalf("conditional sibling changed Export id %s -> %s", before, after)
	}
}
