package mobile

import "github.com/chinmay-sawant/ownframe/examples/telegram/telegram"

// app is the bound screen. A phone build sets it in start; a desktop build
// leaves it nil and every exported call below is a no-op.
var app *telegram.App

// Back accepts the Android back key. It reports false when the app is
// already on a list, so the activity can close.
func Back() bool {
	if app == nil {
		return false
	}

	return app.RequestBack()
}

// AttachRequest reports and clears a pending photo request as "camera",
// "gallery", or "" when there is none.
func AttachRequest() string {
	if app == nil {
		return ""
	}

	return app.TakeAttach()
}

// SetPhoto hands captured image bytes to the page.
func SetPhoto(data []byte) {
	if app != nil {
		app.QueuePhoto(data)
	}
}

// SetInsets passes the system bar insets in CSS pixels to the page.
func SetInsets(top, bottom int) {
	if app != nil {
		app.SetInsets(top, bottom)
	}
}

// Blur queues keyboard focus dismissal on the game loop.
func Blur() {
	if app != nil {
		app.RequestBlur()
	}
}
