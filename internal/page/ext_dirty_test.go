package page_test

import (
	"context"
	"github.com/chinmay-sawant/ownframe/internal/page"
	"github.com/chinmay-sawant/ownframe/internal/page/dirty"
	"image"
	"testing"
)

func TestBigDiffFallsBackToFrame(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	screen, err := page.New(page.Config{
		HTML:   `<body style="margin:0">{{range .}}<div style="width:320px;height:40px;background:{{.C}}">x</div>{{end}}</body>`,
		Width:  320,
		Height: 400,
	})
	if err != nil {
		t.Fatal(err)
	}

	rows := make([]struct{ C string }, 10)
	for i := range rows {
		rows[i].C = "#eeeeee"
	}

	screen.SetData(rows)
	if err = screen.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	screen.TakeDirty()

	for i := range rows {
		rows[i].C = "#111111"
	}

	screen.SetData(rows)
	if err = screen.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	rect, ok := screen.TakeDirty()
	if !ok || rect != image.Rect(0, 0, 320, 400) {
		t.Fatalf("rect %v ok %v", rect, ok)
	}
}

func TestOneClickOneLayoutOneRect(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	n := 0
	screen := counterPage(t, &n)
	before := screen.Generation()

	x, y := boxCenter(t, screen, "inc")
	if err := screen.Click(ctx, x, y); err != nil {
		t.Fatal(err)
	}

	if got := screen.Generation(); got != before+1 {
		t.Fatalf("generation %d, want %d", got, before+1)
	}

	rect, ok := screen.TakeDirty()
	if !ok || rect.Empty() {
		t.Fatalf("rect %v ok %v", rect, ok)
	}

	if _, again := screen.TakeDirty(); again {
		t.Fatal("dirty survived the first take")
	}
}

func TestDirtyRectClampsToContent(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	screen, err := page.New(page.Config{
		HTML:   `<body style="margin:0"><p id="t" style="margin-top:80px">{{.V}}</p></body>`,
		Width:  320,
		Height: 100,
	})
	if err != nil {
		t.Fatal(err)
	}

	screen.SetData(struct{ V string }{V: "a"})
	if err = screen.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	screen.TakeDirty()

	screen.SetData(struct{ V string }{V: "b"})
	if err = screen.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	rect, ok := screen.TakeDirty()
	// Text can extend the content below the viewport. Keep that ink dirty.
	frame := image.Rect(0, 0, screen.Display().Width, screen.Display().Height)
	if !ok || rect.Empty() || !rect.In(frame) {
		t.Fatalf("rect %v ok %v", rect, ok)
	}

	if rect.Max.Y <= 100 {
		t.Fatalf("rect %v lost ink below the viewport", rect)
	}
}

func TestFocusTypingAndBlurDirtyField(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	screen := formPage(t, `<input id="t" type="text"`+wide+`><div id="pad" style="height:40px">pad</div>`)

	clickID(t, screen, "t")
	rect, ok := screen.TakeDirty()
	if !ok || !boxIn(rect, boxFor(t, screen, "t")) {
		t.Fatalf("focus rect %v ok %v", rect, ok)
	}

	if err := screen.Type(ctx, "abc"); err != nil {
		t.Fatal(err)
	}

	rect, ok = screen.TakeDirty()
	if !ok || !boxIn(rect, boxFor(t, screen, "t")) {
		t.Fatalf("typing rect %v ok %v", rect, ok)
	}

	clickID(t, screen, "pad")

	rect, ok = screen.TakeDirty()
	if !ok || !boxIn(rect, boxFor(t, screen, "t")) {
		t.Fatalf("blur rect %v ok %v", rect, ok)
	}
}

func TestFallbackRedrawDirtiesFrame(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	screen, err := page.New(page.Config{
		HTML:   `<style>div { width: 60px; height: 40px; background: #123456; border-radius: 50% / 20%; }</style><div>{{.V}}</div>`,
		Width:  320,
		Height: 400,
	})
	if err != nil {
		t.Fatal(err)
	}

	screen.SetData(struct{ V string }{V: "a"})
	if err = screen.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	screen.TakeDirty()

	screen.SetData(struct{ V string }{V: "b"})
	if err = screen.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	rect, ok := screen.TakeDirty()
	if !ok || rect != image.Rect(0, 0, 320, 400) {
		t.Fatalf("rect %v ok %v", rect, ok)
	}
}

func counterPage(t *testing.T, n *int) *page.Page {
	t.Helper()

	screen, err := page.New(page.Config{
		HTML:   `<body style="margin:0"><p id="count">{{.N}}</p><div id="inc" data-action="add">+</div></body>`,
		Width:  320,
		Height: 200,
	})
	if err != nil {
		t.Fatal(err)
	}

	screen.SetData(struct{ N int }{*n})
	screen.Handle(page.Handlers{Click: func(_ context.Context, box page.Box) error {
		if box.Action == "add" {
			*n++
			screen.SetData(struct{ N int }{*n})
		}

		return nil
	}})

	if err := screen.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}

	screen.TakeDirty()

	return screen
}

func boxFor(t *testing.T, screen *page.Page, id string) page.Box {
	t.Helper()

	for _, b := range screen.Boxes() {
		if b.ID == id {
			return b
		}
	}

	t.Fatalf("no box id=%q", id)

	return page.Box{}
}

func insideBox(t *testing.T, rect image.Rectangle, b page.Box, pad int) {
	t.Helper()

	want := image.Rect(int(b.X)-pad, int(b.Y)-pad, int(b.X+b.W)+pad, int(b.Y+b.H)+pad)
	if !rect.In(want) {
		t.Fatalf("rect %v outside %v", rect, want)
	}
}

func boxIn(rect image.Rectangle, b page.Box) bool {
	return image.Rect(int(b.X), int(b.Y), int(b.X+b.W+0.5), int(b.Y+b.H+0.5)).In(rect)
}

func TestHoverDirtiesBothBoxes(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	screen, err := page.New(page.Config{
		HTML: `<body style="margin:0">
<div id="a" style="width:80px;height:40px;background:#eee">a</div>
<div id="b" style="width:80px;height:40px;background:#ddd">b</div>
</body>`,
		Width:  320,
		Height: 200,
	})
	if err != nil {
		t.Fatal(err)
	}

	if err = screen.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	screen.TakeDirty()

	ax, ay := boxCenter(t, screen, "a")
	if err = screen.Hover(ctx, ax, ay); err != nil {
		t.Fatal(err)
	}

	if _, ok := screen.TakeDirty(); !ok {
		t.Fatal("hover into a not dirty")
	}

	bx, by := boxCenter(t, screen, "b")
	if err = screen.Hover(ctx, bx, by); err != nil {
		t.Fatal(err)
	}

	rect, ok := screen.TakeDirty()
	if !ok {
		t.Fatal("hover into b not dirty")
	}

	if !boxIn(rect, boxFor(t, screen, "a")) || !boxIn(rect, boxFor(t, screen, "b")) {
		t.Fatalf("rect %v misses a hovered box", rect)
	}
}

func TestInvalidateCountBoxOnly(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	n := 0
	screen := counterPage(t, &n)
	screen.Invalidate("count")

	if err := screen.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	rect, ok := screen.TakeDirty()
	if !ok {
		t.Fatal("no dirty rect")
	}

	count := boxFor(t, screen, "count")
	insideBox(t, rect, count, 8)

	if !boxIn(rect, count) {
		t.Fatalf("rect %v does not cover #count", rect)
	}
}

func TestInvalidateCoversDescendant(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	screen, err := page.New(page.Config{
		HTML:   `<body style="margin:0"><div id="count"><div id="inner">1</div></div></body>`,
		Width:  320,
		Height: 200,
	})
	if err != nil {
		t.Fatal(err)
	}

	if err = screen.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	screen.TakeDirty()
	screen.Invalidate("count")

	if err = screen.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	rect, ok := screen.TakeDirty()
	if !ok {
		t.Fatal("no dirty rect")
	}

	if !boxIn(rect, boxFor(t, screen, "inner")) {
		t.Fatalf("rect %v does not cover the nested span", rect)
	}
}

func TestInvalidateMissingIsFullFrame(t *testing.T) {
	t.Parallel()

	n := 0
	screen := counterPage(t, &n)
	want := image.Rect(0, 0, 320, 200)

	for _, id := range []string{"", "missing"} {
		screen.Invalidate(id)

		rect, ok := screen.TakeDirty()
		if !ok || rect != want {
			t.Fatalf("id %q: rect %v ok %v", id, rect, ok)
		}
	}
}

// coveredBy reports whether one of the rects contains b.
func coveredBy(rects []image.Rectangle, b image.Rectangle) bool {
	for _, r := range rects {
		if b.In(r) {
			return true
		}
	}

	return false
}

// Twelve dots past the merge gap in a small area must stay a rect list under
// the damage cap instead of overflowing to one frame-sized box.
func TestDiffRectListUnderWindowCap(t *testing.T) {
	t.Parallel()

	screen, err := page.New(page.Config{
		HTML:  `<body style="margin:0">{{range .}}<div style="position:absolute;left:{{.X}}px;top:{{.Y}}px;width:2px;height:2px;background:{{.C}}"></div>{{end}}</body>`,
		Width: mergeW, Height: mergeH,
	})
	if err != nil {
		t.Fatal(err)
	}

	type dot struct {
		X, Y int
		C    string
	}

	dots := make([]dot, 0, 12)
	for y := 0; y < 60; y += 20 {
		for x := 0; x < 60; x += 20 {
			dots = append(dots, dot{x, y, "#eee"})
		}
	}

	screen.SetData(dots)
	if err = screen.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}
	screen.TakeDirtyRects()

	for i := range dots {
		dots[i].C = "#111"
	}

	screen.SetData(dots)
	if err = screen.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}

	rects, ok := screen.TakeDirtyRects()
	if !ok || len(rects) > dirty.MaxRects {
		t.Fatalf("rects %d ok %v over the damage cap", len(rects), ok)
	}

	if isFullFrame(rects, mergeW, mergeH) {
		t.Fatal("small dots fell back to the frame")
	}

	for _, d := range dots {
		b := image.Rect(d.X, d.Y, d.X+2, d.Y+2)
		if !coveredBy(rects, b) {
			t.Fatalf("rects %v miss dot %v", rects, b)
		}
	}
}

// Changes scattered across the frame must still fall back to the frame.
func TestDiffScatteredChangesFallBack(t *testing.T) {
	t.Parallel()

	screen, err := page.New(page.Config{
		HTML:  `<body style="margin:0">{{range .}}<div style="position:absolute;left:{{.X}}px;top:{{.Y}}px;width:20px;height:20px;background:{{.C}}"></div>{{end}}</body>`,
		Width: mergeW, Height: mergeH,
	})
	if err != nil {
		t.Fatal(err)
	}

	type spot struct {
		X, Y int
		C    string
	}

	spots := []spot{{0, 0, "#eee"}, {300, 0, "#eee"}, {0, 180, "#eee"}, {300, 180, "#eee"}, {150, 90, "#eee"}}

	screen.SetData(spots)
	if err = screen.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}
	screen.TakeDirtyRects()

	for i := range spots {
		spots[i].C = "#111"
	}

	screen.SetData(spots)
	if err = screen.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}

	rects, ok := screen.TakeDirtyRects()
	if !ok || !isFullFrame(rects, mergeW, mergeH) {
		t.Fatalf("scattered %v ok %v", rects, ok)
	}
}

// One change covering more than a third of the frame must fall back.
func TestDiffThirdAreaFallsBack(t *testing.T) {
	t.Parallel()

	screen, err := page.New(page.Config{
		HTML:  `<body style="margin:0"><div style="width:320px;height:100px;background:{{.C}}"></div></body>`,
		Width: mergeW, Height: mergeH,
	})
	if err != nil {
		t.Fatal(err)
	}

	screen.SetData(struct{ C string }{C: "#eee"})
	if err = screen.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}
	screen.TakeDirtyRects()

	screen.SetData(struct{ C string }{C: "#111"})
	if err = screen.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}

	rects, ok := screen.TakeDirtyRects()
	if !ok || !isFullFrame(rects, mergeW, mergeH) {
		t.Fatalf("third area %v ok %v", rects, ok)
	}
}

// A removed op must still dirty its old bounds, even when the same redraw
// also changes another op and the change stays partial.
func TestDiffRemovedOpKeepsOldBounds(t *testing.T) {
	t.Parallel()

	screen, err := page.New(page.Config{
		HTML:  `<body style="margin:0">{{if .Show}}<div id="gone" style="width:60px;height:30px;background:#f00">x</div>{{end}}<p id="stay">{{.V}}</p></body>`,
		Width: mergeW, Height: mergeH,
	})
	if err != nil {
		t.Fatal(err)
	}

	screen.SetData(struct {
		Show bool
		V    string
	}{true, "a"})
	if err = screen.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}
	screen.TakeDirtyRects()

	gone := boxFor(t, screen, "gone")

	screen.SetData(struct {
		Show bool
		V    string
	}{false, "b"})
	if err = screen.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}

	rect, ok := screen.TakeDirty()
	if !ok || !boxIn(rect, gone) {
		t.Fatalf("rect %v ok %v does not cover %v", rect, ok, gone)
	}

	if rect == image.Rect(0, 0, mergeW, mergeH) {
		t.Fatal("removal forced the frame")
	}
}

const mergeW, mergeH = 320, 200

// smallChangePage builds the counter page the diff tests mutate.
func smallChangePage(t *testing.T) (*page.Page, []string) {
	t.Helper()

	screen, err := page.New(page.Config{
		HTML:  `<style>div{height:12px;font-size:10px;line-height:12px}</style><body style="margin:0">{{range $i, $v := .}}<div id="c{{$i}}">{{$v}}</div>{{end}}</body>`,
		Width: mergeW, Height: mergeH, Perf: true,
	})
	if err != nil {
		t.Fatal(err)
	}

	vals := make([]string, 20)
	for i := range vals {
		vals[i] = "0"
	}

	screen.SetData(vals)
	if err = screen.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}

	screen.TakeDirtyRects()

	return screen, vals
}

// rectUnion covers every rect.
func rectUnion(rects []image.Rectangle) image.Rectangle {
	r := rects[0]
	for _, b := range rects[1:] {
		r = r.Union(b)
	}

	return r
}

// isFullFrame reports a take of exactly the whole frame.
func isFullFrame(rects []image.Rectangle, w, h int) bool {
	return len(rects) == 1 && rects[0] == image.Rect(0, 0, w, h)
}

// Twenty small changes in one column must stay partial instead of falling
// back to the frame on an op count.
func TestDiffManySmallChangesStaysPartial(t *testing.T) {
	t.Parallel()

	screen, vals := smallChangePage(t)

	for i := range vals {
		vals[i] = "12345"
	}

	screen.SetData(vals)
	if err := screen.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}

	if got := screen.Stats().ChangedOps; got != len(vals) {
		t.Fatalf("changed %d, want %d", got, len(vals))
	}

	rects, ok := screen.TakeDirtyRects()
	if !ok || len(rects) == 0 || len(rects) > 8 {
		t.Fatalf("rects %v ok %v", rects, ok)
	}

	union := rectUnion(rects)
	if union == image.Rect(0, 0, mergeW, mergeH) {
		t.Fatalf("fell back to the frame: %v", rects)
	}

	if union.Dx()*union.Dy() > mergeW*mergeH/3 {
		t.Fatalf("union %v over a third of the frame", union)
	}
}

// The changed text sits below the window frame. TakeDirty must keep its
// box, since the replay buffer covers the content past the canvas.
func TestDirtyRectBelowFrameIsKept(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	screen, err := page.New(page.Config{
		HTML:   `<body style="margin:0"><p id="t" style="margin-top:300px">{{.V}}</p></body>`,
		Width:  320,
		Height: 100,
	})
	if err != nil {
		t.Fatal(err)
	}

	screen.SetData(struct{ V string }{V: "a"})
	if err = screen.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	screen.TakeDirty()

	screen.SetData(struct{ V string }{V: "b"})
	if err = screen.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	rect, ok := screen.TakeDirty()
	if !ok || rect.Empty() {
		t.Fatalf("rect %v ok %v", rect, ok)
	}

	if rect.Min.Y <= 100 {
		t.Fatalf("rect %v collapsed to the frame", rect)
	}

	b := boxFor(t, screen, "t")
	if rect.Min.Y > int(b.Y+b.H+0.5) || rect.Max.Y < int(b.Y) {
		t.Fatalf("rect %v misses the changed text at %v", rect, b)
	}
}

func TestRemovedOpDirtiesOldBounds(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	screen, err := page.New(page.Config{
		HTML:   `<body style="margin:0">{{if .Show}}<div id="gone" style="width:60px;height:30px;background:#f00">x</div>{{end}}<p id="stay">stay</p></body>`,
		Width:  320,
		Height: 200,
	})
	if err != nil {
		t.Fatal(err)
	}

	screen.SetData(struct{ Show bool }{true})
	if err = screen.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	screen.TakeDirty()

	gone := boxFor(t, screen, "gone")
	screen.SetData(struct{ Show bool }{false})

	if err = screen.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	rect, ok := screen.TakeDirty()
	if !ok {
		t.Fatal("no dirty rect")
	}

	if !boxIn(rect, gone) {
		t.Fatalf("rect %v does not cover the removed box", rect)
	}
}

func TestIdenticalRedrawNoDirty(t *testing.T) {
	t.Parallel()

	n := 0
	screen := counterPage(t, &n)
	if err := screen.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}

	if _, ok := screen.TakeDirty(); ok {
		t.Fatal("identical redraw reported dirty")
	}
}

func TestClickDiffFindsCountBox(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	n := 0
	screen := counterPage(t, &n)
	x, y := boxCenter(t, screen, "inc")
	if err := screen.Click(ctx, x, y); err != nil {
		t.Fatal(err)
	}

	rect, ok := screen.TakeDirty()
	if !ok {
		t.Fatal("no dirty rect")
	}

	insideBox(t, rect, boxFor(t, screen, "count"), 8)
}

func TestThemeDirtiesFullFrame(t *testing.T) {
	t.Parallel()

	n := 0
	screen := counterPage(t, &n)
	if err := screen.SetTheme("body { background: #000; }"); err != nil {
		t.Fatal(err)
	}

	if err := screen.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}

	rect, ok := screen.TakeDirty()
	if !ok || rect != image.Rect(0, 0, 320, 200) {
		t.Fatalf("rect %v ok %v", rect, ok)
	}
}

func TestToggleDirtiesControlAndLabel(t *testing.T) {
	t.Parallel()

	screen := formPage(t, `<label id="lab" style="display:block">Agree</label><input id="agree" type="checkbox">`)

	clickID(t, screen, "agree")

	rect, ok := screen.TakeDirty()
	if !ok {
		t.Fatal("no dirty rect")
	}

	if !boxIn(rect, boxFor(t, screen, "agree")) || !boxIn(rect, boxFor(t, screen, "lab")) {
		t.Fatalf("rect %v misses the control or the label", rect)
	}
}

func TestDiffCoversWrongDeclared(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	screen, err := page.New(page.Config{
		HTML:   `<body style="margin:0"><p id="count">{{.N}}</p><div id="inc" data-action="add">+</div></body>`,
		Width:  320,
		Height: 200,
	})
	if err != nil {
		t.Fatal(err)
	}

	screen.SetData(struct{ N int }{0})
	screen.Handle(page.Handlers{Click: func(_ context.Context, box page.Box) error {
		if box.Action == "add" {
			screen.SetData(struct{ N int }{1})
			screen.Invalidate("inc")
		}

		return nil
	}})

	if err = screen.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	screen.TakeDirty()

	x, y := boxCenter(t, screen, "inc")
	if err = screen.Click(ctx, x, y); err != nil {
		t.Fatal(err)
	}

	rect, ok := screen.TakeDirty()
	if !ok {
		t.Fatal("no dirty rect")
	}

	count := boxFor(t, screen, "count")
	center := image.Pt(int(count.X+count.W/2), int(count.Y+count.H/2))
	if !center.In(rect) {
		t.Fatalf("rect %v misses the diffed #count", rect)
	}
}

func TestPartialVsFullBounds(t *testing.T) {
	t.Parallel()

	t.Run("partial", func(t *testing.T) {
		t.Parallel()
		screen, vals := smallChangePage(t)
		vals[0] = "9"
		screen.SetData(vals)
		if err := screen.Redraw(context.Background()); err != nil {
			t.Fatal(err)
		}
		rects, ok := screen.TakeDirtyRects()
		if !ok {
			t.Fatal("no dirty")
		}
		if isFullFrame(rects, mergeW, mergeH) {
			t.Fatalf("small change became full: %v", rects)
		}
	})

	t.Run("full", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		screen, err := page.New(page.Config{
			HTML:  `<body style="margin:0">{{range .}}<div style="width:320px;height:40px;background:{{.C}}">x</div>{{end}}</body>`,
			Width: 320, Height: 400,
		})
		if err != nil {
			t.Fatal(err)
		}
		rows := make([]struct{ C string }, 10)
		for i := range rows {
			rows[i].C = "#eeeeee"
		}
		screen.SetData(rows)
		if err = screen.Redraw(ctx); err != nil {
			t.Fatal(err)
		}
		screen.TakeDirty()
		for i := range rows {
			rows[i].C = "#111111"
		}
		screen.SetData(rows)
		if err = screen.Redraw(ctx); err != nil {
			t.Fatal(err)
		}
		rect, ok := screen.TakeDirty()
		if !ok || rect != image.Rect(0, 0, 320, 400) {
			t.Fatalf("rect %v ok %v", rect, ok)
		}
	})
}

func flexAbsSource(display string) string {
	return `<html><head><style>body{margin:0}` +
		`#row{` + display + `;width:200px;height:80px}` +
		`#abs{position:absolute;top:10px;left:40px;width:20px;height:20px}` +
		`#sib{width:50px;height:50px}` +
		`</style></head><body>` +
		`<div id="row"><div id="abs"></div><div id="sib"></div></div>` +
		`</body></html>`
}

// TestFlexChildAbsoluteOverlays pins that position:absolute takes a flex
// child out of flow: it paints at its static-plus-offset position and the
// sibling lays out as if it were absent. Block containers behave the same.
func TestFlexChildAbsoluteOverlays(t *testing.T) {
	t.Parallel()

	for _, display := range []string{"display:flex", "position:relative;display:block"} {
		p := newViewportPage(t, flexAbsSource(display), 200, 100)
		abs := viewportBox(t, p, "abs")
		sib := viewportBox(t, p, "sib")

		if abs.X != 40 || abs.Y != 10 {
			t.Fatalf("%s abs = %.0f,%.0f, want 40,10", display, abs.X, abs.Y)
		}

		if sib.X != 0 {
			t.Fatalf("%s sib X = %.0f, want 0 (abs out of flow)", display, sib.X)
		}
	}
}

func TestStickyViewportScrollContract(t *testing.T) {
	t.Parallel()
	source := `<style>body{margin:0}#before{height:80px}#region{height:300px}#head{position:sticky;top:0;height:40px;background:red;z-index:2}#body{height:1000px;background:blue}</style><div id="before"></div><div id="region"><div id="head" data-action="activate" role="button" tabindex="0"></div><div id="body"></div></div>`
	p := newViewportPage(t, source, 200, 100)
	for _, test := range []struct{ scroll, want int }{{0, 80}, {100, 100}, {200, 200}, {340, 340}, {360, 340}, {0, 80}} {
		before := p.Stats()
		p.SetScrollOffset(0, test.scroll)
		head := viewportBox(t, p, "head")
		if head.Y != float64(test.want) {
			t.Fatalf("scroll %d: hit box Y %g want %d", test.scroll, head.Y, test.want)
		}
		if p.Stats().Layouts != before.Layouts {
			t.Fatal("sticky scroll performed a layout")
		}
		if test.scroll == 200 {
			hit, ok := p.BoxAt(10, 210)
			if !ok || hit.ID != "head" {
				t.Fatalf("sticky paint/hit overlap target %+v", hit)
			}
		}
	}
}
