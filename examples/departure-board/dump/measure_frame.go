package dump

import (
	"context"
	"image"
	"time"

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
// window's own order: read FrameDirty, take the damage list, then classify.
// Rect stays the bounding box of the list so the JSON keeps its old shape.
func tickFrame(ctx context.Context, app *board.App) (Frame, error) {
	p := app.Page()
	start := time.Now()
	err := app.Tick(ctx)
	f := Frame{TickNS: int64(time.Since(start))}
	if err != nil {
		return f, err
	}

	full := p.Ticking() && !p.FrameDirty()
	rects, ok := p.Testing().TakeDirtyRects()
	f.Rect = bounds(rects)
	f.Rects = len(rects)
	f.Area = areaOf(rects)

	content := image.Rectangle{}
	d := p.Display()
	if d != nil {
		content = contentRect(d, app)
		f.Ops = len(d.Ops)
		f.FullArea = content.Dx() * content.Dy()
	}

	switch {
	case full:
		f.Kind = kindFullReplay
		f.ReplayedOps = f.Ops
	case !ok:
		f.Kind = kindBlit
	case d == nil:
		f.Kind = kindRect
	case coversFull(rects, content):
		f.Kind = kindFullBuffer
		f.ReplayedOps = f.Ops
	default:
		f.Kind = kindRect
		f.ReplayedOps = meetsAny(d, rects)
	}

	return f, nil
}
