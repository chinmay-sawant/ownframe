package animation

import (
	"time"

	"github.com/chinmay-sawant/ownframe/internal/cssmotion"
)

// Player applies animation effects to retained display operations.
type Player struct {
	targets []Target
	effects []Effect
	states  []effectState
	tables  []effectTable
	current []cssmotion.Paint
	start   time.Time
	started bool
	active  bool
	touched bool
	// restampSlots holds the latest sample per target that the 2D matrix path
	// could not write. restamp is the exposed form, one entry per such target,
	// and restampBuf is the serialization scratch shared by every target.
	restampSlots []restampSlot
	restamp      []RestampValue
	restampBuf   []byte
}

// restampSlot is one target's latest sampled values that need the engine's
// re-projection pass. A transform slot is kept even after a 2D matrix write,
// because a perspective request for the same element re-stamps the element's
// own transform too.
type restampSlot struct {
	transform        TransformValue
	perspective      PerspectiveValue
	hasTransform     bool
	transformRestamp bool
	hasPerspective   bool
}

// active reports whether the slot asks for one re-projection request.
func (s *restampSlot) active() bool {
	return s.transformRestamp || s.hasPerspective
}

// reset clears every sampled value before the next Step.
func (s *restampSlot) reset() {
	s.transform = TransformValue{}
	s.perspective = PerspectiveValue{}
	s.hasTransform = false
	s.transformRestamp = false
	s.hasPerspective = false
}
