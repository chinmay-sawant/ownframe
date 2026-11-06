package ui

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/pagetest"
)

// mountView builds a page and draws it once, so the retained tree, cache, and
// display a patch needs are all in place.
func mountView(t *testing.T, cfg ownframe.Config, root *Element) *View {
	t.Helper()
	if cfg.Width == 0 {
		cfg.Width = testWidth
	}
	if cfg.Height == 0 {
		cfg.Height = testHeight
	}
	view, err := Mount(cfg, root)
	if err != nil {
		t.Fatal(err)
	}
	if err := view.Page.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}
	return view
}

// fillColor returns the first fill operation owned by the box with id.
func fillColor(display *ownframe.Display, id string) (r, g, b float64, ok bool) {
	if display == nil {
		return 0, 0, 0, false
	}
	index := -1
	for i := range display.Boxes {
		if display.Boxes[i].ID == id {
			index = i
			break
		}
	}
	if index < 0 {
		return 0, 0, 0, false
	}
	for i := range display.Ops {
		if i >= len(display.OpOwners) || display.OpOwners[i] != index {
			continue
		}
		op := &display.Ops[i]
		if op.Kind != ownframe.DisplayOpFillRect || op.W <= 0 || op.H <= 0 {
			continue
		}
		return op.R, op.G, op.B, true
	}
	return 0, 0, 0, false
}

const (
	hexColorLength = 7
	colorChannel   = 255
	colorTolerance = 0.01
)

func sameColor(r, g, b float64, hex string) bool {
	if len(hex) != hexColorLength || hex[0] != '#' {
		return false
	}
	var want [3]float64
	for i := range want {
		v, err := strconv.ParseUint(hex[1+2*i:3+2*i], 16, 8)
		if err != nil {
			return false
		}
		want[i] = float64(v) / colorChannel
	}
	near := func(got, want float64) bool {
		return got-want < colorTolerance && want-got < colorTolerance
	}
	return near(r, want[0]) && near(g, want[1]) && near(b, want[2])
}

func TestDiffTreeCollectsKeyedPatches(t *testing.T) {
	t.Parallel()
	last := Div().ID("root").Children(
		Div().ID("title").Text("old"),
		Div().ID("box").Width(Px(10)).Height(Px(20)),
	)
	root := Div().ID("root").Children(
		Div().ID("title").Text("new"),
		Div().ID("box").Width(Px(42)).Height(Px(20)),
	)

	d := diffTree(last, root)
	if !d.ok {
		t.Fatal("matching structure fell back")
	}
	if len(d.patches) != 2 {
		t.Fatalf("patches = %d, want 2", len(d.patches))
	}
	if d.patches[0].ID != "title" || len(d.patches[0].Edits) != 1 ||
		d.patches[0].Edits[0].Text == nil || *d.patches[0].Edits[0].Text != "new" {
		t.Fatalf("title patch = %+v", d.patches[0])
	}
	if d.patches[1].ID != "box" || len(d.patches[1].Edits) != 1 ||
		d.patches[1].Edits[0].Attribute != "style" ||
		d.patches[1].Edits[0].Value != "height:20px;width:42px;" {
		t.Fatalf("box patch = %+v", d.patches[1])
	}
}

func TestDiffTreePatchesOverflowStyles(t *testing.T) {
	t.Parallel()
	build := func(fontSize float64) *Element {
		return Div().ID("root").Width(Px(1)).Height(Px(2)).Padding(Px(3)).
			Margin(Px(4)).Gap(Px(5)).FontSize(Px(fontSize)).Color(Hex("#fff"))
	}
	d := diffTree(build(10), build(12))
	if !d.ok || len(d.patches) != 1 {
		t.Fatalf("ok = %v patches = %d, want one patch", d.ok, len(d.patches))
	}
	edits := d.patches[0].Edits
	if len(edits) != 1 || edits[0].Attribute != "style" {
		t.Fatalf("edits = %+v", edits)
	}
	const want = "color:#fff;font-size:12px;gap:5px;height:2px;margin:4px;padding:3px;width:1px;"
	if edits[0].Value != want {
		t.Fatalf("style = %q, want %q", edits[0].Value, want)
	}
}

func TestDiffTreeRejectsUnpatchableTrees(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		last *Element
		root *Element
	}{
		{
			name: "child count",
			last: Div().ID("root").Children(Div().ID("a").Text("x")),
			root: Div().ID("root").Children(Div().ID("a").Text("x"), Div().ID("b").Text("y")),
		},
		{
			name: "changed node without id",
			last: Div().Children(Span().Text("old")),
			root: Div().Children(Span().Text("new")),
		},
		{
			name: "id rename",
			last: Div().Children(Div().ID("old").Text("same")),
			root: Div().Children(Div().ID("new").Text("same")),
		},
		{
			name: "removed attribute",
			last: Div().ID("root").Class("row"),
			root: Div().ID("root"),
		},
		{
			name: "removed style",
			last: Div().ID("root").Width(Px(10)),
			root: Div().ID("root"),
		},
		{
			name: "onClick appears",
			last: Div().ID("root"),
			root: Div().ID("root").OnClick(func(context.Context) error { return nil }),
		},
		{
			name: "text node with children",
			last: Div().ID("root").Text("old").Children(Div().ID("a").Text("a")),
			root: Div().ID("root").Text("new").Children(Div().ID("a").Text("a")),
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if d := diffTree(tc.last, tc.root); d.ok {
				t.Fatalf("ok = true, want fallback: %+v", d.patches)
			}
		})
	}
}

func TestDiffTreeCapsPatchCount(t *testing.T) {
	t.Parallel()
	tree := func(rows int, class string) *Element {
		root := Div().ID("root")
		for i := 0; i < rows; i++ {
			root.Children(Div().ID("row-" + fmt.Sprint(i)).Class(class))
		}
		return root
	}

	d := diffTree(tree(maxPatches, "a"), tree(maxPatches, "b"))
	if !d.ok || len(d.patches) != maxPatches {
		t.Fatalf("at cap: ok = %v patches = %d, want true and %d", d.ok, len(d.patches), maxPatches)
	}
	d = diffTree(tree(maxPatches+1, "a"), tree(maxPatches+1, "b"))
	if d.ok {
		t.Fatalf("above cap: ok = true with %d patches", len(d.patches))
	}
}

func TestDiffTreeCollectsFreshActions(t *testing.T) {
	t.Parallel()
	oldCalls, newCalls := 0, 0
	last := Div().ID("root").Children(
		Button("One").ID("one").OnClick(func(context.Context) error { oldCalls++; return nil }),
		Button("Two").ID("two").OnClick(func(context.Context) error { return nil }),
	)
	root := Div().ID("root").Children(
		Button("One").ID("one").OnClick(func(context.Context) error { newCalls++; return nil }),
		Button("Two").ID("two").OnClick(func(context.Context) error { return nil }),
	)

	d := diffTree(last, root)
	if !d.ok {
		t.Fatal("callback-only change fell back")
	}
	if len(d.patches) != 0 {
		t.Fatalf("patches = %d, want 0", len(d.patches))
	}
	if len(d.actions) != 2 {
		t.Fatalf("actions = %d, want 2", len(d.actions))
	}
	handler := d.actions[actionName(1)]
	if handler == nil {
		t.Fatal("first action missing")
	}
	if err := handler(context.Background()); err != nil {
		t.Fatal(err)
	}
	if newCalls != 1 || oldCalls != 0 {
		t.Fatalf("new = %d old = %d, want 1 and 0", newCalls, oldCalls)
	}
}

func TestInlineStyleOrdersAndDecodesValues(t *testing.T) {
	t.Parallel()
	e := Div().Width(Px(10)).CustomProperty("--label", `"a&b"`)
	if e.err != nil {
		t.Fatal(e.err)
	}
	if got, want := inlineStyle(e), `--label:"a&b";width:10px;`; got != want {
		t.Fatalf("inlineStyle = %q, want %q", got, want)
	}
}

func TestViewUpdatePatchesTextAndStyle(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	tree := func(label, color string) *Element {
		return Div().ID("root").Children(
			Div().ID("label").Width(Px(120)).Height(Px(20)).Text(label),
			Div().ID("box").Width(Px(100)).Height(Px(20)).BackgroundColor(Hex(color)),
		)
	}
	view := mountView(t, ownframe.Config{}, tree("old", "#112233"))
	before := view.Page.Probe().Stats()

	if err := view.Update(ctx, tree("new", "#445566")); err != nil {
		t.Fatal(err)
	}
	after := view.Page.Probe().Stats()
	if after.Parses != before.Parses || after.Cascades != before.Cascades {
		t.Fatalf("patch parsed %d->%d cascaded %d->%d",
			before.Parses, after.Parses, before.Cascades, after.Cascades)
	}
	label, ok := pagetest.BoxByID(view.Page, "label")
	if !ok || label.Text != "new" {
		t.Fatalf("label = %q (found %v), want %q", label.Text, ok, "new")
	}
	r, g, b, ok := fillColor(view.Page.Probe().Display(), "box")
	if !ok || !sameColor(r, g, b, "#445566") {
		t.Fatalf("box fill = %v,%v,%v (found %v), want #445566", r, g, b, ok)
	}
	if !strings.Contains(view.Page.HTML(), "new") ||
		!strings.Contains(view.Page.HTML(), "background-color:#445566") {
		t.Fatal("adopted source missing the patched text or style")
	}
}

func TestViewUpdatePatchesSelectionCase(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	const selectionCSS = `.sel { background-color: #445566; }`
	tree := func(selected int, summary string) *Element {
		rows := make([]*Element, 0, 3)
		for i := 1; i <= 3; i++ {
			class := "row"
			if i == selected {
				class = "row sel"
			}
			rows = append(rows, Div().ID(fmt.Sprintf("row-%d", i)).Class(class).Height(Px(10)))
		}
		return Div().ID("table").Children(rows...).Children(
			Div().ID("summary").Width(Px(120)).Height(Px(20)).Text(summary),
		)
	}
	view := mountView(t, ownframe.Config{Theme: selectionCSS}, tree(1, "Row 1 selected"))
	before := view.Page.Probe().Stats()

	// Two class changes and one text change: three patches, below the cap.
	if err := view.Update(ctx, tree(2, "Row 2 selected")); err != nil {
		t.Fatal(err)
	}
	after := view.Page.Probe().Stats()
	if after.Parses != before.Parses || after.Cascades != before.Cascades {
		t.Fatalf("selection update parsed %d->%d cascaded %d->%d",
			before.Parses, after.Parses, before.Cascades, after.Cascades)
	}
	display := view.Page.Probe().Display()
	r, g, b, ok := fillColor(display, "row-2")
	if !ok || !sameColor(r, g, b, "#445566") {
		t.Fatalf("selected row fill = %v,%v,%v (found %v), want #445566", r, g, b, ok)
	}
	summary, ok := pagetest.BoxByID(view.Page, "summary")
	if !ok || summary.Text != "Row 2 selected" {
		t.Fatalf("summary = %q (found %v), want %q", summary.Text, ok, "Row 2 selected")
	}
}

// TestViewUpdateAfterNilActionMap pins the pooled-renderer case that panicked
// on a nil actions map: a no-change update with no OnClick handlers clears the
// view registry, a later structural update pools that nil map, and the next
// tree with a handler must still render and dispatch.
func TestViewUpdateAfterNilActionMap(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	tree := func(children ...*Element) *Element {
		return Div().ID("root").Children(children...)
	}
	view := mountView(t, ownframe.Config{}, tree(Span().ID("label").Text("one")))
	if err := view.Update(ctx, tree(Span().ID("label").Text("one"))); err != nil {
		t.Fatal(err)
	}
	if view.actions != nil {
		t.Fatal("no-handler update kept a non-nil action registry")
	}
	if err := view.Update(ctx, tree(
		Span().ID("label").Text("one"),
		Span().ID("extra").Text("two"),
	)); err != nil {
		t.Fatal(err)
	}

	clicks := 0
	button := Button("Run").ID("run").OnClick(func(context.Context) error {
		clicks++

		return nil
	})
	if err := view.Update(ctx, tree(button)); err != nil {
		t.Fatal(err)
	}
	if err := pagetest.Click(ctx, view.Page, "run"); err != nil {
		t.Fatal(err)
	}
	if clicks != 1 {
		t.Fatalf("clicks = %d, want 1", clicks)
	}
}

func TestViewUpdateKeepsNewHandlerAfterPatch(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	oldCalls, newCalls, patchedCalls := 0, 0, 0
	tree := func(label string, handler func(context.Context) error) *Element {
		return Div().ID("root").Children(
			Button(label).ID("go").Width(Px(120)).Height(Px(20)).OnClick(handler),
		)
	}
	view := mountView(t, ownframe.Config{}, tree("Go", func(context.Context) error {
		oldCalls++
		return nil
	}))
	if err := pagetest.Click(ctx, view.Page, "go"); err != nil {
		t.Fatal(err)
	}

	// A callback-only change touches no element: the page stays put and the
	// action registry moves to the rebuilt closure.
	if err := view.Update(ctx, tree("Go", func(context.Context) error {
		newCalls++
		return nil
	})); err != nil {
		t.Fatal(err)
	}
	if err := pagetest.Click(ctx, view.Page, "go"); err != nil {
		t.Fatal(err)
	}

	// A text change takes the patch path; its data-action key must still
	// resolve to the handler from the rebuilt tree.
	before := view.Page.Probe().Stats()
	if err := view.Update(ctx, tree("Go now", func(context.Context) error {
		patchedCalls++
		return nil
	})); err != nil {
		t.Fatal(err)
	}
	after := view.Page.Probe().Stats()
	if after.Parses != before.Parses || after.Cascades != before.Cascades {
		t.Fatalf("handler update parsed %d->%d cascaded %d->%d",
			before.Parses, after.Parses, before.Cascades, after.Cascades)
	}
	if err := pagetest.Click(ctx, view.Page, "go"); err != nil {
		t.Fatal(err)
	}
	if oldCalls != 1 || newCalls != 1 || patchedCalls != 1 {
		t.Fatalf("old = %d new = %d patched = %d, want 1 each", oldCalls, newCalls, patchedCalls)
	}
}

func TestViewUpdateNoChangeLeavesPageAlone(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	calls := 0
	tree := func(handler func(context.Context) error) *Element {
		return Div().ID("root").Children(
			Button("Same").ID("go").Width(Px(120)).Height(Px(20)).OnClick(handler),
		)
	}
	view := mountView(t, ownframe.Config{}, tree(func(context.Context) error { return nil }))
	before := view.Page.Probe().Stats()
	generation := view.Page.Probe().Generation()

	if err := view.Update(ctx, tree(func(context.Context) error { calls++; return nil })); err != nil {
		t.Fatal(err)
	}
	after := view.Page.Probe().Stats()
	if after.Parses != before.Parses || after.Cascades != before.Cascades {
		t.Fatalf("no-op update parsed %d->%d cascaded %d->%d",
			before.Parses, after.Parses, before.Cascades, after.Cascades)
	}
	if got := view.Page.Probe().Generation(); got != generation {
		t.Fatalf("generation = %d, want %d", got, generation)
	}
	if err := pagetest.Click(ctx, view.Page, "go"); err != nil {
		t.Fatal(err)
	}
	if calls != 1 {
		t.Fatalf("callback calls = %d, want 1", calls)
	}
}

func TestViewUpdateFallsBackForStructureChange(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	view := mountView(t, ownframe.Config{}, Div().ID("root").Children(
		Div().ID("label").Text("one"),
	))
	before := view.Page.Probe().Stats()

	if err := view.Update(ctx, Div().ID("root").Children(
		Div().ID("label").Text("two"),
		Div().ID("extra").Text("added"),
	)); err != nil {
		t.Fatal(err)
	}
	after := view.Page.Probe().Stats()
	if after.Parses != before.Parses+1 || after.Cascades != before.Cascades+1 {
		t.Fatalf("fallback parsed %d->%d cascaded %d->%d, want +1 and +1",
			before.Parses, after.Parses, before.Cascades, after.Cascades)
	}
	if !strings.Contains(view.Page.HTML(), "added") {
		t.Fatal("fallback source missing the new node")
	}
}

func TestViewUpdateFallsBackAbovePatchCap(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	const classCSS = `.b { background-color: #445566; }`
	tree := func(class string) *Element {
		root := Div().ID("root")
		for i := 0; i < maxPatches+1; i++ {
			root.Children(Div().ID(fmt.Sprintf("row-%d", i)).Class(class).Height(Px(4)))
		}
		return root
	}
	view := mountView(t, ownframe.Config{Theme: classCSS}, tree("a"))
	before := view.Page.Probe().Stats()

	if err := view.Update(ctx, tree("b")); err != nil {
		t.Fatal(err)
	}
	after := view.Page.Probe().Stats()
	if after.Parses != before.Parses+1 || after.Cascades != before.Cascades+1 {
		t.Fatalf("over-cap update parsed %d->%d cascaded %d->%d, want +1 and +1",
			before.Parses, after.Parses, before.Cascades, after.Cascades)
	}
	r, g, b, ok := fillColor(view.Page.Probe().Display(), "row-0")
	if !ok || !sameColor(r, g, b, "#445566") {
		t.Fatalf("row fill = %v,%v,%v (found %v), want #445566", r, g, b, ok)
	}
}

func TestViewUpdateFallsBackWhenPatchRefused(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	tree := func(width float64) *Element {
		return Div().ID("root").Children(
			Div().ID("box").Width(Px(width)).Height(Px(20)),
		)
	}
	view := mountView(t, ownframe.Config{}, tree(10))
	before := view.Page.Probe().Stats()

	// The engine refuses a subtree patch that changes the element's size, so
	// Update must reparse the snapshot instead of leaving the page stale.
	if err := view.Update(ctx, tree(42)); err != nil {
		t.Fatal(err)
	}
	after := view.Page.Probe().Stats()
	if after.Parses != before.Parses+1 || after.Cascades != before.Cascades+1 {
		t.Fatalf("refused patch parsed %d->%d cascaded %d->%d, want +1 and +1",
			before.Parses, after.Parses, before.Cascades, after.Cascades)
	}
	box, ok := pagetest.BoxByID(view.Page, "box")
	if !ok || box.W != 42 {
		t.Fatalf("box width = %v (found %v), want 42", box.W, ok)
	}
}

func TestViewHandleDispatchAfterPatch(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	var got string
	tree := func(label string) *Element {
		return Div().ID("root").Children(
			Button(label).ID("row").Class("row").Attr("data-action", "select").
				Width(Px(120)).Height(Px(20)),
		)
	}
	view := mountView(t, ownframe.Config{}, tree("Row"))
	view.Handle(ownframe.Handlers{Click: func(_ context.Context, box ownframe.Box) error {
		got = box.Action
		return nil
	}})

	if err := view.Update(ctx, tree("Row two")); err != nil {
		t.Fatal(err)
	}
	if err := pagetest.Click(ctx, view.Page, "row"); err != nil {
		t.Fatal(err)
	}
	if got != "select" {
		t.Fatalf("action = %q, want select", got)
	}
}
