//go:build !js && !android && !ios

package main

import "log"

func goSafe(name string, fn func()) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("desktop cat %s panic: %v", name, r)
			}
		}()
		fn()
	}()
}
