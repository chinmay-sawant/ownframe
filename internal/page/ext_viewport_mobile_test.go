package page_test

import (
	"context"
	"errors"
	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/page"
	"os"
	"testing"
)

func TestAndroidLayoutsResizeAndModalInput(t *testing.T) {
	source, err := os.ReadFile("testdata/android-runtime.html")
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	p, err := page.New(page.Config{HTML: string(source), Width: 360, Height: 640})
	if err != nil {
		t.Fatal(err)
	}
	data := struct{ Email string }{}
	p.SetData(&data)
	clicked := ""
	p.Handle(page.Handlers{Click: func(_ context.Context, b page.Box) error { clicked = b.Action; return nil }})
	for _, size := range [][2]int{{360, 640}, {640, 360}, {412, 915}, {360, 640}} {
		p.SetSize(size[0], size[1])
		if err := p.Redraw(ctx); err != nil {
			t.Fatal(err)
		}
		if p.Display() == nil {
			t.Fatal("basic layout fell back")
		}
		a, b := boxFor(t, p, "flex-a"), boxFor(t, p, "flex-b")
		if size[0] <= 400 && b.Y <= a.Y || size[0] > 400 && b.X <= a.X {
			t.Fatal("responsive flex direction")
		}
		if boxFor(t, p, "grid-b").X <= boxFor(t, p, "grid-a").X {
			t.Fatal("grid columns")
		}
		if boxFor(t, p, "settings").H < 900 {
			t.Fatal("long content truncated")
		}
		modal := boxFor(t, p, "modal")
		if err := p.Click(ctx, modal.X+modal.W/2, modal.Y+modal.H/2); err != nil {
			t.Fatal(err)
		}
		if clicked != "modal" {
			t.Fatal("overlay did not receive input")
		}
	}
	if err := p.Focus(ctx, "email"); err != nil {
		t.Fatal(err)
	}
	if err := p.Type(ctx, "mobile@example.test"); err != nil {
		t.Fatal(err)
	}
	if data.Email != "mobile@example.test" {
		t.Fatal("form binding")
	}
	p.SetSize(640, 360)
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	if p.FocusedField() != "email" || p.FormValue("email") != data.Email {
		t.Fatal("resize lost focus or value")
	}
}

func TestDesktopMinimumSurvivesMobilePrepareError(t *testing.T) {
	p, err := page.New(page.Config{HTML: `<body>desktop</body>`, Width: 420, Height: 934, MinWidth: 320, MinHeight: 480})
	if err != nil {
		t.Fatal(err)
	}
	if err := page.Prepare(context.Background(), p); err != nil {
		t.Fatal(err)
	}
	//nolint:staticcheck // reason: intentionally passes nil to verify nil-context rejection.
	if err := page.PrepareMobile(nil, p); err == nil {
		t.Fatal("nil context accepted")
	}
	p.SetSize(100, 100)
	if w, h := p.Size(); w != 320 || h != 480 {
		t.Fatalf("desktop minimum became %dx%d", w, h)
	}
}

func TestMobilePrepareRejectsNilPage(t *testing.T) {
	if err := page.PrepareMobile(context.Background(), nil); !errors.Is(err, page.ErrNilPage) {
		t.Fatalf("nil page error = %v", err)
	}
}

func TestMobileViewportUsesLandscapeHeight(t *testing.T) {
	p, err := page.New(page.Config{HTML: `<body>mobile</body>`, Width: 420, Height: 934, MinWidth: 320, MinHeight: 480})
	if err != nil {
		t.Fatal(err)
	}
	if err := page.PrepareMobile(context.Background(), p); err != nil {
		t.Fatal(err)
	}
	for _, size := range [][2]int{{420, 934}, {885, 420}, {420, 934}} {
		p.SetSize(size[0], size[1])
		w, h := p.Size()
		if w != size[0] || h != size[1] {
			t.Fatalf("native viewport %v became %dx%d", size, w, h)
		}
	}
}

func TestMobileLockedViewKeepsCanvasMinimum(t *testing.T) {
	p, err := page.New(page.Config{HTML: `<body>game</body>`, Width: 1280, Height: 720, MinWidth: 1280, MinHeight: 720, LockView: true})
	if err != nil {
		t.Fatal(err)
	}
	if err := page.PrepareMobile(context.Background(), p); err != nil {
		t.Fatal(err)
	}
	p.SetSize(420, 320)
	if w, h := p.Size(); w != 1280 || h != 720 {
		t.Fatalf("locked canvas became %dx%d", w, h)
	}
}

// viewportBox returns the box with an id after a relayout.
func viewportBox(t *testing.T, p *page.Page, id string) layout.Box {
	t.Helper()

	for _, b := range p.Boxes() {
		if b.ID == id {
			return b
		}
	}

	t.Fatalf("no box id=%q", id)

	return layout.Box{}
}

// newViewportPage draws source at width x height.
func newViewportPage(t *testing.T, source string, width, height int) *page.Page {
	t.Helper()

	p, err := page.New(page.Config{HTML: source, Width: width, Height: height})
	if err != nil {
		t.Fatal(err)
	}

	if err := p.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}

	return p
}

func TestMediaOrientationFollowsTheFrame(t *testing.T) {
	t.Parallel()

	source := `<html><head><style>body{margin:0}` +
		`#a{width:40px}` +
		`@media (orientation: landscape){#a{width:120px}}` +
		`</style></head><body><div id="a"></div></body></html>`

	p := newViewportPage(t, source, 800, 600)
	if got := viewportBox(t, p, "a").W; got != 120 {
		t.Fatalf("landscape width = %.0f, want 120", got)
	}

	p.SetSize(400, 600)
	if err := p.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}

	if got := viewportBox(t, p, "a").W; got != 40 {
		t.Fatalf("portrait width = %.0f, want 40", got)
	}
}

func TestPercentAndFullHeightFollowTheFrame(t *testing.T) {
	t.Parallel()

	source := `<html><head><style>body{margin:0}` +
		`#outer{width:50%;height:300px}` +
		`#inner{width:50%;height:100%}` +
		`</style></head><body><div id="outer"><div id="inner"></div></div></body></html>`

	p := newViewportPage(t, source, 800, 600)
	if box := viewportBox(t, p, "inner"); box.W != 200 || box.H != 300 {
		t.Fatalf("inner = %.0f x %.0f, want 200 x 300", box.W, box.H)
	}

	p.SetSize(400, 600)
	if err := p.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}

	if box := viewportBox(t, p, "inner"); box.W != 100 || box.H != 300 {
		t.Fatalf("inner after relayout = %.0f x %.0f, want 100 x 300", box.W, box.H)
	}
}

// hasFill reports a fill operation of that color in the display list.
func hasFill(p *page.Page, r, g, b float64) bool {
	d := p.Display()
	if d == nil {
		return false
	}

	for i := range d.Ops {
		op := &d.Ops[i]
		if op.Kind == layout.DisplayOpFillRect && op.R == r && op.G == g && op.B == b {
			return true
		}
	}

	return false
}

func TestFocusAndCheckedSurviveARelayout(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	source := `<html><head><style>body{margin:0}` +
		`#name{width:120px;height:20px}` +
		`#name:focus{background:#ff0000}` +
		`#name:focus-visible{background:#0000ff}` +
		`#ok{width:30px;height:20px}` +
		`#ok:checked{background:#00ff00}` +
		`</style></head><body>` +
		`<input id="name"><input id="ok" type="checkbox">` +
		`</body></html>`

	p := newViewportPage(t, source, 320, 200)

	x, y := boxCenter(t, p, "name")
	if err := p.Click(ctx, x, y); err != nil {
		t.Fatal(err)
	}

	if p.FocusedField() != "name" {
		t.Fatalf("focus = %q, want name", p.FocusedField())
	}

	p.SetFormChecked("ok", true)
	p.SetSize(640, 400)

	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	if !hasFill(p, 0, 0, 1) {
		t.Fatal("focus-visible style lost after the relayout")
	}

	if !hasFill(p, 0, 1, 0) {
		t.Fatal("checked style lost after the relayout")
	}
}

func TestMediaWidthFollowsTheFrame(t *testing.T) {
	t.Parallel()

	source := `<html><head><style>body{margin:0}` +
		`#a{width:40px;height:30px}` +
		`@media (min-width: 700px){#a{width:120px}}` +
		`</style></head><body><div id="a"></div></body></html>`

	p := newViewportPage(t, source, 800, 600)
	if got := viewportBox(t, p, "a").W; got != 120 {
		t.Fatalf("wide width = %.0f, want 120", got)
	}

	p.SetSize(400, 300)
	if err := p.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}

	if got := viewportBox(t, p, "a").W; got != 40 {
		t.Fatalf("narrow width = %.0f, want 40", got)
	}
}

func TestMediaHeightFollowsTheFrame(t *testing.T) {
	t.Parallel()

	source := `<html><head><style>body{margin:0}#a{height:30px}` +
		`@media (min-height:500px){#a{height:90px}}` +
		`@media (max-height:400px){#a{height:60px}}` +
		`</style></head><body><div id="a"></div></body></html>`
	p := newViewportPage(t, source, 800, 499)
	for _, test := range []struct{ height, want int }{{499, 30}, {500, 90}, {501, 90}, {500, 90}, {499, 30}, {401, 30}, {400, 60}, {399, 60}, {400, 60}, {401, 30}, {600, 90}} {
		before := p.Stats()
		p.SetSize(800, test.height)
		if err := p.Redraw(t.Context()); err != nil {
			t.Fatal(err)
		}
		if got := viewportBox(t, p, "a").H; got != float64(test.want) {
			t.Fatalf("viewport height %d: box %g want %d", test.height, got, test.want)
		}
		after := p.Stats()
		if after.Parses != before.Parses || after.Cascades != before.Cascades {
			t.Fatal("height resize reparsed or recascaded document")
		}
	}

}

func TestThemeMediaQueryFollowsTheFrame(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	// #a is the template sheet's first rule, so the theme rule ties its
	// index and wins. A theme rule behind the template rule's index would
	// lose the tie whatever the media query says.
	source := `<html><head><style>` +
		`#a{width:60px;height:10px}` +
		`</style></head><body><div id="a"></div></body></html>`

	p, err := page.New(page.Config{
		HTML:  source,
		Theme: `@media (min-width: 700px){#a{width:150px}}`,
		Width: 800, Height: 600,
	})
	if err != nil {
		t.Fatal(err)
	}

	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	if got := viewportBox(t, p, "a").W; got != 150 {
		t.Fatalf("wide theme width = %.0f, want 150", got)
	}

	p.SetSize(400, 600)
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	if got := viewportBox(t, p, "a").W; got != 60 {
		t.Fatalf("narrow theme width = %.0f, want 60", got)
	}
}

func TestViewportUnitsFollowTheFrame(t *testing.T) {
	t.Parallel()

	source := `<html><head><style>body{margin:0}` +
		`#w{width:50vw;height:1px}` +
		`#vh{height:20vh} #dvh{height:20dvh}` +
		`#svh{height:20svh} #lvh{height:20lvh}` +
		`</style></head><body>` +
		`<div id="w"></div><div id="vh"></div><div id="dvh"></div>` +
		`<div id="svh"></div><div id="lvh"></div></body></html>`

	p := newViewportPage(t, source, 800, 600)
	if got := viewportBox(t, p, "w").W; got != 400 {
		t.Fatalf("50vw = %.0f, want 400", got)
	}

	for _, id := range []string{"vh", "dvh", "svh", "lvh"} {
		if got := viewportBox(t, p, id).H; got != 120 {
			t.Fatalf("%s = %.0f, want 120", id, got)
		}
	}
}

func TestVarChainReadingAViewportUnitRecomputes(t *testing.T) {
	t.Parallel()

	source := `<html><head><style>body{margin:0}` +
		`:root{--a:var(--b);--b:30vw}` +
		`#a{width:var(--a);height:10px}` +
		`</style></head><body><div id="a"></div></body></html>`

	p := newViewportPage(t, source, 800, 600)
	if got := viewportBox(t, p, "a").W; got != 240 {
		t.Fatalf("wide var width = %.0f, want 240", got)
	}

	p.SetSize(400, 600)
	if err := p.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}

	if got := viewportBox(t, p, "a").W; got != 120 {
		t.Fatalf("narrow var width = %.0f, want 120", got)
	}
}
