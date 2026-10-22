//go:build android

package mobile

import "github.com/chinmay-sawant/ownframe"

func setAndroidPageProfile(page *ownframe.Page) {
	page.Window().SetTouchZoomAllowed(false)
	page.Window().SetPerf(true)
}
