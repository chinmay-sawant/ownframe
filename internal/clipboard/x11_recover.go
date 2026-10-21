//go:build linux && !android

package clipboard

import (
	"fmt"
	"log"
)

func recoverX(x *xClient, where string) {
	if r := recover(); r != nil {
		log.Printf("ownframe clipboard: %s panic: %s", cleanLog(where), cleanLog(fmt.Sprint(r)))
		x.fail()
	}
}
