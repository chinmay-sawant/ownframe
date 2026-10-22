package ui

import "fmt"

// View is the printable page state. A handler rebuilds it with syncView and
// calls SetData; the tick then edits retained paint operations on top of it.
type View struct {
	Dark    bool
	Dummy   bool
	DataDir string
	Notice  string

	Active  []Row
	Detail  *Row
	Summary Summary
	Stats   Stats

	Filter  Filter
	History []Row
	Page    int
	Pages   int
	Total   int
	HasPrev bool
	HasNext bool
	Loading bool
}

// Stats is the measured tick and repaint work the footer shows.
type Stats struct {
	LastTickUS int64
	Paints     int64
	Redraws    int64
	Applied    int64
}

// TickText is the last tick duration in milliseconds.
func (s Stats) TickText() string {
	return fmt.Sprintf("%.2f ms", float64(s.LastTickUS)/microsPerMilli)
}
