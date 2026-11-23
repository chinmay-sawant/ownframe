package window

// uiaGeometryRefresh is the frame count after which an unchanged window is
// queried again anyway, so a move the cheap signature missed (a monitor DPI
// change, a client-area change from the OS) still reaches UI Automation.
const uiaGeometryRefresh = 120

// geometrySig is the cheap window state whose change can move the client
// area on screen: outer position, outer size, device scale, and the built-in
// title bar inset.
type geometrySig struct {
	x, y, w, h int
	scale      float64
	inset      int
}

// geometryThrottle decides whether placeUIA must re-query the native window
// (three Win32 calls plus a provider lock) on this frame. It is pure state
// so it is testable on every platform.
type geometryThrottle struct {
	last  geometrySig
	valid bool
	idle  int
}

// due reports whether geometry should be re-queried for sig. It resets the
// idle counter whenever it answers true.
func (g *geometryThrottle) due(sig geometrySig) bool {
	if g.valid && sig == g.last && g.idle < uiaGeometryRefresh {
		g.idle++
		return false
	}
	g.last, g.valid, g.idle = sig, true, 0
	return true
}

// forget makes the next call due, for example after a failed query.
func (g *geometryThrottle) forget() { g.valid = false }
