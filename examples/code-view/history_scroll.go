package main

// historyLoadStep is the smallest limit increase. Half of a short window is
// only ten rows, so the first fetches add at least twenty.
const historyLoadStep = 20

// historyScroll is the vertical window over the loaded history rail. Offset is
// the first visible row, Limit is how many commits the last fetch asked for,
// and Loading guards a fetch that is already in flight. Accum keeps the
// sub-row part of a wheel delta between events, so a slow trackpad still moves
// the rail once its fractions add up to a whole row.
type historyScroll struct {
	Offset  int
	Limit   int
	Loading bool
	Accum   float64
}

// Rows pins Offset to the scrollable range and returns the visible window as
// start inclusive and end exclusive. A history that fits leaves Offset at zero
// and shows every commit, and a window taller than the history still ends at
// total.
func (s *historyScroll) Rows(total, visible int) (start, end int) {
	if total < 1 || visible < 1 {
		s.Offset = 0

		return 0, 0
	}

	start = historyPinOffset(s.Offset, total, visible)
	s.Offset = start

	return start, min(start+visible, total)
}

// ScrollBy moves the window by whole rows and pins it to the history. It moves
// from the pinned position, so a scroll after the history shrank starts from
// the row the user can see.
func (s *historyScroll) ScrollBy(rows, total, visible int) {
	s.Offset = historyPinOffset(historyPinOffset(s.Offset, total, visible)+rows, total, visible)
}

// SetOffset jumps to a row and pins it to the history. No offset applies when
// the whole history fits, and the last offset keeps a full window.
func (s *historyScroll) SetOffset(offset, total, visible int) {
	s.Offset = historyPinOffset(offset, total, visible)
}

// Thumb lays the scrollbar thumb out over trackPx. total and visible are row
// counts, and minThumb is the shortest thumb in pixels. It reports show false
// when every row fits or there is no track, leaving the track empty. The thumb
// follows a pinned copy of Offset and does not write it back.
func (s *historyScroll) Thumb(trackPx, total, visible, minThumb int) (top, height int, show bool) {
	if trackPx < 1 || visible < 1 || total <= visible {
		return 0, 0, false
	}

	height = max(trackPx*visible/total, minThumb, 1)
	height = min(height, trackPx)

	offset := historyPinOffset(s.Offset, total, visible)
	top = (trackPx - height) * offset / historyMaxOffset(total, visible)

	return top, height, true
}

// ShouldLoad reports whether the rail has passed half of the rows it loaded
// and may fetch another page. It stays quiet while a fetch is in flight. A
// history shorter than Limit came back short of its request, so it is the end
// of history and nothing more is asked for.
func (s *historyScroll) ShouldLoad(total, visible int) bool {
	if s.Loading || s.Limit < 1 || total < 1 || visible < 1 {
		return false
	}

	if total < s.Limit {
		return false
	}

	return historyPinOffset(s.Offset, total, visible)+visible >= total/halfDivisor
}

// StartLoad marks a fetch as in flight so ShouldLoad stops asking for it.
func (s *historyScroll) StartLoad() { s.Loading = true }

// DoneLoad clears the in-flight mark once the fetched page is appended, or the
// fetch failed. The next redraw may ask again from the new offset.
func (s *historyScroll) DoneLoad() { s.Loading = false }

// WheelRows turns a wheel delta measured in rows into whole rows to scroll,
// keeping the fraction in Accum for the next event. A positive delta scrolls
// toward the end, so a caller passes -dy multiplied by its rows per notch.
func (s *historyScroll) WheelRows(delta float64) int {
	s.Accum += delta

	rows := int(s.Accum)
	s.Accum -= float64(rows)

	return rows
}

// NextHistoryLimit returns the limit for the next history fetch: half of the
// current limit, but at least historyLoadStep more rows, so 20 grows to 40,
// then 60, then 90. A positive max caps the result; zero or less grows without
// a cap.
func NextHistoryLimit(current, max int) int {
	if current < 0 {
		current = 0
	}

	step := current / halfDivisor
	if step < historyLoadStep {
		step = historyLoadStep
	}

	next := current + step
	if max > 0 && next > max {
		return max
	}

	return next
}

// historyMaxOffset is the last first-row offset that still fills visible rows.
func historyMaxOffset(total, visible int) int {
	if visible < 1 || total <= visible {
		return 0
	}

	return total - visible
}

// historyPinOffset pins offset between the first row and the last full window.
func historyPinOffset(offset, total, visible int) int {
	return max(0, min(offset, historyMaxOffset(total, visible)))
}
