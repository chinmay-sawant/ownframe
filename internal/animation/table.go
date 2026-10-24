package animation

import (
	"strconv"
	"strings"
	"time"

	"github.com/chinmay-sawant/ownframe/internal/cssmotion"
)

// fieldCount is the number of paint fields sampled per keyframe. It stays in
// step with the fieldX..fieldB indices in fields.go.
const fieldCount = fieldB + 1

// extraStops is the room reserved for the 0 and 1 sentinels.
const extraStops = 2

// easingFn maps progress inside one keyframe interval to eased progress.
type easingFn func(float64) float64

// stop is one precomputed keyframe value at a normalized offset. fromBase is
// the coefficient of the target's own paint for a sentinel: a missing keyframe
// after a replacing effect holds the underlying paint, so the value at that
// stop is base-dependent. A real keyframe has fromBase 0. ease is the timing
// function declared on the keyframe at this stop, which eases the interval that
// starts here; a synthetic sentinel carries the effect's own function.
type stop struct {
	at, value float64
	fromBase  float64
	ease      easingFn
}

// segment is the interval between two adjacent stops. span is next.at-prev.at,
// and ease is the easing parsed once at bind for that interval.
type segment struct {
	span float64
	ease easingFn
}

// fieldTable is the precomputed stop table for one effect and one paint field.
// stops includes the explicit 0 and 1 sentinels, so it is sorted and never
// shorter than two entries once present is true. cursor and lastAt let a
// monotonic clock scan forward instead of searching from zero every frame.
type fieldTable struct {
	stops    []stop
	segments []segment
	cursor   int
	lastAt   float64
	sampled  bool
	present  bool
}

// effectTable is the bound sampling state for one effect. fields and mask are
// built once and read every frame. transform marks an effect whose keyframes
// carry a parsed transform list; those effects stay on compose until phase 02
// gives the transform its own stop table.
type effectTable struct {
	fields    [fieldCount]fieldTable
	mask      uint32
	mode      Composition
	transform bool
	wasPaused bool
	lastNow   time.Time
}

// valueSet is the result of sampling one effect at one time. value is the
// interpolated constant and fromBase is the coefficient of the target's own
// paint, so the set can be applied to every target without resampling.
type valueSet struct {
	value    [fieldCount]float64
	fromBase [fieldCount]float64
	mask     uint32
}

// valueForField returns the keyframe accessor for a paint field.
func valueForField(field int) valueOf {
	switch field {
	case fieldX:
		return moveX
	case fieldY:
		return moveY
	case fieldOpacity:
		return opacity
	case fieldR:
		return red
	case fieldG:
		return green
	case fieldB:
		return blue
	}
	return nil
}

// buildEffectTable precomputes every field table for one effect.
func buildEffectTable(e Effect) effectTable {
	table := effectTable{mode: e.Timing.Composition}
	if hasTransform(e.Frames) {
		// compose samples the decomposed transform per target. Keep that path
		// until the transform has a precomputed table of its own.
		table.transform = true
		return table
	}
	ease := effectEase(e.Timing)
	for field := 0; field < fieldCount; field++ {
		ft := buildFieldTable(e.Frames, valueForField(field), ease,
			sentinelBase(e.Timing.Composition, field))
		table.fields[field] = ft
		if ft.present {
			table.mask |= 1 << field
		}
	}
	return table
}

// buildFieldTable flattens one field's keyframes into a sorted stop list with
// explicit 0 and 1 sentinels when the keyframes omit those offsets. Each real
// stop carries the timing function declared on its keyframe, so the interval
// that starts at that keyframe eases with it.
func buildFieldTable(frames []Keyframe, get valueOf, ease easingFn, fallback float64) fieldTable {
	present := 0
	for i := range frames {
		if _, ok := get(frames[i]); ok {
			present++
		}
	}
	if present == 0 {
		return fieldTable{}
	}
	stops := make([]stop, 0, present+extraStops)
	first, last := 0.0, 0.0
	seen := 0
	for i := range frames {
		v, ok := get(frames[i])
		if !ok {
			continue
		}
		if seen == 0 {
			first = frames[i].Offset
			if first > 0 {
				stops = append(stops, stop{at: 0, fromBase: fallback, ease: ease})
			}
		}
		stops = append(stops, stop{at: frames[i].Offset, value: v,
			ease: frames[i].frameEasing(ease)})
		last = frames[i].Offset
		seen++
	}
	if last < 1 {
		stops = append(stops, stop{at: 1, fromBase: fallback, ease: ease})
	}
	segments := make([]segment, len(stops)-1)
	for i := range segments {
		segments[i] = segment{span: stops[i+1].at - stops[i].at, ease: stops[i].ease}
	}
	return fieldTable{stops: stops, segments: segments, present: true}
}

// sentinelBase reports whether the sentinel value for a field is the target's
// own paint. A replacing effect holds the underlying paint for an omitted
// keyframe; an additive effect adds zero.
func sentinelBase(mode Composition, field int) float64 {
	if mode != Replace {
		return 0
	}
	switch field {
	case fieldOpacity, fieldR, fieldG, fieldB:
		return 1
	}
	return 0
}

// effectEase turns the effect timing function into one function reused by
// every segment. Timing.Easing wins over the CSS string, as it does today.
// cssmotion exposes ease as a string sampler, so a named timing function keeps
// one string and avoids the per-sample branch; parametric steps(), linear(),
// and cubic-bezier() still parse inside cssmotion.Ease until it exports a
// parsed form.
func effectEase(t Timing) easingFn {
	if t.Easing != nil {
		return easingFn(t.Easing)
	}
	return cssEase(t.TimingFunction)
}

// frameEasing returns the easing for the interval that starts at this keyframe.
// A keyframe that declares animation-timing-function eases that interval; one
// that does not uses the effect's function. The binder fills ease once, so the
// sample path never parses the string.
func (k Keyframe) frameEasing(fallback easingFn) easingFn {
	if k.ease != nil {
		return k.ease
	}
	if k.TimingFunction != "" {
		return cssEase(k.TimingFunction)
	}
	return fallback
}

// cssEase parses one CSS timing-function string into a sampler. An invalid
// steps(1, jump-none) is dropped by the cascade, so a direct Timing that still
// carries it falls back to the initial ease instead of a linear ramp.
func cssEase(value string) easingFn {
	if value == "" || strings.EqualFold(strings.TrimSpace(value), "linear") {
		return linearEase
	}
	if invalidStepsJumpNone(value) {
		value = "ease"
	}
	return func(part float64) float64 { return cssmotion.Ease(value, part) }
}

// stepsJumpNoneMin is the smallest step count jump-none accepts. CSS Easing
// rejects steps(1, jump-none), so that function has no sample to compute.
const stepsJumpNoneMin = 2

// invalidStepsJumpNone reports steps(n, jump-none) with fewer than two steps.
func invalidStepsJumpNone(value string) bool {
	const prefix, suffix = "steps(", ")"
	lower := strings.ToLower(strings.TrimSpace(value))
	if !strings.HasPrefix(lower, prefix) || !strings.HasSuffix(lower, suffix) {
		return false
	}
	parts := strings.Split(strings.TrimSuffix(lower[len(prefix):], suffix), ",")
	if len(parts) != stepsJumpNoneMin || strings.TrimSpace(parts[1]) != "jump-none" {
		return false
	}
	count, err := strconv.Atoi(strings.TrimSpace(parts[0]))
	return err == nil && count < stepsJumpNoneMin
}

func linearEase(part float64) float64 { return part }

// sample writes the values for every present field at at into set.
func (t *effectTable) sample(at float64, set *valueSet) {
	set.mask = t.mask
	for field := 0; field < fieldCount; field++ {
		if t.mask&(1<<field) == 0 {
			continue
		}
		set.value[field], set.fromBase[field] = t.fields[field].sample(at)
	}
}

// sample resolves one field at at. A monotonic at advances the stored cursor;
// a repeated or backwards at falls back to the first segment when it moves
// backwards, so a reverse or alternate playback stays correct.
func (t *fieldTable) sample(at float64) (float64, float64) {
	if len(t.stops) == 0 {
		return 0, 0
	}
	if at <= t.stops[0].at {
		t.cursor, t.lastAt, t.sampled = 0, at, true
		return t.stops[0].value, t.stops[0].fromBase
	}
	last := len(t.stops) - 1
	if at > t.stops[last].at {
		t.cursor, t.lastAt, t.sampled = last, at, true
		return t.stops[last].value, t.stops[last].fromBase
	}
	i := t.cursor
	if !t.sampled || at < t.lastAt || i < 1 || i >= len(t.stops) {
		i = 1
	}
	for i < len(t.stops)-1 && at > t.stops[i].at {
		i++
	}
	t.cursor, t.lastAt, t.sampled = i, at, true
	return interpolate(t.stops[i-1], t.stops[i], at, &t.segments[i-1])
}

// interpolate blends two stops and their base coefficients.
func interpolate(prev, next stop, at float64, seg *segment) (float64, float64) {
	if seg.span <= 0 {
		return next.value, next.fromBase
	}
	part := seg.ease((at - prev.at) / seg.span)
	value := prev.value + (next.value-prev.value)*part
	fromBase := prev.fromBase + (next.fromBase-prev.fromBase)*part
	return value, fromBase
}

// applyTo composes the sampled set onto current for one target. It uses
// compose.go's apply so the field semantics stay in one place.
func (s *valueSet) applyTo(current *cssmotion.Paint, base cssmotion.Paint, mode Composition) {
	for field := 0; field < fieldCount; field++ {
		if s.mask&(1<<field) == 0 {
			continue
		}
		value := s.value[field]
		if s.fromBase[field] != 0 {
			value += s.fromBase[field] * baseFieldValue(base, field)
		}
		apply(current, base, field, value, mode)
	}
}

// baseFieldValue returns the target paint a replacing sentinel holds.
func baseFieldValue(base cssmotion.Paint, field int) float64 {
	switch field {
	case fieldOpacity:
		return base.Opacity
	case fieldR:
		return base.R
	case fieldG:
		return base.G
	case fieldB:
		return base.B
	}
	return 0
}

// bind precomputes the per-effect tables and the per-target compose scratch. It
// also parses every keyframe timing function once, because the compose path
// reads a frame's easing per sample.
func (p *Player) bind() {
	p.tables = make([]effectTable, len(p.effects))
	for i := range p.effects {
		ease := effectEase(p.effects[i].Timing)
		for j := range p.effects[i].Frames {
			frame := &p.effects[i].Frames[j]
			if frame.TimingFunction == "" {
				frame.ease = ease
			} else {
				frame.ease = cssEase(frame.TimingFunction)
			}
		}
		p.tables[i] = buildEffectTable(p.effects[i])
	}
	p.current = make([]cssmotion.Paint, len(p.targets))
}

// ensureBound binds lazily. NewComposite owns the other bind work, so the
// first Start or Step after construction is the earliest point this file can
// build the tables without changing new.go.
func (p *Player) ensureBound() {
	if len(p.tables) != len(p.effects) || len(p.current) != len(p.targets) {
		p.bind()
	}
}

// resetCursors drops the forward scan state after a resume or a clock that
// moved backwards.
func (p *Player) resetCursors() {
	for i := range p.tables {
		p.tables[i].resetCursors()
	}
}

func (t *effectTable) resetCursors() {
	t.lastNow = time.Time{}
	for i := range t.fields {
		t.fields[i].cursor, t.fields[i].lastAt, t.fields[i].sampled = 0, 0, false
	}
}
