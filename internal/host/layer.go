package host

import (
	"math"

	"github.com/chinmay-sawant/blinkless/layout"
)

// LayerSpec asks the renderer to composite one element as a retained layer.
// A page supplies these through LayerSource; the renderer owns resolution,
// caching, and damage.
//
// Matrix and Base are both display point-space matrices. Matrix is the
// animation delta from the captured pose to the current pose. Base maps the
// operation transforms the display already carries back to the captured pose;
// a page whose display carries the sampled animation matrix passes
// InvertLayerMatrix(Matrix) here, and the layer captures the untransformed
// pose. A zero matrix means the identity for both fields.
//
// Alpha is the element's group opacity, 0..1. The zero value is fully
// transparent, so a caller that does not animate opacity sets 1. Mode is the
// element's mix-blend-mode, empty for normal compositing. A non-normal mode is
// supported only with Alpha 1 and a translation matrix; other combinations
// stay direct until the shared blend shader takes an alpha uniform.
type LayerSpec struct {
	// ElementID names the element across frames; it joins the box index and
	// generation in the cache key.
	ElementID string
	// BoxIndex is the element's index in Display.Boxes.
	BoxIndex int
	// Matrix is the animation delta in display point space.
	Matrix layout.Matrix2D
	// Base maps the display's current operation matrices back to the captured
	// pose. Its zero value is the identity.
	Base layout.Matrix2D
	// Alpha is the element's group opacity, 0..1; see the type comment.
	Alpha float64
	// Mode is the element's mix-blend-mode, empty for normal compositing.
	Mode string
}

// LayerSource is the optional screen role the window asks for this frame's
// promoted elements. A screen without it keeps the direct replay path.
//
// The interface is deliberately small: the page owns animation state and knows
// which elements are active, while the renderer owns the layer cache and the
// paint skip. The page hook needed to turn promotion on is a method that
// returns only elements whose active animation carries transform, opacity,
// filter, or clip-path, with their 2D-only delta matrix and group alpha.
type LayerSource interface {
	LayerSpecs(display *layout.Display) []LayerSpec
}

// layerDetEpsilon is the smallest determinant a 2D affine matrix may carry and
// still be inverted for a layer.
const layerDetEpsilon = 1e-12

// InvertLayerMatrix returns the inverse of a 2D layer matrix. It is exported
// so a page can derive LayerSpec.Base from the sampled operation matrix.
func InvertLayerMatrix(m layout.Matrix2D) (layout.Matrix2D, bool) {
	det := m.A*m.D - m.B*m.C
	if math.Abs(det) < layerDetEpsilon {
		return layout.Matrix2D{}, false
	}
	inv := 1 / det

	return layout.Matrix2D{
		A: m.D * inv,
		B: -m.B * inv,
		C: -m.C * inv,
		D: m.A * inv,
		E: (m.C*m.F - m.D*m.E) * inv,
		F: (m.B*m.E - m.A*m.F) * inv,
	}, true
}
