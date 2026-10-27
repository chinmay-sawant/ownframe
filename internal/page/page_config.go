package page

import "github.com/chinmay-sawant/ownframe/internal/host"

// Config is the template, the optional theme, and the frame size for a new
// page. Theme is an extra stylesheet applied after the template's own styles;
// empty means no theme.
// Width and Height are the first frame, in CSS pixels.
// MinWidth and MinHeight are the smallest frame. Zero means 1.
// MaxWidth and MaxHeight ask the host to cap the window. Zero means no cap.
// They do not cap the picture, so the layout follows the window.
type Config struct {
	Title string
	HTML  string
	// File reads the page source from disk at New and watches it. Setting
	// File and HTML together is an error.
	File  string
	Theme string
	// ThemeFile reads the theme stylesheet from disk and watches it.
	ThemeFile string
	// DisableHotReload turns the file watch off. It is ignored without File
	// or ThemeFile.
	DisableHotReload bool
	// RetainState keeps the layout state a single element patch needs. It is
	// off by default, so a page that never patches pays nothing. With it on,
	// Page.PatchElement can re-lay out one element.
	RetainState bool
	// DevTools starts the window overlay on.
	DevTools bool
	// Perf records pipeline timing, dirty counts, and allocs in Stats. It
	// is off by default, so end users pay nothing; SetPerf opts in.
	Perf bool
	// LockView scales the page to the window and turns off scroll and zoom.
	LockView bool
	// Scrollbar selects thumb visibility. Auto shows thumbs except on
	// Android and iOS. Visible and Hidden override that platform default.
	Scrollbar host.ScrollbarMode
	// DisableTouchZoom keeps two-finger touch gestures from zooming the page.
	// Wheel and keyboard zoom remain available unless LockView,
	// DisableWheelZoom, or DisableKeyZoom is also set.
	DisableTouchZoom bool
	// DisableWheelZoom keeps Ctrl+wheel and Command+wheel from scaling the
	// page. The wheel scrolls. Keyboard zoom is unchanged.
	DisableWheelZoom bool
	// DisableKeyZoom keeps Ctrl+=, Ctrl++, Ctrl+-, and Ctrl+0 from scaling
	// the page. The window still consumes those chords. Wheel zoom is
	// unchanged.
	DisableKeyZoom bool
	// DeepLinkSchemes lists URL schemes accepted by DispatchDeepLink.
	DeepLinkSchemes []string
	Width           int
	Height          int
	MinWidth        int
	MinHeight       int
	MaxWidth        int
	MaxHeight       int
}
