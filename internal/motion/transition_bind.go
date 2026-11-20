package motion

import (
	"errors"
	"fmt"
	"github.com/chinmay-sawant/blinkless/layout"
	"time"
)

// BindTransition builds a transition renderer for one CSS property on a box. The
// destination comes from the display, so a caller that rebinds later wants
// BindTransitionTo with the paints it captured on the first bind. BindTransition trusts the
// captured indices, because the paint it writes as the destination is the
// one the display holds. BindTransitionTo checks them, because it writes a paint the
// caller captured on an earlier display.
func BindTransition(d *layout.Display, box layout.Box, property string,
	indices []int, initial []Paint, start time.Time) (*Renderer, error) {
	return BindTransitionTo(d, box, property, indices, initial, nil, start)
}

// adjust returns the captured paint with every field the operation changed
// since the capture set to the operation's current value. The mask and the
// matrix are captured values, not ranges: a transition keeps the fields its
// property owns, and the matrix pair is what Blend decomposes.
func adjust(captured, initial, final, current Paint) Paint {
	return Paint{
		X:         field(captured.X, initial.X, final.X, current.X),
		Y:         field(captured.Y, initial.Y, final.Y, current.Y),
		W:         field(captured.W, initial.W, final.W, current.W),
		H:         field(captured.H, initial.H, final.H, current.H),
		R:         field(captured.R, initial.R, final.R, current.R),
		G:         field(captured.G, initial.G, final.G, current.G),
		B:         field(captured.B, initial.B, final.B, current.B),
		Alpha:     field(captured.Alpha, initial.Alpha, final.Alpha, current.Alpha),
		Opacity:   field(captured.Opacity, initial.Opacity, final.Opacity, current.Opacity),
		Matrix:    captured.Matrix,
		HasMatrix: captured.HasMatrix,
		Mask:      captured.Mask,
	}
}

// field returns the captured value when the operation's current value still
// lies inside the range the capture spans, and the current value otherwise.
func field(captured, initial, final, current float64) float64 {
	lo, hi := initial, final
	if lo > hi {
		lo, hi = hi, lo
	}

	if current < lo-paintSlack || current > hi+paintSlack {
		return current
	}

	return captured
}

// paintSlack is the room a captured paint keeps when a display hands the
// operation back. A layout that rounds a value differently still describes
// the same operation.
const paintSlack = 0.5

// endpoints returns the start and destination paints of one operation. A
// field the operation still carries inside the captured range keeps the
// captured value, so the transition animates it. A field the layout changed
// keeps the operation's current value, so a rebind does not write a stale
// size or position back.
func endpoints(op *layout.DisplayOp, initial, final Paint) (Paint, Paint) {
	current := Read(op)

	return adjust(initial, initial, final, current), adjust(final, initial, final, current)
}

// BindTransitionTo builds a transition renderer with an explicit destination paint.
// A nil final reads the destination from the display, which is what a first
// bind wants. A caller that rebinds later passes the paints it captured on
// the first bind, so a redraw in the middle of the transition does not make
// the destination the value the transition had reached.
//
// The captured paints describe the operations they were read from. A rebind
// checks that the operations are still the ones the box owns, and a field
// the layout changed keeps the operation's current value. See bind_verify.go
// and bind_paint.go.
func BindTransitionTo(d *layout.Display, box layout.Box, property string,
	indices []int, initial, final []Paint, start time.Time) (*Renderer, error) {
	if d == nil {
		return nil, errors.New("csstransition: display is nil")
	}
	record, ok := transitionRecordOf(box, property)
	if !ok {
		return nil, errors.New("csstransition: box has no matching transition metadata")
	}
	ops := Resolve(d, indices)
	if len(ops) != len(indices) {
		return nil, errors.New("csstransition: captured operations are missing from display")
	}
	if len(ops) == 0 || len(initial) != len(ops) {
		return nil, errors.New("csstransition: target operation paints do not match")
	}
	if final != nil && len(final) != len(ops) {
		return nil, errors.New("csstransition: destination paints do not match")
	}
	if final != nil {
		if err := ownedByBox(d, box, ops); err != nil {
			return nil, err
		}
	}
	renderer, err := newTransitionRenderer(buildTracks(ops, initial, final, record, property))
	if err != nil {
		return nil, err
	}
	renderer.Start(start)
	return renderer, nil
}

// buildTracks turns the resolved operations and the captured paints into the
// tracks the renderer samples. The captured destination is the paint the
// first bind read, and the endpoints keep the operation's current value for
// every field the layout changed since the capture.
func buildTracks(ops []*layout.DisplayOp, initial, final []Paint,
	record transitionRecord, property string) []Track {
	tracks := make([]Track, len(ops))
	discrete := discretePaintProperty(property)
	for i := range tracks {
		to := Read(ops[i])
		if final != nil {
			to = final[i]
		}
		from, dest := endpoints(ops[i], initial[i], to)
		tracks[i] = Track{Op: ops[i], From: from, To: dest,
			Duration: millis(record.DurationMS), Delay: millis(record.DelayMS),
			Easing:   record.TimingFunction,
			Discrete: discrete && record.Behavior == "allow-discrete"}
		if discrete && record.Behavior != "allow-discrete" {
			tracks[i].Duration, tracks[i].Delay = 0, 0
		}
	}

	return tracks
}

func millis(value float64) time.Duration { return time.Duration(value * float64(time.Millisecond)) }

// transitionRecord is the timing a box published for a property.
type transitionRecord struct {
	DurationMS, DelayMS float64
	TimingFunction      string
	Behavior            string
}

// transitionRecordOf returns the box's transition record for the property.
// A record for "all" covers every property.
func transitionRecordOf(box layout.Box, property string) (transitionRecord, bool) {
	for i := range box.Transitions {
		item := box.Transitions[i]
		if item.Property == property || item.Property == "all" {
			return transitionRecord{DurationMS: item.DurationMS, DelayMS: item.DelayMS,
				TimingFunction: item.TimingFunction, Behavior: item.Behavior}, true
		}
	}

	return transitionRecord{}, false
}

// ownedByBox refuses when an operation left the box. The captured paints
// name the operations of the box, so an operation that left it means the
// indices resolve to other operations, and the rebind refuses instead of
// writing a captured paint onto them.
func ownedByBox(d *layout.Display, box layout.Box, ops []*layout.DisplayOp) error {
	for _, op := range ops {
		if !boxOwnsOp(d, box, op) {
			return errors.New("csstransition: captured operations no longer belong to the box")
		}
	}

	return nil
}

// boxOwnsOp reports whether the box owns the operation. A box with
// ownership IDs names them; an older box falls back to the geometric test
// the capture used.
func boxOwnsOp(d *layout.Display, box layout.Box, op *layout.DisplayOp) bool {
	if op == nil {
		return false
	}

	if box.OperationIDs == nil {
		return Contains(d, box, *op)
	}

	for _, id := range box.OperationIDs {
		if id == op.ID {
			return true
		}
	}

	return false
}

// Track describes one display operation's transition.
type Track struct {
	Op              *layout.DisplayOp
	From, To        Paint
	Duration, Delay time.Duration
	Easing          string
	// Discrete switches endpoints at 50% of eased progress.
	Discrete bool
}

// Renderer samples tracks from their start time. Track operations must remain
// valid until Step reports that every track has finished.
type Renderer struct {
	tracks  []Track
	start   time.Time
	started bool
}

// newTransitionRenderer validates and copies the tracks. Negative delays are allowed.
func newTransitionRenderer(tracks []Track) (*Renderer, error) {
	if len(tracks) == 0 {
		return nil, errors.New("transition: at least one track is required")
	}
	copyOfTracks := append([]Track(nil), tracks...)
	for i, track := range copyOfTracks {
		if track.Op == nil {
			return nil, fmt.Errorf("transition: track %d has no display operation", i)
		}
		if track.Duration < 0 {
			return nil, fmt.Errorf("transition: track %d has negative duration", i)
		}
	}
	return &Renderer{tracks: copyOfTracks}, nil
}

// Start sets the shared start time. Calling it again restarts every track.
func (r *Renderer) Start(now time.Time) {
	if r == nil {
		return
	}
	r.start, r.started = now, true
}

// Retarget points every track at the operation at the same position in ops.
// A display commit rebuilds the operation list, so a renderer that keeps its
// clock and endpoints must write through the new list. It refuses a different
// track count, because a renderer built for one target set must not half-write
// another.
func (r *Renderer) Retarget(ops []*layout.DisplayOp) error {
	if r == nil {
		return errors.New("transition: renderer is nil")
	}
	if len(ops) != len(r.tracks) {
		return fmt.Errorf("transition: retarget has %d operations, want %d",
			len(ops), len(r.tracks))
	}
	for i := range ops {
		if ops[i] == nil {
			return fmt.Errorf("transition: retarget operation %d is nil", i)
		}
		r.tracks[i].Op = ops[i]
	}
	return nil
}

// easeMidpoint is the eased progress where a discrete track flips.
const easeMidpoint = 0.5

// ProgressFor returns the eased progress of one transition record at now and
// whether the run is still active. It is the clock Step and Progress share,
// for callers that carry a record rather than a renderer: a geometry or
// discrete transition has no display operation to bind, but it follows the
// same duration, delay, and timing function.
//
// A run inside a positive delay reports zero progress and stays active. A
// finished or zero-duration run reports one and inactive, so the caller can
// install the destination once and drop the entry.
func ProgressFor(record layout.Transition, start, now time.Time) (float64, bool) {
	if record.DurationMS <= 0 {
		return 1, false
	}

	elapsed := float64(now.Sub(start)) - record.DelayMS*float64(time.Millisecond)
	duration := record.DurationMS * float64(time.Millisecond)
	if elapsed < 0 {
		return 0, true
	}
	if elapsed >= duration {
		return 1, false
	}

	return Ease(record.TimingFunction, elapsed/duration), true
}

// Progress returns the timing-function output of the run at now, on the same
// clock Step samples. A run inside a positive delay reports zero, and a
// finished or zero-duration run reports one. A normal bind shares one record
// across every track, so the first track with a duration speaks for the run.
//
// The value is the eased progress, not the raw time fraction, because CSS
// Transitions defines the reversing shortening factor as the portion of the
// value range a run has crossed. A caller that reverses a run uses this value
// to shorten the reverse by that portion.
func (r *Renderer) Progress(now time.Time) float64 {
	if r == nil || !r.started {
		return 0
	}
	for _, track := range r.tracks {
		if track.Duration <= 0 {
			continue
		}
		progress := float64(now.Sub(r.start)-track.Delay) / float64(track.Duration)
		if progress <= 0 {
			return 0
		}
		if progress >= 1 {
			return 1
		}
		return Ease(track.Easing, progress)
	}

	return 1
}

// Step writes the current sample and reports whether any track remains active.
func (r *Renderer) Step(now time.Time) bool {
	if r == nil || !r.started {
		return false
	}
	active := false
	for _, track := range r.tracks {
		if track.Duration == 0 {
			Write(track.Op, track.To)
			continue
		}
		progress := float64(now.Sub(r.start)-track.Delay) / float64(track.Duration)
		if progress < 0 {
			Write(track.Op, track.From)
			active = true
			continue
		}
		if progress >= 1 {
			Write(track.Op, track.To)
			continue
		}
		if track.Discrete {
			if Ease(track.Easing, progress) < easeMidpoint {
				Write(track.Op, track.From)
			} else {
				Write(track.Op, track.To)
			}
			active = true
			continue
		}
		value := Blend(track.From, track.To, Ease(track.Easing, progress))
		Write(track.Op, value)
		active = true
	}
	return active
}
