package motion_test

import (
	"testing"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/internal/motion"
)

// colorCaptures returns a color transition's display and box, the captured
// indices, the start paint of the old display, and the destination paint of
// the new one.
func colorCaptures(t *testing.T) (*ownframe.Display, ownframe.Box, []int,
	[]motion.Paint, []motion.Paint) {
	t.Helper()

	oldDisplay, oldBox := colorFrame(t, "#000000")
	newDisplay, box := colorFrame(t, "#ffffff")

	return newDisplay, box, motion.OperationIndices(newDisplay, box),
		motion.Capture(oldDisplay, oldBox), motion.Capture(newDisplay, box)
}
