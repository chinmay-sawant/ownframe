package page

import (
	"context"
	"github.com/chinmay-sawant/ownframe/internal/host"
	"testing"
)

// TestClickRepaintsOneBox clicks a counter whose handler declares the changed
// id, then checks that every operation the click changed lies inside the
// dirty rect. Operations outside it must be byte-identical to the previous
// display list.
func TestClickRepaintsOneBox(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	n := 0

	p, err := New(Config{
		HTML:   `<body style="margin:0"><div id="count"><span id="num">{{.N}}</span></div><div id="inc" data-action="add">+</div></body>`,
		Width:  320,
		Height: 200,
	})
	if err != nil {
		t.Fatal(err)
	}

	p.SetData(struct{ N int }{0})
	p.Handle(Handlers{Click: func(_ context.Context, box Box) error {
		if box.Action == "add" {
			n++
			p.SetData(struct{ N int }{n})
			p.Invalidate("count")
		}

		return nil
	}})

	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	p.TakeDirty()
	before := p.draw.Display

	x, y := clickCenterOf(t, p, "inc")
	if err := p.Click(ctx, x, y); err != nil {
		t.Fatal(err)
	}

	rect, ok := p.TakeDirty()
	if !ok {
		t.Fatal("no dirty rect after the click")
	}

	after := p.draw.Display
	if before == nil || after == nil || len(before.Ops) != len(after.Ops) {
		t.Fatalf("op counts = %d then %d", opCount(before), opCount(after))
	}

	for i := range after.Ops {
		if sameOp(&before.Ops[i], &after.Ops[i]) {
			continue
		}

		if opBounds(&after.Ops[i], after).Intersect(rect).Empty() {
			t.Fatalf("op %d changed outside the dirty rect %v", i, rect)
		}
	}
}

func opCount(d *Display) int {
	if d == nil {
		return -1
	}

	return len(d.Ops)
}

// clickCenterOf returns the center of one box for a click.
func clickCenterOf(t *testing.T, p *Page, id string) (float64, float64) {
	t.Helper()

	for _, box := range p.boxes {
		if box.ID == id {
			return box.X + box.W/2, box.Y + box.H/2
		}
	}

	t.Fatalf("box %q not found", id)

	return 0, 0
}

// TestClickPaintKeepsTheStateLayout checks that a paint-only click in a
// batch still lays out a state the sheets match: the handler painted the
// pixels, but the focus state needs a layout.
func TestClickPaintKeepsTheStateLayout(t *testing.T) {
	ctx := context.Background()
	p := patchPage(t, `<style>#go{width:100px;height:40px}#go:focus{width:120px}</style>`+
		`<button id="go">Go</button>`)
	p.Handle(Handlers{
		ClickPaint: func(context.Context, Box) (bool, error) { return true, nil },
	})

	before := p.boxByID("go").W
	box := p.boxByID("go")
	x, y := box.X+box.W/2, box.Y+box.H/2

	p.BeginRedrawBatch()
	if err := p.Focus(ctx, "go"); err != nil {
		t.Fatal(err)
	}
	if err := p.Click(ctx, x, y); err != nil {
		t.Fatal(err)
	}
	if err := p.EndRedrawBatch(ctx); err != nil {
		t.Fatal(err)
	}

	if got := p.boxByID("go").W; got == before {
		t.Fatalf("width %v, want the focused width", got)
	}
}

// TestClickPaintSkipsTheLayoutOfAnUnusedState checks that a paint-only click
// in a batch on a page whose sheets match no state keeps the layout out.
func TestClickPaintSkipsTheLayoutOfAnUnusedState(t *testing.T) {
	ctx := context.Background()
	p := patchPage(t, `<style>#go{width:100px;height:40px}</style><button id="go">Go</button>`)
	p.Handle(Handlers{
		ClickPaint: func(context.Context, Box) (bool, error) { return true, nil },
	})

	layouts := p.Stats().Layouts
	box := p.boxByID("go")
	x, y := box.X+box.W/2, box.Y+box.H/2

	p.BeginRedrawBatch()
	if err := p.Focus(ctx, "go"); err != nil {
		t.Fatal(err)
	}
	if err := p.Click(ctx, x, y); err != nil {
		t.Fatal(err)
	}
	if err := p.EndRedrawBatch(ctx); err != nil {
		t.Fatal(err)
	}

	if got := p.Stats().Layouts; got != layouts {
		t.Fatalf("layouts %d, want %d", got, layouts)
	}
}

// TestClickSingleSpotResolvesAnonymousChild pins the one-spot click path: the
// spot scanned for the animation-trigger pass is the spot the sequence
// resolves, so a click on an anonymous child still reaches the nearest
// id-bearing ancestor and its data-action.
func TestClickSingleSpotResolvesAnonymousChild(t *testing.T) {
	ctx := context.Background()
	p := overlapPage(t)
	p.setBoxes([]Box{
		{ID: "card", Action: "open", X: 0, Y: 0, W: 100, H: 100},
		{ID: "", X: 10, Y: 10, W: 20, H: 20},
	})

	spot := p.boxAt(15, 15)
	if !spot.ok || spot.box.ID != "" || spot.id != "card" {
		t.Fatalf("spot = %+v, want the anonymous child under card", spot)
	}

	var clicked Box
	p.Handle(Handlers{Click: func(_ context.Context, b Box) error {
		clicked = b

		return nil
	}})

	if err := p.Click(ctx, 15, 15); err != nil {
		t.Fatal(err)
	}
	if clicked.ID != "card" || clicked.Action != "open" {
		t.Fatalf("clicked = %+v, want the card control through its child", clicked)
	}
}

func TestCursorShapes(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	p := bftPage(t, `<input id="t" type="text"`+bedWide+`>`+
		`<button id="go">Go</button>`+
		`<a id="l" href="http://example.com/">link</a>`+
		`<div id="ord">plain</div>`)
	bftDraw(t, p, nil)

	steps := []struct {
		id   string
		want host.Shape
	}{
		{"t", host.ShapeText},
		{"go", host.ShapePointer},
		{"ord", host.ShapeDefault},
	}

	for _, step := range steps {
		b := p.boxByID(step.id)
		if err := p.Hover(ctx, b.X+b.W/2, b.Y+b.H/2); err != nil {
			t.Fatal(err)
		}

		if got := p.CursorShape(); got != step.want {
			t.Fatalf("%s shape %v want %v", step.id, got, step.want)
		}
	}

	link := runOp(p, "link")
	left, width := runSpan(p, link)
	if err := p.Hover(ctx, left+width/2, runBase(p, link)-4); err != nil {
		t.Fatal(err)
	}

	if got := p.CursorShape(); got != host.ShapePointer {
		t.Fatalf("link shape %v", got)
	}

	if err := p.Hover(ctx, 1, 1); err != nil {
		t.Fatal(err)
	}

	if got := p.CursorShape(); got != host.ShapeDefault {
		t.Fatalf("background shape %v", got)
	}
}

// TestFallbackRebuildsFixedHits checks that the bitmap path builds the hit
// map as well, so a fixed box still wins over the content it covers.
func TestFallbackRebuildsFixedHits(t *testing.T) {
	ctx := context.Background()
	p, err := New(Config{
		HTML: `<style>div { border-radius: 50% / 20%; }</style>` +
			`<div id="head" style="position:fixed;width:100px;height:50px;background:#c00"></div>` +
			`<div id="body" style="width:100px;height:50px;background:#00c"></div>`,
		Width:  200,
		Height: 200,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	if p.draw.Display != nil {
		t.Fatal("the elliptical radius replayed, want the bitmap path")
	}
	if len(p.hits) == 0 {
		t.Fatal("the bitmap path left no hit map")
	}

	spot := p.boxAt(10, 10)
	if !spot.ok || spot.id != "head" {
		t.Fatalf("boxAt = %+v, want the fixed header", spot)
	}
}

// fixedHeaderPage lays out a fixed header over a content block. The engine
// moves the fixed header behind the content box in the box list, so the hit
// test has to follow the paint order rather than the list order.
func fixedHeaderPage(t *testing.T) *Page {
	t.Helper()
	p, err := New(Config{
		HTML: `<div id="wrap"><div id="head" style="position:fixed;` +
			`width:100px;height:50px;background:#c00"></div></div>` +
			`<div id="body" style="width:100px;height:50px;background:#00c"></div>`,
		Width:  200,
		Height: 200,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}
	return p
}

// TestFixedBoxWinsOverContent checks that a fixed box wins the hit test
// against the content it covers, that a click lands on it, and that the
// content outside the fixed box still wins its own point.
func TestFixedBoxWinsOverContent(t *testing.T) {
	ctx := context.Background()
	p := fixedHeaderPage(t)

	spot := p.boxAt(10, 10)
	if !spot.ok || spot.id != "head" {
		t.Fatalf("boxAt = %+v, want the fixed header", spot)
	}

	var clicked Box
	p.Handle(Handlers{Click: func(_ context.Context, b Box) error {
		clicked = b
		return nil
	}})
	if err := p.Click(ctx, 10, 10); err != nil {
		t.Fatal(err)
	}
	if clicked.ID != "head" {
		t.Fatalf("clicked = %+v, want the fixed header", clicked)
	}

	spot = p.boxAt(10, 55)
	if !spot.ok || spot.id != "body" {
		t.Fatalf("boxAt = %+v, want the content block", spot)
	}
}

// TestFixedBoxKeepsInnerControl checks that a control inside a fixed box
// still wins, because the engine paints the whole subtree of a fixed
// element with it.
func TestFixedBoxKeepsInnerControl(t *testing.T) {
	ctx := context.Background()
	p, err := New(Config{
		HTML: `<div id="head" style="position:fixed;width:100px;height:50px;` +
			`background:#c00"><button id="go" style="width:40px;height:20px">go</button></div>` +
			`<div id="body" style="width:100px;height:50px;background:#00c"></div>`,
		Width:  200,
		Height: 200,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	spot := p.boxAt(10, 10)
	if !spot.ok || spot.id != "go" {
		t.Fatalf("boxAt = %+v, want the control inside the fixed box", spot)
	}
}

func TestLongPressSelectsWord(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	p := bftPage(t, `<input id="t" type="text" value="hello world"`+bedWide+`>`)
	bftDraw(t, p, nil)

	op := runOp(p, "hello world")
	left, width := runSpan(p, op)
	base := runBase(p, op)

	claimed, err := p.LongPress(ctx, left+width*0.1, base-4)
	if err != nil {
		t.Fatal(err)
	}

	if !claimed {
		t.Fatal("a text-field hold is unclaimed")
	}

	if start, end := p.formState().Bounds(11); start != 0 || end != 5 {
		t.Fatalf("word [%d,%d)", start, end)
	}
}

func TestLongPressCallsHandler(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	p := bftPage(t, `<div id="card" style="width:120px;height:40px">`+
		`<div style="width:20px;height:20px">hit</div></div>`)
	bftDraw(t, p, nil)

	got := ""
	p.Handle(Handlers{LongPress: func(_ context.Context, box Box) error {
		got = box.ID

		return nil
	}})

	child := Box{}

	for _, b := range p.Boxes() {
		if b.Tag == "div" && b.ID == "" && b.W > 0 {
			child = b
		}
	}

	if child.W <= 0 {
		t.Fatal("no child div box")
	}

	claimed, err := p.LongPress(ctx, child.X+child.W/2, child.Y+child.H/2)
	if err != nil {
		t.Fatal(err)
	}

	if !claimed {
		t.Fatal("a handled hold is unclaimed")
	}

	if got != "card" {
		t.Fatalf("handler box = %q, want card", got)
	}
}

func TestLongPressNilHandlerUnclaimed(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	p := bftPage(t, `<div id="card" style="width:120px;height:40px">tap</div>`)
	bftDraw(t, p, nil)

	box := p.boxByID("card")

	claimed, err := p.LongPress(ctx, box.X+box.W/2, box.Y+box.H/2)
	if err != nil {
		t.Fatal(err)
	}

	if claimed {
		t.Fatal("a nil handler claimed the press")
	}
}

func TestSwipeHandlerGetsDelta(t *testing.T) {
	t.Parallel()

	p := bedPage(t, `<p>hi</p>`, nil)

	var gotX, gotY float64
	p.Handle(Handlers{Swipe: func(_ context.Context, dx, dy float64) error {
		gotX, gotY = dx, dy

		return nil
	}})

	if err := p.Swipe(context.Background(), 3, -40); err != nil {
		t.Fatal(err)
	}

	if gotX != 3 || gotY != -40 {
		t.Fatalf("swipe = %v, %v", gotX, gotY)
	}
}

func TestSwipeWithoutHandlerDrawsNothing(t *testing.T) {
	t.Parallel()

	p := bedPage(t, `<p>hi</p>`, nil)
	p.TakeDirty()
	before := p.Stats().Redraws

	if err := p.Swipe(context.Background(), 0, -60); err != nil {
		t.Fatal(err)
	}

	if after := p.Stats().Redraws; after != before {
		t.Fatalf("redraws = %d, want %d", after, before)
	}

	if _, ok := p.TakeDirty(); ok {
		t.Fatal("a swipe with no handler dirtied the page")
	}
}

func hoverWidth(t *testing.T, p *Page) float64 {
	t.Helper()

	for _, b := range p.Boxes() {
		if b.ID == "e" {
			return b.W
		}
	}

	t.Fatal("no box e")

	return 0
}

func TestHoverPressRelease(t *testing.T) {
	t.Parallel()

	source := `<html><head><style>` +
		`#e{width:100px;height:20px}` +
		`#e:hover{width:120px}` +
		`#e:active{width:140px}` +
		`</style></head><body><div id="e"></div></body></html>`

	p, err := New(Config{HTML: source, Width: 320, Height: 200})
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	if got := hoverWidth(t, p); got != 100 {
		t.Fatalf("plain width = %v", got)
	}

	if err := p.Hover(ctx, 10, 10); err != nil {
		t.Fatal(err)
	}

	if got := hoverWidth(t, p); got != 120 {
		t.Fatalf("hover width = %v", got)
	}

	if err := p.Press(ctx, 10, 10); err != nil {
		t.Fatal(err)
	}

	if got := hoverWidth(t, p); got != 140 {
		t.Fatalf("active width = %v", got)
	}

	if err := p.Release(ctx); err != nil {
		t.Fatal(err)
	}

	if got := hoverWidth(t, p); got != 120 {
		t.Fatalf("released width = %v", got)
	}

	if err := p.Hover(ctx, 300, 150); err != nil {
		t.Fatal(err)
	}

	if got := hoverWidth(t, p); got != 100 {
		t.Fatalf("unhovered width = %v", got)
	}
}
