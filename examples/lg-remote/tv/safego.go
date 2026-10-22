package tv

import "log"

func goSafe(name string, fn func()) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("lg-remote %s panic: %v", name, r)
			}
		}()
		fn()
	}()
}
