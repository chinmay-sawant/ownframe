package page_test

import (
	"bytes"
	"context"
	"errors"
	"github.com/chinmay-sawant/ownframe/internal/page"
	"image/png"
	"testing"
)

func TestAddHandlersKeepsPrior(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	p, err := page.New(page.Config{HTML: `<div>hi</div>`, Width: 200, Height: 120})
	if err != nil {
		t.Fatal(err)
	}
	var down, up int
	p.Handle(page.Handlers{
		KeyDown: func(context.Context, string) error { down++; return nil },
	})
	p.AddHandlers(page.Handlers{
		KeyUp: func(context.Context, string) error { up++; return nil },
	})
	if err := p.KeyDown(ctx, "a"); err != nil {
		t.Fatal(err)
	}
	if err := p.KeyUp(ctx, "a"); err != nil {
		t.Fatal(err)
	}
	if down != 1 || up != 1 {
		t.Fatalf("down=%d up=%d", down, up)
	}
	p.AddHandlers(page.Handlers{})
	if err := p.KeyDown(ctx, "b"); err != nil {
		t.Fatal(err)
	}
	if down != 2 {
		t.Fatalf("down=%d", down)
	}
}

// TestClickReachesControlThroughChild checks a click that lands on an
// anonymous child falls back to the innermost id-bearing element, so a
// child icon does not swallow the control's action.
func TestClickReachesControlThroughChild(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	screen, err := page.New(page.Config{
		HTML: `<body style="margin:0"><div id="go" data-action="go" style="width:100px;height:40px">` +
			`<div style="width:20px;height:20px">hit</div></div></body>`,
		Width:  320,
		Height: 400,
	})
	if err != nil {
		t.Fatal(err)
	}

	got := ""
	screen.Handle(page.Handlers{
		Click: func(_ context.Context, box page.Box) error {
			got = box.ID + ":" + box.Action

			return nil
		},
	})

	if err := screen.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	child := page.Box{}

	for _, box := range screen.Boxes() {
		if box.Tag == "div" && box.ID == "" && box.W > 0 {
			child = box
		}
	}

	if child.W <= 0 {
		t.Fatal("no child div box")
	}

	if err := screen.Click(ctx, child.X+child.W/2, child.Y+child.H/2); err != nil {
		t.Fatal(err)
	}

	if got != "go:go" {
		t.Fatalf("click reached %q, want the control go:go", got)
	}
}

// TestClickPaintEndsBatch checks that a paint-only click inside a redraw
// batch leaves no draw pending, so the batch waits for nothing.
func TestClickPaintEndsBatch(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	screen := formPage(t, `<button id="go">Go</button>`)
	screen.Handle(page.Handlers{
		ClickPaint: func(context.Context, page.Box) (bool, error) { return true, nil },
	})

	before := screen.Stats()
	screen.BeginRedrawBatch()
	if err := screen.Focus(ctx, "go"); err != nil {
		t.Fatal(err)
	}
	clickID(t, screen, "go")
	if err := screen.EndRedrawBatch(ctx); err != nil {
		t.Fatal(err)
	}

	after := screen.Stats()
	if after.Redraws != before.Redraws {
		t.Fatalf("redraws = %d, want %d", after.Redraws, before.Redraws)
	}
	if after.Layouts != before.Layouts {
		t.Fatalf("layouts = %d, want %d", after.Layouts, before.Layouts)
	}
	if got := screen.FocusID(); got != "go" {
		t.Fatalf("focus = %q, want go", got)
	}
}

// TestClickPaintRunsAfterClick checks that the paint handler runs after the
// click handler and after the control changed.
func TestClickPaintRunsAfterClick(t *testing.T) {
	t.Parallel()

	screen := formPage(t, `<input id="c" type="checkbox">`)
	var order []string
	screen.Handle(page.Handlers{
		Click: func(context.Context, page.Box) error {
			order = append(order, "click")
			return nil
		},
		ClickPaint: func(context.Context, page.Box) (bool, error) {
			order = append(order, "paint")
			if !screen.FormChecked("c") {
				t.Error("the control did not change before the paint handler")
			}
			return true, nil
		},
	})

	clickID(t, screen, "c")

	if len(order) != 2 || order[0] != "click" || order[1] != "paint" {
		t.Fatalf("order = %v, want click then paint", order)
	}
}

// TestClickPaintKeepsDamage checks that the damage the handler marked stays
// for the window after a paint-only click, so the window repaints it.
func TestClickPaintKeepsDamage(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	screen := newCachePage(t, `<p id="p" style="display:block;width:80px;height:20px">x</p>`)
	screen.Handle(page.Handlers{
		Click: func(context.Context, page.Box) error {
			screen.Invalidate("p")
			return nil
		},
		ClickPaint: func(context.Context, page.Box) (bool, error) { return true, nil },
	})
	if err := screen.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	screen.TakeDirty()

	gen := screen.Generation()
	clickID(t, screen, "p")

	if got := screen.Generation(); got != gen {
		t.Fatalf("generation = %d, want %d", got, gen)
	}

	rect, ok := screen.TakeDirty()
	if !ok {
		t.Fatal("no dirty rect after the paint-only click")
	}
	if !boxIn(rect, boxFor(t, screen, "p")) {
		t.Fatalf("rect %v misses the box", rect)
	}
}

// TestClickPaintPlainClick checks that a click on a plain box with a paint
// handler skips the redraw too.
func TestClickPaintPlainClick(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	screen := newCachePage(t, `<div id="d">{{.N}}</div>`)
	screen.SetData(struct{ N int }{1})
	screen.Handle(page.Handlers{
		ClickPaint: func(context.Context, page.Box) (bool, error) { return true, nil },
	})
	if err := screen.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	before := screen.Stats()
	gen := screen.Generation()
	clickID(t, screen, "d")

	if got := screen.Stats().Redraws; got != before.Redraws {
		t.Fatalf("redraws = %d, want %d", got, before.Redraws)
	}
	if got := screen.Generation(); got != gen {
		t.Fatalf("generation = %d, want %d", got, gen)
	}
}

// TestClickPaintFalseRedraws checks that a click whose paint handler does
// not paint still runs the redraw sequence. The click changes no page input,
// so the identical frame commits nothing.
func TestClickPaintFalseRedraws(t *testing.T) {
	t.Parallel()

	screen := formPage(t, `<button id="go">Go</button>`)
	screen.Handle(page.Handlers{
		ClickPaint: func(context.Context, page.Box) (bool, error) { return false, nil },
	})

	before := screen.Stats()
	gen := screen.Generation()
	clickID(t, screen, "go")

	if got := screen.Stats().Redraws; got != before.Redraws+1 {
		t.Fatalf("redraws = %d, want %d", got, before.Redraws+1)
	}
	if got := screen.Generation(); got != gen {
		t.Fatalf("generation = %d, want %d", got, gen)
	}
}

// TestClickPaintSkipsRedraw checks that a click whose paint handler says it
// painted leaves the generation and the redraw counter alone, keeps the
// focus, and lets a later state change redraw again.
func TestClickPaintSkipsRedraw(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	screen := newCachePage(t, `<button id="go">{{.N}}</button>`)
	screen.SetData(struct{ N int }{1})
	calls := 0
	screen.Handle(page.Handlers{
		ClickPaint: func(context.Context, page.Box) (bool, error) {
			calls++
			return calls == 1, nil
		},
	})
	if err := screen.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	before := screen.Stats()
	gen := screen.Generation()
	clickID(t, screen, "go")

	after := screen.Stats()
	if after.Redraws != before.Redraws {
		t.Fatalf("redraws = %d, want %d", after.Redraws, before.Redraws)
	}
	if after.Layouts != before.Layouts {
		t.Fatalf("layouts = %d, want %d", after.Layouts, before.Layouts)
	}
	if got := screen.Generation(); got != gen {
		t.Fatalf("generation = %d, want %d", got, gen)
	}
	if got := screen.FocusID(); got != "go" {
		t.Fatalf("focus = %q, want go", got)
	}

	// A later event that changes the state redraws again.
	screen.SetData(struct{ N int }{2})
	clickID(t, screen, "go")

	if got := screen.Stats().Redraws; got != before.Redraws+1 {
		t.Fatalf("redraws = %d, want %d", got, before.Redraws+1)
	}
	if got := screen.Generation(); got <= gen {
		t.Fatalf("generation = %d, want past %d", got, gen)
	}
}

func TestClickRewritesThePicture(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	screen, err := page.New(page.Config{
		HTML:   `<div id="n" data-action="add">{{.N}}</div>`,
		Width:  320,
		Height: 400,
	})
	if err != nil {
		t.Fatal(err)
	}

	data := struct{ N int }{N: 1}
	screen.Handle(page.Handlers{
		Click: func(_ context.Context, box page.Box) error {
			if box.Action == "add" {
				data.N++
			}

			screen.SetData(data)

			return nil
		},
	})
	screen.SetData(data)

	if err := screen.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	before := screen.Generation()
	x, y := boxCenter(t, screen, "n")
	if err := screen.Click(ctx, x, y); err != nil {
		t.Fatal(err)
	}

	if screen.Generation() <= before {
		t.Fatalf("generation = %d, before = %d", screen.Generation(), before)
	}

	cfg, err := png.DecodeConfig(bytes.NewReader(screen.PNG()))
	if err != nil {
		t.Fatal(err)
	}

	if cfg.Width < 320 || cfg.Height < 400 {
		t.Fatalf("png = %d x %d", cfg.Width, cfg.Height)
	}

	if got := boxText(t, screen, "n"); got != "2" {
		t.Fatalf("text = %q", got)
	}
}

func TestDropErrorSkipsRedraw(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	screen, err := page.New(page.Config{HTML: `<p>hi</p>`, Width: 200, Height: 120})
	if err != nil {
		t.Fatal(err)
	}

	want := errors.New("drop refused")
	screen.Handle(page.Handlers{
		Drop: func(context.Context, []page.Drop) error { return want },
	})

	if err := screen.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	before := screen.Generation()

	if err := screen.Drop(ctx, []page.Drop{{Name: "a.txt"}}); !errors.Is(err, want) {
		t.Fatalf("err = %v", err)
	}

	if screen.Generation() != before {
		t.Fatalf("generation = %d, before = %d", screen.Generation(), before)
	}
}

func TestDropRedrawsWithoutHandler(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	screen, err := page.New(page.Config{HTML: `<p>hi</p>`, Width: 200, Height: 120})
	if err != nil {
		t.Fatal(err)
	}

	if err := screen.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	before := screen.Generation()

	if err := screen.Drop(ctx, []page.Drop{{Name: "a.txt"}}); err != nil {
		t.Fatal(err)
	}

	// No handler changed page input, so the redraw runs but commits an
	// identical frame: one more redraw, no new generation.
	if got := screen.Stats().Redraws; got != 2 {
		t.Fatalf("redraws = %d, want 2", got)
	}
	if screen.Generation() != before {
		t.Fatalf("generation = %d, before = %d", screen.Generation(), before)
	}
}

func TestDropHandlerReceivesNameAndBytes(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	screen, err := page.New(page.Config{
		HTML:   `<p id="line">{{.}}</p>`,
		Width:  240,
		Height: 160,
	})
	if err != nil {
		t.Fatal(err)
	}

	var name string
	var data []byte

	screen.Handle(page.Handlers{
		Drop: func(_ context.Context, files []page.Drop) error {
			if len(files) != 1 {
				t.Fatalf("files = %d, want 1", len(files))
			}

			name = files[0].Name
			data, err = files[0].Read()

			return err
		},
	})

	if err := screen.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	before := screen.Generation()

	files := []page.Drop{{
		Name: "note.txt",
		Size: 5,
		Read: func() ([]byte, error) { return []byte("hello"), nil },
	}}

	if err := screen.Drop(ctx, files); err != nil {
		t.Fatal(err)
	}

	if name != "note.txt" || string(data) != "hello" {
		t.Fatalf("name = %q, data = %q", name, data)
	}

	// The handler changed no page input, so the redraw runs but commits an
	// identical frame: one more redraw, no new generation.
	if got := screen.Stats().Redraws; got != 2 {
		t.Fatalf("redraws = %d, want 2", got)
	}
	if screen.Generation() != before {
		t.Fatalf("generation = %d, before = %d", screen.Generation(), before)
	}
}

func TestKeyHandlersRunWithoutDrawing(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	screen, err := page.New(page.Config{HTML: `<p>hi</p>`, Width: 80, Height: 40})
	if err != nil {
		t.Fatal(err)
	}

	if err := screen.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	var got []string

	screen.Handle(page.Handlers{
		KeyDown: func(_ context.Context, key string) error {
			got = append(got, "down "+key)

			return nil
		},
		KeyUp: func(_ context.Context, key string) error {
			got = append(got, "up "+key)

			return nil
		},
	})

	before := screen.Generation()
	if err := screen.KeyDown(ctx, "space"); err != nil {
		t.Fatal(err)
	}

	if err := screen.KeyUp(ctx, "space"); err != nil {
		t.Fatal(err)
	}

	if len(got) != 2 || got[0] != "down space" || got[1] != "up space" {
		t.Fatalf("handlers saw %q", got)
	}

	if screen.Generation() != before {
		t.Fatal("a key handler drew the page")
	}
}

func TestNilKeyHandlersDoNothing(t *testing.T) {
	t.Parallel()

	screen, err := page.New(page.Config{HTML: `<p>hi</p>`, Width: 80, Height: 40})
	if err != nil {
		t.Fatal(err)
	}

	if err := screen.KeyDown(context.Background(), "space"); err != nil {
		t.Fatal(err)
	}

	if err := screen.KeyUp(context.Background(), "space"); err != nil {
		t.Fatal(err)
	}
}
