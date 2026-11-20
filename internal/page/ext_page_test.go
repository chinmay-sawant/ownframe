package page_test

import (
	"context"
	"errors"
	"github.com/chinmay-sawant/ownframe/internal/host"
	"github.com/chinmay-sawant/ownframe/internal/page"
	"image"
	"testing"
)

func TestNewRejectsEmptyHTML(t *testing.T) {
	t.Parallel()

	_, err := page.New(page.Config{Width: 320, Height: 400})
	if !errors.Is(err, page.ErrEmptyHTML) {
		t.Fatalf("err = %v", err)
	}
}

func TestSetSizeClamps(t *testing.T) {
	t.Parallel()

	screen, err := page.New(page.Config{
		HTML:      "<p>Hi</p>",
		Width:     480,
		Height:    640,
		MinWidth:  320,
		MinHeight: 400,
		MaxWidth:  800,
		MaxHeight: 900,
	})
	if err != nil {
		t.Fatal(err)
	}

	screen.SetSize(10, 10)
	width, height := screen.Size()
	if width != 320 || height != 400 {
		t.Fatalf("min clamp = %d x %d", width, height)
	}

	// The max is a window bound: the frame follows the window past it.
	screen.SetSize(5000, 5000)
	width, height = screen.Size()
	if width != 5000 || height != 5000 {
		t.Fatalf("frame past the max = %d x %d", width, height)
	}

	maxW, maxH := screen.MaxSize()
	if maxW != 800 || maxH != 900 {
		t.Fatalf("window bound = %d x %d", maxW, maxH)
	}
}

func TestNewRejectsMinAboveMax(t *testing.T) {
	t.Parallel()

	_, err := page.New(page.Config{
		HTML:      "<p>Hi</p>",
		Width:     480,
		Height:    640,
		MinWidth:  900,
		MaxWidth:  800,
		MinHeight: 400,
		MaxHeight: 900,
	})
	if !errors.Is(err, page.ErrBadSize) {
		t.Fatalf("err = %v, want ErrBadSize", err)
	}
}

func boxCenter(t *testing.T, screen *page.Page, id string) (float64, float64) {
	t.Helper()

	for _, box := range screen.Boxes() {
		if box.ID != id || box.W <= 0 || box.H <= 0 {
			continue
		}

		return box.X + box.W/2, box.Y + box.H/2
	}

	t.Fatalf("no box id=%q", id)

	return 0, 0
}

func boxText(t *testing.T, screen *page.Page, id string) string {
	t.Helper()

	for _, box := range screen.Boxes() {
		if box.ID == id {
			return box.Text
		}
	}

	t.Fatalf("no box id=%q", id)

	return ""
}

// newCachePage builds a page for the cache tests.
func newCachePage(t *testing.T, html string) *page.Page {
	t.Helper()

	p, err := page.New(page.Config{HTML: html, Width: 320, Height: 200})
	if err != nil {
		t.Fatal(err)
	}

	return p
}

func TestResizeReusesParseAndCascade(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	p := newCachePage(t, `<p>hello cache</p>`)

	for i := 0; i < 10; i++ {
		p.SetSize(320+i, 200+i)
		if err := p.Redraw(ctx); err != nil {
			t.Fatal(err)
		}
	}

	stats := p.Stats()
	if stats.Parses != 1 || stats.Cascades != 1 {
		t.Fatalf("parses = %d, cascades = %d, want 1 and 1", stats.Parses, stats.Cascades)
	}
	if stats.Redraws != 10 || stats.Repaints != 10 || stats.Layouts != 10 {
		t.Fatalf("redraws = %d, repaints = %d, layouts = %d, want 10 each",
			stats.Redraws, stats.Repaints, stats.Layouts)
	}
	if stats.LastRedraw <= 0 || stats.LastDraw <= 0 {
		t.Fatalf("last redraw = %v, last draw = %v", stats.LastRedraw, stats.LastDraw)
	}
}

type resizeTemplateData struct {
	Reads int
	Value string
}

func (d *resizeTemplateData) Text() string {
	d.Reads++
	return d.Value
}

func TestResizeKeepsCleanTemplateButRelayouts(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	p := newCachePage(t, `<style>#a{width:40px}@media(min-width:700px){#a{width:120px}}</style><p id="a">{{.Text}}</p>`)
	data := &resizeTemplateData{Value: "before"}
	p.SetData(data)
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	reads := data.Reads
	for _, width := range []int{800, 400, 800} {
		p.SetSize(width, 200)
		if err := p.Redraw(ctx); err != nil {
			t.Fatal(err)
		}
		if data.Reads != reads {
			t.Fatalf("resize executed template %d extra times", data.Reads-reads)
		}
		want := 40.0
		if width == 800 {
			want = 120
		}
		if got := viewportBox(t, p, "a").W; got != want {
			t.Fatalf("width %d: media-query box = %g, want %g", width, got, want)
		}
	}
	data.Value = "after"
	p.SetData(data)
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	if data.Reads != reads+1 || viewportBox(t, p, "a").Text != "after" {
		t.Fatal("SetData did not refresh template output after resizing")
	}
}

func TestResizeStillPublishesChangedScrollWindow(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name     string
		reported bool
	}{
		{name: "requests redraw", reported: true},
		{name: "declines redraw", reported: false},
	} {
		t.Run(test.name, func(t *testing.T) {
			p := newCachePage(t, `<p id="a">{{.Text}}</p>`)
			data := &resizeTemplateData{Value: "before"}
			p.SetData(data)
			p.SetWindowing(true)
			lastHeight := 0
			p.SetScrollWindow(func(_, height int) bool {
				if height == lastHeight {
					return false
				}
				lastHeight = height
				data.Value = "initial window"
				if height == 400 {
					data.Value = "resized window"
				}
				return test.reported
			})
			if err := p.Redraw(t.Context()); err != nil {
				t.Fatal(err)
			}
			reads := data.Reads
			p.SetSize(320, 400)
			if err := p.Redraw(t.Context()); err != nil {
				t.Fatal(err)
			}
			if data.Reads != reads+1 || viewportBox(t, p, "a").Text != "resized window" {
				t.Fatal("changed scroll window did not invalidate template output")
			}
		})
	}
}

func TestHoverRelayoutKeepsCache(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	p := newCachePage(t, `<style>#a:hover{color:#ff0000}</style><p id="a">hover</p>`)
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	x, y := boxCenter(t, p, "a")
	if err := p.Hover(ctx, x, y); err != nil {
		t.Fatal(err)
	}

	stats := p.Stats()
	if stats.Parses != 1 || stats.Cascades != 1 {
		t.Fatalf("parses = %d, cascades = %d, want 1 and 1", stats.Parses, stats.Cascades)
	}
	if stats.Layouts != 2 {
		t.Fatalf("layouts = %d, want 2", stats.Layouts)
	}
}

// fakeInspector is a screen double that satisfies host.Inspector.
type fakeInspector struct {
	on    bool
	stats host.Stats
}

func (f *fakeInspector) DevTools() bool      { return f.on }
func (f *fakeInspector) SetDevTools(on bool) { f.on = on }
func (f *fakeInspector) Stats() host.Stats   { return f.stats }

var _ host.Inspector = (*fakeInspector)(nil)

func TestDevToolsDefaultsOff(t *testing.T) {
	t.Parallel()

	p := newCachePage(t, `<p>static</p>`)
	if p.DevTools() {
		t.Fatal("devtools on by default")
	}

	p.SetDevTools(true)
	if !p.DevTools() {
		t.Fatal("SetDevTools(true) did not turn the overlay on")
	}

	p.SetDevTools(false)
	if p.DevTools() {
		t.Fatal("SetDevTools(false) did not turn the overlay off")
	}
}

func TestConfigDevToolsStartsOn(t *testing.T) {
	t.Parallel()

	p, err := page.New(page.Config{
		HTML:     `<p>static</p>`,
		Width:    320,
		Height:   200,
		DevTools: true,
	})
	if err != nil {
		t.Fatal(err)
	}

	if !p.DevTools() {
		t.Fatal("Config.DevTools did not start the overlay")
	}
}

func TestRoundedBorderReplays(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	screen, err := page.New(page.Config{
		HTML:   `<style>div { width: 60px; height: 40px; border: 2px solid #123456; border-radius: 6px; }</style><div></div>`,
		Width:  320,
		Height: 400,
	})
	if err != nil {
		t.Fatal(err)
	}

	if err := screen.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	if screen.Display() == nil {
		t.Fatal("rounded border did not replay")
	}
}

func TestEllipticalFillFallsBackToBitmap(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	screen, err := page.New(page.Config{
		HTML:   `<style>div { width: 60px; height: 40px; background: #123456; border-radius: 50% / 20%; }</style><div></div>`,
		Width:  320,
		Height: 400,
	})
	if err != nil {
		t.Fatal(err)
	}

	if err := screen.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	if screen.Display() != nil {
		t.Fatal("elliptical border was replayed")
	}

	if screen.Image() == nil {
		t.Fatal("no fallback bitmap")
	}
}

func TestFullFallbackReport(t *testing.T) {
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
	if err := screen.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	if screen.Display() != nil {
		t.Fatal("elliptical border was replayed")
	}
	if screen.FallbackDisplay() == nil {
		t.Fatal("no fallback display")
	}
	if screen.Image() == nil {
		t.Fatal("no fallback bitmap")
	}
	if !screen.LastPatch().Full {
		t.Fatalf("patch = %+v", screen.LastPatch())
	}
	rect, ok := screen.TakeDirty()
	if !ok || rect != image.Rect(0, 0, 320, 400) {
		t.Fatalf("rect %v ok %v", rect, ok)
	}
}

// TestNewLeavesTheMaxUnbounded checks a page with no max in Config asks the
// host for no cap, so the window keeps its OS maximize control.
func TestNewLeavesTheMaxUnbounded(t *testing.T) {
	t.Parallel()

	screen, err := page.New(page.Config{
		HTML:   "<p>Hi</p>",
		Width:  480,
		Height: 640,
	})
	if err != nil {
		t.Fatal(err)
	}

	maxW, maxH := screen.MaxSize()
	if maxW != 0 || maxH != 0 {
		t.Fatalf("max = %d x %d, want no cap", maxW, maxH)
	}
}

func TestTallPagePaintsPastTheWindow(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	screen, err := page.New(page.Config{
		HTML:   `<style>body { margin: 0; }</style><div style="height: 1200px">tall</div>`,
		Width:  320,
		Height: 400,
	})
	if err != nil {
		t.Fatal(err)
	}

	if err := screen.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	display := screen.Display()
	if display == nil {
		t.Fatal("display is nil")
	}

	if display.Height <= 400 {
		t.Fatalf("display height = %d, want the content past the window", display.Height)
	}
}

// samePicture reports whether two decoded pictures share bounds and every
// pixel, so encoder byte drift never fails a paint-equality check.
func samePicture(a, b image.Image) bool {
	if !a.Bounds().Eq(b.Bounds()) {
		return false
	}

	for y := a.Bounds().Min.Y; y < a.Bounds().Max.Y; y++ {
		for x := a.Bounds().Min.X; x < a.Bounds().Max.X; x++ {
			if a.At(x, y) != b.At(x, y) {
				return false
			}
		}
	}

	return true
}

func TestTextValidationLabelFocus(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	screen := formPage(t, `<label for="email" style="display:block;width:140px;height:24px">Email<span style="display:inline-block;width:16px;height:16px">*</span></label>`+
		`<input id="email" type="text"`+wide+`>`+
		`<input id="other" type="text"`+wide+`>`)
	clickID(t, screen, "other")
	if screen.FocusID() != "other" {
		t.Fatalf("setup focus %q", screen.FocusID())
	}

	clickBox(t, screen, func(box page.Box) bool {
		return box.Tag == "span" && box.Text == "*"
	})
	if screen.FocusID() != "email" {
		t.Fatalf("label click focus %q", screen.FocusID())
	}

	clickID(t, screen, "other")
	clickBox(t, screen, func(box page.Box) bool {
		return box.Tag == "label"
	})
	if screen.FocusID() != "email" {
		t.Fatalf("label text focus %q", screen.FocusID())
	}

	if err := screen.Type(ctx, "a"); err != nil {
		t.Fatal(err)
	}
	if screen.FormValue("email") != "a" {
		t.Fatalf("typed %q", screen.FormValue("email"))
	}
}

func TestTextValidationOTPSlots(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	screen := formPage(t, `<input id="code-0" type="text" inputmode="numeric" maxlength="1"`+wide+`>`+
		`<input id="code-1" type="text" inputmode="numeric" maxlength="1"`+wide+`>`+
		`<input id="code-2" type="text" inputmode="numeric" maxlength="1"`+wide+`>`)
	clickID(t, screen, "code-0")
	if err := screen.Type(ctx, "a"); err != nil {
		t.Fatal(err)
	}
	if screen.FormValue("code-0") != "" || screen.FocusID() != "code-0" {
		t.Fatalf("non-digit value %q focus %q", screen.FormValue("code-0"), screen.FocusID())
	}

	if err := screen.Type(ctx, "4"); err != nil {
		t.Fatal(err)
	}
	if screen.FormValue("code-0") != "4" || screen.FocusID() != "code-1" {
		t.Fatalf("digit value %q focus %q", screen.FormValue("code-0"), screen.FocusID())
	}

	if err := screen.Backspace(ctx); err != nil {
		t.Fatal(err)
	}
	if screen.FocusID() != "code-0" || screen.FormValue("code-0") != "4" || screen.FormValue("code-1") != "" {
		t.Fatalf("backspace focus %q code-0 %q code-1 %q", screen.FocusID(), screen.FormValue("code-0"), screen.FormValue("code-1"))
	}

	clickID(t, screen, "code-2")
	if err := screen.Type(ctx, "9"); err != nil {
		t.Fatal(err)
	}
	if screen.FormValue("code-2") != "9" || screen.FocusID() != "code-2" {
		t.Fatalf("last slot value %q focus %q", screen.FormValue("code-2"), screen.FocusID())
	}
}

func clickBox(t *testing.T, screen *page.Page, match func(page.Box) bool) {
	t.Helper()

	for _, box := range screen.Boxes() {
		if box.W <= 0 || box.H <= 0 || !match(box) {
			continue
		}

		if err := screen.Click(context.Background(), box.X+box.W/2, box.Y+box.H/2); err != nil {
			t.Fatal(err)
		}

		return
	}

	t.Fatal("no matching box")
}
