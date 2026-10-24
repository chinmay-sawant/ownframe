package csstransition_test

import (
	"testing"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/internal/cssmotion"
)

// colorCaptures returns a color transition's display and box, the captured
// indices, the start paint of the old display, and the destination paint of
// the new one.
func colorCaptures(t *testing.T) (*ownframe.Display, ownframe.Box, []int,
	[]cssmotion.Paint, []cssmotion.Paint) {
	t.Helper()

	oldDisplay, oldBox := colorFrame(t, "#000000")
	newDisplay, box := colorFrame(t, "#ffffff")

	return newDisplay, box, cssmotion.OperationIndices(newDisplay, box),
		cssmotion.Capture(oldDisplay, oldBox), cssmotion.Capture(newDisplay, box)
}
