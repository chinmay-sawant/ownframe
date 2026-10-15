//go:build linux && !android

package clipboard

import "log"

func recoverX(x *xClient, where string) {
	if r := recover(); r != nil {
		log.Printf("ownframe clipboard: %s panic: %v", where, r)
		x.fail()
	}
}
