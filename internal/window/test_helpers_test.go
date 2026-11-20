package window

import (
	"context"
	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/host"
	"github.com/hajimehoshi/ebiten/v2"
	"image"
	"time"
)

func (f *fakeScreen) Image() image.Image       { return nil }
func (f *fakeScreen) Display() *layout.Display { return nil }
func (f *fakeScreen) PNG() []byte              { return nil }
func (f *fakeScreen) Generation() uint64 {
	if f.redraws < 0 {
		return 0
	}
	return uint64(f.redraws)
}
func (f *fakeScreen) Boxes() []layout.Box                           { return f.boxes }
func (f *fakeScreen) Click(context.Context, float64, float64) error { return nil }
func (f *fakeScreen) Hover(context.Context, float64, float64) error { return nil }
func (f *fakeScreen) Press(context.Context, float64, float64) error { return nil }
func (f *fakeScreen) Release(context.Context) error                 { return nil }
func (f *fakeScreen) Type(context.Context, string) error            { return nil }
func (f *fakeScreen) KeyDown(context.Context, string) error         { return nil }
func (f *fakeScreen) KeyUp(context.Context, string) error           { return nil }
func (f *fakeScreen) Backspace(context.Context) error               { return nil }
func (f *fakeScreen) DeleteWord(context.Context) error              { return nil }
func (f *fakeScreen) Submit(context.Context) error                  { return nil }
func (f *fakeScreen) Copy(context.Context) (string, bool, error)    { return "", false, nil }
func (f *fakeScreen) Cut(context.Context) (string, bool, error)     { return "", false, nil }
func (f *fakeScreen) Paste(context.Context, string) error           { return nil }
func (f *fakeScreen) SelectAll(context.Context) error               { return nil }
func (f *fakeScreen) Undo(context.Context) error                    { return nil }
func (f *fakeScreen) Redo(context.Context) error                    { return nil }

var (
	_ host.Pointer   = (*fakeScreen)(nil)
	_ host.Keyboard  = (*fakeScreen)(nil)
	_ host.Clipboard = (*fakeScreen)(nil)
	_ host.Screen    = (*fakeScreen)(nil)
)

type fakeScreen struct {
	width   int
	height  int
	minW    int
	minH    int
	maxW    int
	maxH    int
	redraws int
	boxes   []layout.Box
}

func (f *fakeScreen) Title() string       { return "fake" }
func (f *fakeScreen) Size() (int, int)    { return f.width, f.height }
func (f *fakeScreen) MinSize() (int, int) { return f.minW, f.minH }
func (f *fakeScreen) SetSize(width, height int) {
	f.width, f.height = f.Clamp(width, height)
}
func (f *fakeScreen) Clamp(width, height int) (int, int) {
	return clampRange(width, f.minW, f.maxW), clampRange(height, f.minH, f.maxH)
}
func (f *fakeScreen) Redraw(context.Context) error {
	f.redraws++

	return nil
}
func clampRange(value, low, high int) int {
	if value < low {
		return low
	}

	if high > 0 && value > high {
		return high
	}

	return value
}

// devScreen is a screen double with an inspector and a click count.
type devScreen struct {
	*fakeScreen
	on       bool
	clicked  int
	hovers   int
	draw     time.Duration
	boxReads int
}

func (d *devScreen) Boxes() []layout.Box {
	d.boxReads++
	return d.fakeScreen.Boxes()
}

func (d *devScreen) DevTools() bool              { return d.on }
func (d *devScreen) SetDevTools(on bool)         { d.on = on }
func (d *devScreen) Stats() host.Stats           { return host.Stats{LastDraw: d.draw} }
func (d *devScreen) SetDrawTime(v time.Duration) { d.draw = v }

func (d *devScreen) Click(context.Context, float64, float64) error {
	d.clicked++

	return nil
}

func (d *devScreen) Hover(context.Context, float64, float64) error {
	d.hovers++

	return nil
}

func newDevScreen() *devScreen {
	return &devScreen{
		fakeScreen: &fakeScreen{
			width:  800,
			height: 600,
			boxes:  []layout.Box{{ID: "known", Tag: "div", X: 10, Y: 10, W: 100, H: 40}},
		},
		on: true,
	}
}

func newDevShell(d *devScreen) *shell {
	return &shell{
		app:     d,
		ctx:     context.Background(),
		screenW: 1920,
		screenH: 1080,
		dev:     devState{on: d.on, available: d.on, watch: newKeyWatch()},
	}
}

// selectScreen records the range calls the window makes. focus models the
// state a click handler sets and SelectAt clears.
type selectScreen struct {
	*fakeScreen
	calls []string
	focus bool
}

func (f *selectScreen) Press(context.Context, float64, float64) error {
	f.calls = append(f.calls, "press")

	return nil
}

func (f *selectScreen) Click(context.Context, float64, float64) error {
	f.calls = append(f.calls, "click")
	f.focus = true

	return nil
}

func (f *selectScreen) Release(context.Context) error {
	f.calls = append(f.calls, "release")

	return nil
}

func (f *selectScreen) SelectAt(context.Context, float64, float64) error {
	f.calls = append(f.calls, "at")
	f.focus = false

	return nil
}

func (f *selectScreen) Drag(context.Context, float64, float64) error {
	f.calls = append(f.calls, "drag")

	return nil
}

func (f *selectScreen) SelectWordAt(context.Context, float64, float64) error {
	f.calls = append(f.calls, "word")

	return nil
}

func (f *selectScreen) SelectLineAt(context.Context, float64, float64) error {
	f.calls = append(f.calls, "line")

	return nil
}

// focusScreen is a fakeScreen with a focus order.
type focusScreen struct {
	*fakeScreen
	ids   []string
	at    int
	focus string
	nexts int
	prevs int
}

func (f *focusScreen) FocusNext(context.Context) error {
	f.nexts++
	if len(f.ids) == 0 {
		return nil
	}

	if f.focus == "" {
		f.at = 0
	} else {
		f.at = (f.at + 1) % len(f.ids)
	}

	f.focus = f.ids[f.at]

	return nil
}

func (f *focusScreen) FocusPrev(context.Context) error {
	f.prevs++
	if len(f.ids) == 0 {
		return nil
	}

	if f.focus == "" {
		f.at = len(f.ids) - 1
	} else {
		f.at = (f.at + len(f.ids) - 1) % len(f.ids)
	}

	f.focus = f.ids[f.at]

	return nil
}

func (f *focusScreen) Focus(_ context.Context, id string) error {
	f.focus = id

	return nil
}

func (f *focusScreen) FocusID() string { return f.focus }

var _ host.Focuser = (*focusScreen)(nil)

// menuScreen is a fakeScreen with context menu rows.
type menuScreen struct {
	*fakeScreen
	items  []host.MenuItem
	pasted string
}

func (f *menuScreen) ContextMenu() []host.MenuItem { return f.items }

func (f *menuScreen) Paste(_ context.Context, text string) error {
	f.pasted = text

	return nil
}

type pointerRouteScreen struct {
	fakeScreen
	presses, clicks, releases, drags int
}

func (f *pointerRouteScreen) Press(context.Context, float64, float64) error {
	f.presses++
	return nil
}
func (f *pointerRouteScreen) Click(context.Context, float64, float64) error {
	f.clicks++
	return nil
}
func (f *pointerRouteScreen) Release(context.Context) error {
	f.releases++
	return nil
}
func (f *pointerRouteScreen) SelectAt(context.Context, float64, float64) error { return nil }
func (f *pointerRouteScreen) Drag(context.Context, float64, float64) error {
	f.drags++
	return nil
}
func (f *pointerRouteScreen) SelectWordAt(context.Context, float64, float64) error { return nil }
func (f *pointerRouteScreen) SelectLineAt(context.Context, float64, float64) error { return nil }

// reloadScreen is a fake screen whose page reports one reload.
type reloadScreen struct {
	*fakeScreen
	polls   int
	left    int
	err     error
	display *layout.Display
	hovers  int
	hoverX  float64
	hoverY  float64
}

func (r *reloadScreen) Hover(_ context.Context, x, y float64) error {
	r.hovers++
	r.hoverX, r.hoverY = x, y

	return nil
}

func newReloadScreen() *reloadScreen {
	return &reloadScreen{
		fakeScreen: &fakeScreen{width: 80, height: 60, minW: 1, minH: 1, redraws: 1},
		display:    &layout.Display{Width: 80, Height: 60},
	}
}

func (r *reloadScreen) PollReload(context.Context) (bool, error) {
	r.polls++

	if r.err != nil {
		return false, r.err
	}

	if r.left <= 0 {
		return false, nil
	}

	r.left--
	r.redraws++
	r.boxes = nil
	r.display = &layout.Display{Width: 80, Height: 60}

	return true, nil
}

func (r *reloadScreen) Display() *layout.Display { return r.display }

// partialDisplay is a small replayable page with one box.
func partialDisplay(w, h int) *layout.Display {
	return &layout.Display{
		Width: w, Height: h, PointsPerPixel: 0.75, PixelPerPoint: 4.0 / 3,
		Ops: []layout.DisplayOp{
			{Kind: layout.DisplayOpFillRect, X: 0, Y: 0, W: float64(w), H: float64(h), R: 1, G: 1, B: 1, Alpha: 1},
			{Kind: layout.DisplayOpFillRect, X: 9, Y: 9, W: 15, H: 15, R: 1, Alpha: 1},
		},
		Order: []int{0, 1},
	}
}

func partialBuf(w, h int) *ebiten.Image {
	return ebiten.NewImage(w, h)
}

// dirtyScreen is a fakeScreen that keeps a display list and reports a
// scripted damage list, the way a page with TakeDirtyRects does.
type dirtyScreen struct {
	*fakeScreen
	display *layout.Display
	gen     uint64
	layouts int
	rects   []image.Rectangle
	ok      bool
	takes   int
	ticking bool
	frame   bool
}

func (d *dirtyScreen) FrameDirty() bool { return d.frame }

func (d *dirtyScreen) Display() *layout.Display { return d.display }
func (d *dirtyScreen) Generation() uint64       { return d.gen }
func (d *dirtyScreen) Ticking() bool            { return d.ticking }

func (d *dirtyScreen) TakeDirtyRects() ([]image.Rectangle, bool) {
	d.takes++

	return d.rects, d.ok
}

// click stands in for one counter click: the page lays out once and reports
// one changed box.
func (d *dirtyScreen) click(rect image.Rectangle) {
	d.gen++
	d.layouts++
	d.rects = []image.Rectangle{rect}
	d.ok = true
}

func newDirtyShell(screen *dirtyScreen) *shell {
	return &shell{
		app: screen, ctx: context.Background(),
		display: screen.display, seq: screen.gen,
		screenW: screen.width, screenH: screen.height,
	}
}

// oldScreen is a page from before the damage list: it reports one rect
// through TakeDirty only, so the window wraps it as a one-element list.
type oldScreen struct {
	*fakeScreen
	display *layout.Display
	gen     uint64
	rect    image.Rectangle
	ok      bool
	takes   int
}

func (o *oldScreen) Display() *layout.Display { return o.display }
func (o *oldScreen) Generation() uint64       { return o.gen }

func (o *oldScreen) TakeDirty() (image.Rectangle, bool) {
	o.takes++

	return o.rect, o.ok
}

const pinLayer = 2

type pinScreen struct {
	*dirtyScreen
}

func (p *pinScreen) ViewportPinZ() int { return pinLayer }

func pinOp(x, y, w, h, r, g, b float64, z int) layout.DisplayOp {
	op := layout.DisplayOp{
		Kind: layout.DisplayOpFillRect, X: x, Y: y, W: w, H: h,
		R: r, G: g, B: b, Alpha: 1,
	}
	if z > 0 {
		op.ZIndex, op.ZIndexSet = z, true
	}

	return op
}

func pinDisplay() *layout.Display {
	return &layout.Display{
		Width: 100, Height: 200, PointsPerPixel: 1,
		Ops: []layout.DisplayOp{
			pinOp(0, 0, 100, 400, 1, 1, 1, 0),
			pinOp(10, 80, 80, 40, 0, 1, 0, 0),
			pinOp(0, 0, 100, 40, 1, 0, 0, pinLayer),
			pinOp(0, 160, 100, 40, 0, 0, 1, pinLayer),
		},
		Order: []int{0, 1, 2, 3},
		Boxes: []layout.Box{{X: 0, Y: 0, W: 100, H: 400}},
	}
}

func newPinShell() (*pinScreen, *shell) {
	d := pinDisplay()
	screen := &pinScreen{dirtyScreen: &dirtyScreen{
		fakeScreen: &fakeScreen{width: 100, height: 200},
		display:    d, gen: 1,
	}}
	s := newDirtyShell(screen.dirtyScreen)
	s.app, s.display = screen, d
	s.screenW, s.screenH = 100, 200

	return screen, s
}
