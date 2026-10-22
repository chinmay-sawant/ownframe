//go:build !android

package mobile

import "github.com/chinmay-sawant/ownframe"

func setAndroidPageProfile(*ownframe.Page) {} //nolint:unused // reason: called from start_mobile.go on ios builds; no caller on desktop.
