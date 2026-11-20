package motion_test

import (
	"testing"
	"time"

	"github.com/chinmay-sawant/ownframe/internal/motion"
)

// TestBindToKeepsTheSizeAfterAResize checks that a rebind after the layout
// resized the operation keeps the operation's current size instead of
// writing the captured size back.
func TestBindToKeepsTheSizeAfterAResize(t *testing.T) {
	display, box, indices, from, to := colorCaptures(t)
	ops := motion.Resolve(display, indices)

	resized := to[0]
	resized.W = 120
	motion.Write(ops[0], resized)

	start := time.Unix(8, 0)
	r, err := motion.BindTransitionTo(display, box, "background-color", indices, from, to, start)
	if err != nil {
		t.Fatal(err)
	}
	r.Step(start.Add(time.Second))

	if got := motion.Read(ops[0]); got.W != 120 {
		t.Fatalf("the rebind wrote the captured width back: %+v", got)
	}
}
