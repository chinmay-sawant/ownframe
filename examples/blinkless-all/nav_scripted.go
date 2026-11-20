package main

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/chinmay-sawant/ownframe"
)

// liveRig drives the navigation session with scripted input and a manual clock.
// The tests and the -demo mode both use it.
type liveRig struct {
	ctx    context.Context
	page   *ownframe.Page
	runner *liveRunner
	nav    *navSession
	clock  *manualLiveClock
	// The input the next step reports, in window coordinates.
	cx, cy float64
	click  bool
	down   bool
	keys   []string
	text   string
}

const rigStep = 100 * time.Millisecond

// newLiveRig opens the tab's page with its live sessions attached.
//
//nolint:mnd // pixel sizes, timings and markup of the live demo
func newLiveRig(ctx context.Context, app *App, tab string, width, height int) (*liveRig, error) {
	page, err := app.NewPage(ctx, tab, width, height)
	if err != nil {
		return nil, err
	}

	r := &liveRig{ctx: ctx, page: page, clock: &manualLiveClock{t: time.Unix(1_700_000_000, 0)}}
	r.nav = newNavSession()
	r.runner = &liveRunner{page: page, sessions: []liveSession{r.nav}, clock: r.clock, src: liveSource{
		cursor: func() (float64, float64) { return r.cx, r.cy },
		click:  func() bool { return r.click },
		down:   func() bool { return r.down },
		keys:   func() []string { return r.keys },
		text:   func() string { return r.text },
	}}

	return r, nil
}

// step moves the clock by d and runs one tick with the current input. The
// one-frame inputs (click, keys, text) are cleared after it.
func (r *liveRig) step(d time.Duration) error {
	r.clock.advance(d)

	err := r.runner.tick(r.ctx)

	r.click, r.keys, r.text = false, nil, ""

	return err
}

// idle runs ticks until d has passed.
func (r *liveRig) idle(d time.Duration) error {
	for left := d; left > 0; left -= rigStep {
		if err := r.step(min(left, rigStep)); err != nil {
			return err
		}
	}

	return nil
}

// model reads the page's current boxes.
func (r *liveRig) model() navModel { return navModelOf(r.page.Probe().Display()) }

// item returns the element with this id as laid out now.
func (r *liveRig) item(id string) (navItem, bool) {
	m := r.model()
	i := m.ByID(id)

	if i < 0 {
		return navItem{}, false
	}

	return m.Items[i], true
}

// pointAt scrolls the page so the element shows near the middle of the window
// and puts the cursor at fx, fy (0 to 1) of its box.
func (r *liveRig) pointAt(id string, fx, fy float64) error {
	it, ok := r.item(id)
	if !ok {
		return fmt.Errorf("no element %q", id)
	}

	_, h := r.page.Size()
	r.page.Scroll().SetScrollOffset(0, int(max(it.Y-float64(h)/3, 0)))
	_, sy := r.page.Scroll().ScrollOffset()
	r.cx, r.cy = it.X+it.W*fx, it.Y+it.H*fy-float64(sy)

	return nil
}

// press moves to the element, presses the left button and holds it.
func (r *liveRig) press(id string, fx, fy float64) error {
	if err := r.pointAt(id, fx, fy); err != nil {
		return err
	}

	r.click, r.down = true, true

	return r.step(rigStep)
}

// release lets go of the button.
func (r *liveRig) release() error {
	r.down = false

	return r.step(rigStep)
}

// tap is press then release.
//
//nolint:mnd // pixel sizes, timings and markup of the live demo
func (r *liveRig) tap(id string) error {
	if err := r.press(id, 0.5, 0.5); err != nil {
		return err
	}

	return r.release()
}

// drag moves the held pointer to fx, fy of the element.
func (r *liveRig) dragTo(id string, fx, fy float64) error {
	it, ok := r.item(id)
	if !ok {
		return fmt.Errorf("no element %q", id)
	}

	_, sy := r.page.Scroll().ScrollOffset()
	r.cx, r.cy = it.X+it.W*fx, it.Y+it.H*fy-float64(sy)

	return r.step(rigStep)
}

func (r *liveRig) key(k string) error {
	r.keys = []string{k}

	return r.step(rigStep)
}

func (r *liveRig) typ(text string) error {
	r.text = text

	return r.step(rigStep)
}

// box returns the laid out box of an element, failing when it is missing.
func (r *liveRig) box(id string) navItem {
	it, _ := r.item(id)

	return it
}

// text reads the text content of an element as laid out now.
func (r *liveRig) textOf(id string) string {
	for _, b := range r.page.Probe().Display().Boxes {
		if b.ID == id {
			return b.Text
		}
	}

	return ""
}

// writeFrame saves what the window would show: the viewport slice of the full
// page picture with a marker at the cursor.
//
//nolint:mnd // pixel sizes, timings and markup of the live demo
func (r *liveRig) writeFrame(dir string, n int) error {
	full := r.page.Probe().PNGWithContext(r.ctx)
	if len(full) == 0 {
		return fmt.Errorf("png: empty picture")
	}

	img, err := png.Decode(bytes.NewReader(full))
	if err != nil {
		return err
	}

	w, h := r.page.Size()
	_, sy := r.page.Scroll().ScrollOffset()
	view := image.NewRGBA(image.Rect(0, 0, w, h))

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			view.Set(x, y, img.At(img.Bounds().Min.X+x, img.Bounds().Min.Y+sy+y))
		}
	}

	for dy := -4; dy <= 4; dy++ {
		for dx := -4; dx <= 4; dx++ {
			c := color.RGBA{R: 220, G: 38, B: 38, A: 255}
			if dx == -4 || dx == 4 || dy == -4 || dy == 4 {
				c = color.RGBA{R: 255, G: 255, B: 255, A: 255}
			}

			if p := image.Pt(int(r.cx)+dx, int(r.cy)+dy); p.In(view.Bounds()) {
				view.Set(p.X, p.Y, c)
			}
		}
	}

	var buf bytes.Buffer
	if err = png.Encode(&buf, view); err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(dir, fmt.Sprintf("frame-%04d.png", n)), buf.Bytes(), demoFileMode)
}

// demoStep is one scripted action: it moves the rig and names itself.
type demoStep struct {
	Label string
	Do    func(r *liveRig) error
	// Report says what changed, read after the step.
	Report func(r *liveRig) string
}

// navDemoScript plays every live example once.
//
//nolint:mnd // pixel sizes, timings and markup of the live demo
func navDemoScript() []demoStep {
	chipOf := func(id string) func(r *liveRig) string {
		return func(r *liveRig) string { return r.textOf(id) }
	}
	key := func(k string) func(r *liveRig) error { return func(r *liveRig) error { return r.key(k) } }
	tap := func(id string) func(r *liveRig) error { return func(r *liveRig) error { return r.tap(id) } }

	return []demoStep{
		{"focus: click n-a", tap("n-a"), chipOf("chip-focus")},
		{"focus: ArrowDown (nav-down #n-f)", key("ArrowDown"), chipOf("chip-focus")},
		{"focus: ArrowLeft", key("ArrowLeft"), chipOf("chip-focus")},
		{"focus: ArrowLeft", key("ArrowLeft"), chipOf("chip-focus")},
		{"focus: on-screen Up", tap("arr-focus-up"), chipOf("chip-focus")},
		{"grid: click gn-a", tap("gn-a"), chipOf("chip-grid")},
		{"grid: normal, on-screen Down", tap("arr-grid-down"), chipOf("chip-grid")},
		{"grid: click gg-a", tap("gg-a"), chipOf("chip-grid")},
		{"grid: grid function, on-screen Down", tap("arr-grid-down"), chipOf("chip-grid")},
		{"contain: click c-b", tap("c-b"), chipOf("chip-contain")},
		{"contain: on-screen Down", tap("arr-contain-down"), chipOf("chip-contain")},
		{"action: on-screen Down", tap("arr-action-down"), chipOf("chip-action")},
		{"action: ArrowDown", key("ArrowDown"), chipOf("chip-action")},
		{"action: on-screen Right", tap("arr-action-right"), chipOf("chip-action")},
		{"action: on-screen Up (back)", tap("arr-action-up"), chipOf("chip-action")},
		{"action: on-screen Left (back)", tap("arr-action-left"), chipOf("chip-action")},
		{"caret: click cell, type abc", func(r *liveRig) error {
			if err := r.tap("cr-bar"); err != nil {
				return err
			}

			return r.typ("abc")
		}, chipOf("chip-caret")},
		{"caret: 500ms later", func(r *liveRig) error { return r.idle(500 * time.Millisecond) }, chipOf("chip-caret")},
		{"caret: 1000ms later", func(r *liveRig) error { return r.idle(500 * time.Millisecond) }, chipOf("chip-caret")},
		{"field: click content field, type", func(r *liveRig) error {
			if err := r.tap("fs-content"); err != nil {
				return err
			}

			return r.typ(" grows and grows")
		}, chipOf("chip-field")},
		{"field: click fixed field, type", func(r *liveRig) error {
			if err := r.tap("fs-fixed"); err != nil {
				return err
			}

			return r.typ(" and stays")
		}, chipOf("chip-field")},
		{"security: type into auto and none", func(r *liveRig) error {
			if err := r.tap("pw-auto"); err != nil {
				return err
			}

			if err := r.typ("x"); err != nil {
				return err
			}

			if err := r.tap("pw-none"); err != nil {
				return err
			}

			return r.typ("y")
		}, chipOf("chip-security")},
		{"select: click word in us-text", func(r *liveRig) error { return r.press("us-text", 0.28, 0.5) }, chipOf("chip-select")},
		{"select: click us-all", func(r *liveRig) error { return r.press("us-all", 0.28, 0.5) }, chipOf("chip-select")},
		{"select: click us-none", func(r *liveRig) error { return r.press("us-none", 0.28, 0.5) }, chipOf("chip-select")},
		{"adjust: toggle inflation", tap("ta-btn"), chipOf("chip-adjust")},
		{"slider: drag sl-h", func(r *liveRig) error {
			if err := r.press("sl-h", 0.1, 0.5); err != nil {
				return err
			}

			if err := r.dragTo("sl-h", 0.7, 0.5); err != nil {
				return err
			}

			return r.release()
		}, chipOf("chip-slider")},
		{"slider: drag sl-rtl", func(r *liveRig) error {
			if err := r.press("sl-rtl", 0.9, 0.5); err != nil {
				return err
			}

			if err := r.dragTo("sl-rtl", 0.25, 0.5); err != nil {
				return err
			}

			return r.release()
		}, chipOf("chip-slider")},
		{"slider: drag sl-v", func(r *liveRig) error {
			if err := r.press("sl-v", 0.5, 0.9); err != nil {
				return err
			}

			if err := r.dragTo("sl-v", 0.5, 0.2); err != nil {
				return err
			}

			return r.release()
		}, chipOf("chip-slider")},
		{"drag: press the title and move", func(r *liveRig) error {
			if err := r.press("drag-title", 0.5, 0.5); err != nil {
				return err
			}

			if err := r.dragTo("drag-area", 0.7, 0.8); err != nil {
				return err
			}

			return r.release()
		}, chipOf("chip-drag")},
		{"drag: press the close box", func(r *liveRig) error {
			if err := r.press("drag-btn", 0.5, 0.5); err != nil {
				return err
			}

			if err := r.dragTo("drag-area", 0.1, 0.1); err != nil {
				return err
			}

			return r.release()
		}, chipOf("chip-drag")},
		{"layer: click wc", func(r *liveRig) error {
			if err := r.tap("wc"); err != nil {
				return err
			}

			return r.idle(400 * time.Millisecond)
		}, chipOf("wc-badge")},
		{"layer: click wc2", func(r *liveRig) error {
			if err := r.tap("wc2"); err != nil {
				return err
			}

			return r.idle(400 * time.Millisecond)
		}, chipOf("wc-badge")},
		{"appearance: hold ap-btn", func(r *liveRig) error { return r.press("ap-btn", 0.5, 0.5) }, chipOf("chip-appear")},
		{"appearance: hold ap-none", func(r *liveRig) error {
			if err := r.release(); err != nil {
				return err
			}

			return r.press("ap-none", 0.5, 0.5)
		}, chipOf("chip-appear")},
		{"appearance: release", func(r *liveRig) error { return r.release() }, chipOf("chip-appear")},
		{"image: frames at 0", func(r *liveRig) error { return r.pointAt("ia-normal", 0.5, 0.5) }, chipOf("chip-image")},
		{"image: 200ms", func(r *liveRig) error { return r.idle(200 * time.Millisecond) }, chipOf("chip-image")},
		{"image: 400ms", func(r *liveRig) error { return r.idle(200 * time.Millisecond) }, chipOf("chip-image")},
		{"image: pause normal", tap("ia-btn"), chipOf("chip-image")},
		{"image: 600ms while paused", func(r *liveRig) error { return r.idle(600 * time.Millisecond) }, chipOf("chip-image")},
		{"image: resume", tap("ia-btn"), chipOf("chip-image")},
		{"image: 300ms after resume", func(r *liveRig) error { return r.idle(300 * time.Millisecond) }, chipOf("chip-image")},
		{"focus: back to the first example", func(r *liveRig) error { return r.pointAt("n-a", 0.5, 0.5) }, chipOf("chip-focus")},
	}
}

// runNavDemo plays the scripted nav session with no window and writes one PNG
// per step into dir.
func runNavDemo(ctx context.Context, app *App, width, height int, dir string, out io.Writer) (int, error) {
	r, err := newLiveRig(ctx, app, "nav", width, height)
	if err != nil {
		return 0, err
	}

	if err = r.idle(rigStep); err != nil {
		return 0, err
	}

	steps := navDemoScript()
	for n, s := range steps {
		if err = s.Do(r); err != nil {
			return n, fmt.Errorf("step %d (%s): %w", n, s.Label, err)
		}

		if _, err = fmt.Fprintf(out, "step %03d %-40s | %s\n", n, s.Label, s.Report(r)); err != nil {
			return n, err
		}

		if err = r.writeFrame(dir, n); err != nil {
			return n, err
		}
	}

	return len(steps), nil
}

// demoDirMode and demoFileMode are the permissions of the -demo output.
const (
	demoDirMode  = 0o750
	demoFileMode = 0o600
)
