package main

// changesScroll is the vertical scroll model for the Source Control
// changed-file list. Offset is the first visible row. Accum keeps the sub-row
// part of a wheel delta between events, so a trackpad moves the list once its
// fractions add up to a whole row.
type changesScroll struct {
	Offset int
	Accum  float64
}

// Rows pins Offset to the scrollable range and returns the visible window as
// start inclusive and end exclusive. A list that fits leaves Offset at zero
// and shows every row. A window taller than the list still ends at total.
func (s *changesScroll) Rows(total, visible int) (start, end int) {
	start = changesClampOffset(s.Offset, total, visible)
	s.Offset = start

	end = min(start+visible, total)
	if end < start {
		end = start
	}

	return start, end
}

// ScrollBy moves Offset by rows, where a positive count moves toward the end
// of the list, and pins the result to the scrollable range. A wheel notch or a
// key repeats the same call.
func (s *changesScroll) ScrollBy(rows, total, visible int) {
	s.Offset = changesClampOffset(s.Offset+rows, total, visible)
}

// SetOffset jumps to a row and pins it to the scrollable range.
func (s *changesScroll) SetOffset(offset, total, visible int) {
	s.Offset = changesClampOffset(offset, total, visible)
}

// Page moves one window less a row in direction: negative for PageUp and
// positive for PageDown. The row of overlap keeps the reading position in
// view. A one-row window still moves one row.
func (s *changesScroll) Page(direction, total, visible int) {
	step := max(visible-1, 1)
	s.ScrollBy(direction*step, total, visible)
}

// Thumb lays the scrollbar thumb out over trackPx. total and visible are row
// counts, and minThumb is the shortest thumb in pixels. It reports show false
// when every row fits or there is no track, leaving the track empty. The thumb
// follows a pinned copy of Offset and does not write it back.
func (s *changesScroll) Thumb(trackPx, total, visible, minThumb int) (top, height int, show bool) {
	if trackPx < 1 || visible < 1 || total <= visible {
		return 0, 0, false
	}

	height = trackPx * visible / total
	height = min(max(height, minThumb, 1), trackPx)

	offset := changesClampOffset(s.Offset, total, visible)
	travel := trackPx - height
	top = travel * offset / changesMaxOffset(total, visible)

	return top, height, true
}

// WheelRows turns a wheel delta measured in rows into whole rows to scroll,
// keeping the fraction in Accum for the next event. A positive delta scrolls
// toward the end, so a caller passes -dy multiplied by its rows per notch.
func (s *changesScroll) WheelRows(delta float64) int {
	s.Accum += delta
	rows := int(s.Accum)
	s.Accum -= float64(rows)

	return rows
}

// changesMaxOffset is the last Offset that still fills visible rows.
func changesMaxOffset(total, visible int) int {
	if visible < 1 || total <= visible {
		return 0
	}

	return total - visible
}

// changesClampOffset pins offset between the first row and the last full
// window.
func changesClampOffset(offset, total, visible int) int {
	return max(0, min(offset, changesMaxOffset(total, visible)))
}
