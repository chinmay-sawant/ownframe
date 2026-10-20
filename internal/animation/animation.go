// Package animation samples CSS-style keyframes onto retained display ops.
package animation

import (
	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/internal/cssmotion"
)

// Easing maps progress within one keyframe interval to eased progress.
type Easing func(float64) float64

// Target pairs a display operation with its captured starting paint.
type Target struct {
	Op   *ownframe.DisplayOp
	From cssmotion.Paint
}

// Keyframe stores values declared at Offset. X and Y are display-list points.
type Keyframe struct {
	Offset                     float64
	X, Y                       float64
	Opacity                    float64
	R, G, B                    float64
	Move, OpacitySet, ColorSet bool
}
