// Package renderbackend defines the experimental boundary between OwnFrame's
// display-list replay and a graphics implementation.
//
// This package deliberately has no Ebiten dependency. The existing Ebiten
// replay remains the production implementation while alternative backends are
// evaluated. The interface is intentionally minimal: concrete display-list
// operations and GPU resources will be adapted after profiling their costs.
package renderbackend

import (
    "errors"
    "time"

    "github.com/chinmay-sawant/blinkless/layout"
)

// ErrClosed indicates that a backend has been shut down.
var ErrClosed = errors.New("render backend closed")

// Viewport describes logical dimensions and device scale.
type Viewport struct {
    Width, Height int
    DeviceScale float64
}

// Frame is an opaque frame destination owned by the backend.
type Frame interface {
    // Present submits the completed frame to the window surface.
    Present() error
}

// Backend draws Blinkless display lists. Implementations must preserve
// display-list paint order, fixed-position semantics, and layer composition.
// Backend implementations own their GPU resources and must be closed.
type Backend interface {
    BeginFrame(Viewport) (Frame, error)
    Draw(Frame, *layout.Display, float64, float64) error
    Close() error
}

// EventHost manages window events independently of graphics operations.
// WaitUntil should block until input, a redraw request, or the deadline.
// A zero deadline means no animation is currently scheduled.
type EventHost interface {
    WaitUntil(time.Time) (dirty bool, closed bool, err error)
    RequestRedraw()
}

// Run executes an event-driven frame loop. The callback is responsible for
// producing/presenting a frame; scheduling the next animation deadline is
// intentionally owned by the caller.
func Run(host EventHost, nextDeadline func() time.Time, draw func() error) error {
    dirty := true // draw initial frame
    for {
        if dirty {
            if err := draw(); err != nil {
                return err
            }
            dirty = false
        }
        changed, closed, err := host.WaitUntil(nextDeadline())
        if err != nil {
            return err
        }
        if closed {
            return nil
        }
        dirty = changed
    }
}
