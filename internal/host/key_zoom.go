package host

// KeyZoomPolicy reports whether Ctrl+=, Ctrl++, Ctrl+-, and Ctrl+0 may scale
// the page. A screen that does not implement it keeps the default key zoom.
type KeyZoomPolicy interface {
	KeyZoomAllowed() bool
}
