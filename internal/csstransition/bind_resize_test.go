package csstransition

import (
	"testing"
	"time"

	"github.com/chinmay-sawant/ownframe/internal/cssmotion"
)

// TestBindToKeepsTheSizeAfterAResize checks that a rebind after the layout
// resized the operation keeps the operation's current size instead of
// writing the captured size back.
func TestBindToKeepsTheSizeAfterAResize(t *testing.T) {
	display, box, indices, from, to := colorCaptures(t)
	ops := cssmotion.Resolve(display, indices)

	resized := to[0]
	resized.W = 120
	cssmotion.Write(ops[0], resized)

	start := time.Unix(8, 0)
	r, err := BindTo(display, box, "background-color", indices, from, to, start)
	if err != nil {
		t.Fatal(err)
	}
	r.Step(start.Add(time.Second))

	if got := cssmotion.Read(ops[0]); got.W != 120 {
		t.Fatalf("the rebind wrote the captured width back: %+v", got)
	}
}
