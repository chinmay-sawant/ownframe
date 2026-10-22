package mobile

import "github.com/chinmay-sawant/ownframe/examples/depth-alert/depthalert"

// app is set by start on mobile platforms.
var app *depthalert.App

// ObserveDepth passes a center-view measurement and returns an alert state.
func ObserveDepth(nearMM, coverage int32, tracking bool) int32 {
	if app == nil {
		return 0
	}
	return app.ObserveDepth(nearMM, coverage, tracking)
}

// DepthUnavailable clears measurements when tracking or the camera stops.
func DepthUnavailable() int32 {
	if app == nil {
		return 0
	}
	return app.Unavailable()
}

// SetScanning resets the classifier across explicit start and stop actions.
func SetScanning(on bool) {
	if app != nil {
		app.SetScanning(on)
	}
}

// CancelTouches ends any pointer gesture interrupted by Android lifecycle changes.
func CancelTouches() {
	if app != nil {
		app.Page().Testing().CancelTouches()
	}
}
