//go:build android || ios

package mobile

// app is the bound screen. A phone build sets it in start; a desktop build
// leaves it nil and every exported call below is a no-op.
// (app is declared in start_mobile.go)

// SetMotion feeds screen-aligned tilt deltas, measured in g, to the bound page.
func SetMotion(x, y float64) {
	if app == nil {
		return
	}
	app.SetMotion(x, y)
}

// Shake forces the dizzy/spiral expression, used by the Android sensor.
func Shake() {
	if app != nil {
		app.Shake()
	}
}
