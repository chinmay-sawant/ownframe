// Package safego launches background goroutines that survive panics.
// A panic is reported through the crash package and logged instead of
// killing the whole app. Use it for clipboard, notify, and timer loops.
package safego

import (
	"fmt"
	"log"

	"github.com/chinmay-sawant/ownframe/internal/crash"
)

// Go runs fn on a new goroutine and recovers a panic into a crash
// report. name identifies the loop in the report and the log line.
func Go(name string, fn func()) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				reason := fmt.Sprintf("%s panic: %v", name, r)
				if path, err := crash.Write(name, reason); err != nil {
					log.Printf("ownframe %s panic: %v (report failed: %v)", name, r, err)
				} else {
					log.Printf("ownframe %s panic: %v (%s)", name, r, path)
				}
			}
		}()

		fn()
	}()
}
