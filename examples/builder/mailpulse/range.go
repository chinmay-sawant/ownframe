package main

import "strings"

// rangeDays maps the header range pill to a day count. Unknown labels fall
// back to the default 30-day window.
func (a *app) rangeDays() int {
	switch strings.TrimSpace(a.rangeLabel) {
	case "Last 7 Days":
		return 7
	case "Last 90 Days":
		return 90
	default:
		return 30
	}
}

// rangeWindow returns the lower-cased window phrase for card descriptions,
// e.g. "last 7 days".
func (a *app) rangeWindow() string {
	switch a.rangeDays() {
	case 7:
		return "last 7 days"
	case 90:
		return "last 90 days"
	default:
		return "last 30 days"
	}
}

// rangeNewLabel returns the short "New (7d)/(30d)/(90d)" KPI label.
func (a *app) rangeNewLabel() string {
	switch a.rangeDays() {
	case 7:
		return "New (7d)"
	case 90:
		return "New (90d)"
	default:
		return "New (30d)"
	}
}
