package window

// Options controls the desktop window. Desktop placement and mouse
// passthrough do not apply to browser canvases or mobile views.
type Options struct {
	// Transparent keeps unpainted pixels transparent. Use transparent HTML
	// backgrounds and a replayable page; bitmap pages may paint a background.
	Transparent bool
	// Borderless removes the title bar and window borders.
	Borderless bool
	// Floating keeps the window above other normal windows.
	Floating bool
	// MousePassthrough passes all pointer input, including painted pixels,
	// to windows underneath. It also starts unfocused and ticks in background.
	// A full pass-through window never accepts input: Interactive is
	// ignored while this is set, so untrusted content cannot click.
	MousePassthrough bool
	// Interactive reports clickable window-local CSS pixels. When supplied,
	// only other pixels pass through. It runs on the window update thread.
	// Reserve only visibly rendered app pixels (never empty margins), and
	// only app code may set it: page content cannot toggle passthrough.
	Interactive func(x, y int) bool
	// Draggable starts a window drag only on a user press in this region.
	Draggable func(x, y int) bool
	// FixedSize disables resizing by the user.
	FixedSize bool
	// Perf samples frame times, stage timings, and runtime numbers for the
	// DevTools performance rows. It is off by default; end users pay
	// nothing unless a developer opts in.
	Perf bool
	// BottomRight places the window in the current monitor's bottom right.
	BottomRight bool
	// Margin is the inset in CSS pixels, clamped to zero. Placement uses
	// monitor bounds, which include taskbars and docks.
	Margin int
	// NoIdlePace keeps the window presenting at the vsync frame rate even
	// while the page is idle. Pacing is on by default on the desktop: an
	// idle page stops presenting and runs page ticks at a low keep-alive
	// rate, and input or damage returns the window to vsync.
	NoIdlePace bool
}
