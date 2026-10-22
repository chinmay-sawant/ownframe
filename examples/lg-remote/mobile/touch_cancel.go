package mobile

import "github.com/chinmay-sawant/ownframe"

var inputPage *ownframe.Page

// CancelTouches queues cancellation from the Android UI thread.
func CancelTouches() {
	if inputPage != nil {
		inputPage.CancelTouches()
	}
}
