package runtimebehavior

import "time"

// Reset clears a loading marker.
func (m *LoadingMarker) Reset() {
	if m == nil {
		return
	}
	m.Label = ""
	m.Visible = false
	m.Started = time.Time{}
	m.Slow = false
}

// IsSlow reports whether elapsed time exceeds threshold.
func (m *LoadingMarker) IsSlow(now time.Time, threshold time.Duration) bool {
	if m == nil || !m.Visible {
		return false
	}
	return m.Elapsed(now) >= threshold
}

// Await resolves after setting fallback.
func (s *SuspensePlaceholder) Await(fallback string) {
	if s == nil {
		return
	}
	s.Fallback = fallback
	if !s.Resolved {
		s.Show(fallback)
	}
}

// IsResolved reports resolution state.
func (s *SuspensePlaceholder) IsResolved() bool {
	if s == nil {
		return false
	}
	return s.Resolved
}

// FallbackText returns the placeholder fallback.
func (s *SuspensePlaceholder) FallbackText() string {
	if s == nil {
		return ""
	}
	return s.Fallback
}
