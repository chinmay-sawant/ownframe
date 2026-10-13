// Package runtimebehavior adds browser-like UX helpers for partial updates.
package runtimebehavior

import "time"

// LoadingMarker is shown in place of a whole-page fallback while async data
// loads.
type LoadingMarker struct {
	Label   string
	Visible bool
	Started time.Time
	Slow    bool
}

// Show marks the marker visible.
func (m *LoadingMarker) Show(label string) {
	if m == nil {
		return
	}
	m.Label = label
	m.Visible = true
	m.Started = time.Now()
	m.Slow = false
}

// Start records the load start without changing visibility.
func (m *LoadingMarker) Start(now time.Time) {
	if m == nil {
		return
	}
	m.Started = now
	m.Slow = false
}

// Hide clears the marker.
func (m *LoadingMarker) Hide() {
	if m == nil {
		return
	}
	m.Visible = false
	m.Slow = false
}

// Elapsed reports how long the load has been running.
func (m *LoadingMarker) Elapsed(now time.Time) time.Duration {
	if m == nil || m.Started.IsZero() {
		return 0
	}
	if now.Before(m.Started) {
		return 0
	}
	return now.Sub(m.Started)
}

// MarkSlow sets Slow once elapsed reaches threshold.
func (m *LoadingMarker) MarkSlow(now time.Time, threshold time.Duration) bool {
	if m == nil || !m.Visible {
		return false
	}
	if threshold <= 0 {
		return false
	}
	if m.Elapsed(now) >= threshold {
		m.Slow = true
	}
	return m.Slow
}

// SuspensePlaceholder is a lightweight placeholder for deferred data.
type SuspensePlaceholder struct {
	LoadingMarker
	Resolved bool
	Fallback string
	Data     any
}

// Resolve marks the deferred data as loaded.
func (s *SuspensePlaceholder) Resolve() {
	if s == nil {
		return
	}
	s.Resolved = true
	s.Hide()
}

// SetData stores deferred data and resolves the placeholder.
func (s *SuspensePlaceholder) SetData(data any) {
	if s == nil {
		return
	}
	s.Data = data
	s.Resolve()
}
