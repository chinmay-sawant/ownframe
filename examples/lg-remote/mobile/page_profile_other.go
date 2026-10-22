//go:build !android

package mobile

import "github.com/chinmay-sawant/ownframe"

func setAndroidPageProfile(*ownframe.Page) {} //nolint:unused // reason: used on ios builds via start_mobile.go (android||ios); this stub covers !android.
