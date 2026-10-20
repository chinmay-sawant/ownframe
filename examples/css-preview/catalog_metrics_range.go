package main

// rangeFor returns the visible row range with two rows of overscan.
func (m *catalogMetrics) rangeFor(offsetY, viewH int) (int, int) {
	rel := offsetY - m.gridTop
	start := 0
	for start < len(m.rowTop) && m.rowTop[start]+m.rowH[start]-m.gridTop < rel {
		start++
	}
	start -= catalogWindowOverscan
	if start < 0 {
		start = 0
	}

	end := start
	limit := rel + viewH
	for end < len(m.rowTop) && m.rowTop[end]-m.gridTop <= limit {
		end++
	}
	end += catalogWindowOverscan
	if end > len(m.rowTop) {
		end = len(m.rowTop)
	}
	if end <= start {
		end = start + 1
		if end > len(m.rowTop) {
			end = len(m.rowTop)
		}
	}

	return start, end
}

func (m *catalogMetrics) topPad(start int) int {
	if start <= 0 {
		return 0
	}
	pad := m.rowTop[start] - m.gridTop - m.gap
	if pad < 0 {
		return 0
	}

	return pad
}

func (m *catalogMetrics) botPad(end int) int {
	if end >= len(m.rowTop) {
		return 0
	}
	last := m.rowTop[end-1] + m.rowH[end-1]
	pad := m.gridBottom - last - m.gap
	if pad < 0 {
		return 0
	}

	return pad
}
