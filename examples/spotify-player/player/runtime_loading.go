package player

import "time"

type loadingMarker struct {
	Label   string
	Visible bool
	Started time.Time
	Slow    bool
}

func (m *loadingMarker) Show(label string) {
	if m == nil {
		return
	}
	m.Label = label
	m.Visible = true
	m.Started = time.Now()
	m.Slow = false
}

func (m *loadingMarker) Hide() {
	if m == nil {
		return
	}
	m.Visible = false
	m.Slow = false
}

func (m *loadingMarker) elapsed(now time.Time) time.Duration {
	if m == nil || m.Started.IsZero() || now.Before(m.Started) {
		return 0
	}
	return now.Sub(m.Started)
}

func (m *loadingMarker) MarkSlow(now time.Time, threshold time.Duration) bool {
	if m == nil || !m.Visible || threshold <= 0 {
		return false
	}
	if m.elapsed(now) >= threshold {
		m.Slow = true
	}
	return m.Slow
}

func (m *loadingMarker) Reset() {
	if m == nil {
		return
	}
	m.Label = ""
	m.Visible = false
	m.Started = time.Time{}
	m.Slow = false
}
