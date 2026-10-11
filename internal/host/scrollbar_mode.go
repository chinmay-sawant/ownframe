package host

// ScrollbarMode controls whether the window paints and handles page thumbs.
type ScrollbarMode uint8

const (
	// ScrollbarAuto shows thumbs except on Android and iOS.
	ScrollbarAuto ScrollbarMode = iota
	// ScrollbarVisible shows thumbs on every platform.
	ScrollbarVisible
	// ScrollbarHidden hides thumbs on every platform.
	ScrollbarHidden
)

// ScrollbarModeProvider optionally sets a screen's scrollbar mode.
type ScrollbarModeProvider interface {
	ScrollbarMode() ScrollbarMode
}
