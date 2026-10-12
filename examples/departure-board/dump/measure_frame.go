package dump

import (
	"context"
	"image"
	"time"

	"github.com/chinmay-sawant/blinkless/layout"

	"github.com/chinmay-sawant/ownframe/examples/departure-board/board"
)

// Frame kinds, named as the window picks them.
const (
	kindFullReplay = "full-replay"
	kindFullBuffer = "full-buffer"
	kindRect       = "rect"
	kindBlit       = "blit"
)

// tickFrame ticks app once and emulates the window's replay choice in the
// window's own order: read FrameDirty, take the dirty rect, then classify.
func tickFrame(ctx context.Context, app *board.App) (Frame, error) {
	p := app.Page()
	start := time.Now()
	err := app.Tick(ctx)
	f := Frame{TickNS: int64(time.Since(start))}
	if err != nil {
		return f, err
	}

	full := p.Ticking() && !p.FrameDirty()
	rect, ok := p.TakeDirty()
	f.Rect = Rect{X0: rect.Min.X, Y0: rect.Min.Y, X1: rect.Max.X, Y1: rect.Max.Y}

	d := p.Display()
	if d != nil {
		f.Ops = len(d.Ops)
		f.FullArea = d.Width * d.Height
	}

	switch {
	case full:
		f.Kind = kindFullReplay
		f.ReplayedOps = f.Ops
	case !ok:
		f.Kind = kindBlit
	case d == nil:
		f.Kind = kindRect
	case rect == image.Rect(0, 0, d.Width, d.Height):
		f.Kind = kindFullBuffer
		f.ReplayedOps = f.Ops
	default:
		f.Kind = kindRect
		f.ReplayedOps = meets(d, rect)
	}

	return f, nil
}

// meets counts the ops whose raw X, Y, W, H box meets rect. It approximates
// the real replay filter (internal/replay opTouches), which converts
// point-space bounds to CSS pixels and pads them for strokes and
// antialiasing; this test compares the raw fields, so ReplayedOps is an
// estimate, wanted only for the dump.
func meets(d *layout.Display, rect image.Rectangle) int {
	n := 0
	for i := range d.Ops {
		op := &d.Ops[i]
		box := image.Rect(int(op.X), int(op.Y), int(op.X+op.W), int(op.Y+op.H))
		if !box.Intersect(rect).Empty() {
			n++
		}
	}
	return n
}
