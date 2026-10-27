package host

// WheelZoomPolicy reports whether Ctrl+wheel or Command+wheel may scale the
// page. A screen that does not implement it keeps the default wheel zoom.
type WheelZoomPolicy interface {
	WheelZoomAllowed() bool
}
