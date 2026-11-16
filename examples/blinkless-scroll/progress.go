package main

// Scroll and view timeline math. A timeline position is a number of CSS
// pixels along the timeline; progress is that position divided by the length
// of the range being read, clamped to 0..1.

func clamp01(v float64) float64 {
	switch {
	case v < 0:
		return 0
	case v > 1:
		return 1
	}

	return v
}

// ScrollProgress is offset / (scrollHeight - viewport) on the axis. A page
// that cannot scroll reports 0.
func ScrollProgress(v Viewport, axis Axis) float64 {
	max := v.content(axis) - v.size(axis)
	if max <= 0 {
		return 0
	}

	return clamp01(v.offset(axis) / max)
}

// RangeSpace is the ruler a range bound is measured on: the total length and
// the start and end position of each named range.
type RangeSpace struct {
	Total float64
	Named map[string][2]float64
}

// ViewState is where a subject sits against a scroller right now.
type ViewState struct {
	// T is how far the subject has travelled: 0 when its leading edge first
	// touches the far edge of the scrollport, Total when its trailing edge
	// leaves the near edge.
	T     float64
	Space RangeSpace
}

// ViewTimeline places subject in the scrollport shrunk by the two insets.
// Named ranges follow CSS: cover is the whole trip, entry is the subject
// sliding in, exit is it sliding out, contain is the stretch where the
// smaller of the two boxes sits fully inside the larger.
func ViewTimeline(v Viewport, subject Rect, axis Axis, insetStart, insetEnd float64) ViewState {
	port := v.size(axis)
	near := insetStart
	far := port - insetEnd
	d := far - near // scrollport length after insets
	l := subject.size(axis)
	screen := subject.start(axis) - v.offset(axis) // subject leading edge in the window

	total := d + l
	space := RangeSpace{
		Total: total,
		Named: map[string][2]float64{
			"cover": {0, total},
			"entry": {0, l},
			"exit":  {d, total},
		},
	}

	lo, hi := l, d
	if lo > hi {
		lo, hi = hi, lo
	}

	space.Named["contain"] = [2]float64{lo, hi}

	return ViewState{T: far - screen, Space: space}
}

// Progress reads the state against a named range. An empty or unknown name
// reads cover. A zero-length range is 1 once the position reaches it.
func (s ViewState) Progress(name string) float64 {
	r, ok := s.Space.Named[name]
	if !ok {
		r = s.Space.Named["cover"]
	}

	span := r[1] - r[0]
	if span <= 0 {
		if s.T >= r[0] {
			return 1
		}

		return 0
	}

	return clamp01((s.T - r[0]) / span)
}

// ScrollSpace is the ruler of a scroll timeline: 0 to the scroll distance.
func ScrollSpace(v Viewport, axis Axis) (pos float64, space RangeSpace) {
	max := v.content(axis) - v.size(axis)
	if max < 0 {
		max = 0
	}

	return v.offset(axis), RangeSpace{Total: max, Named: map[string][2]float64{"cover": {0, max}}}
}

// Resolve turns a bound into a position on the ruler. isEnd picks which edge
// of a bare range name the bound means.
func (s RangeSpace) Resolve(b RangeBound, isEnd bool) float64 {
	if b.Name == "" {
		if !b.HasPercent && !b.HasLength {
			if isEnd {
				return s.Total
			}

			return 0
		}

		pos := 0.0
		if b.HasPercent {
			pos += s.Total * b.Percent / 100
		}

		if b.HasLength {
			pos += b.Length
		}

		return pos
	}

	r, ok := s.Named[b.Name]
	if !ok {
		r = [2]float64{0, s.Total}
	}

	base := r[0]
	if isEnd && !b.HasPercent && !b.HasLength {
		return r[1]
	}

	if b.HasPercent {
		base += (r[1] - r[0]) * b.Percent / 100
	}

	if b.HasLength {
		base += b.Length
	}

	return base
}
