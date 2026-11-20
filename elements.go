// Package ownframe re-exports the retained element model.
package ownframe

import (
	"github.com/chinmay-sawant/ownframe/internal/element"
	"github.com/chinmay-sawant/ownframe/internal/frame"
)

// Element is a retained element instance.
type Element = element.Node

// ElementKind is an element kind label.
type ElementKind = element.Kind

// ElementRegistry retains element instances by stable key.
type ElementRegistry = element.Registry

// ElementLayout holds CSS layout geometry apart from metadata.
type ElementLayout = element.Layout

// NewElement returns an empty retained node.
func NewElement(id string, kind ElementKind) *Element {
	return element.NewNode(id, kind)
}

// NewElementRegistry returns an empty retained-element registry.
func NewElementRegistry() *ElementRegistry {
	return element.NewRegistry()
}

// BoxUnits converts a hit-test box to display-list units: x, y, w, h.
// Display operations carry points; boxes carry CSS pixels.
func BoxUnits(d *Display, b Box) (x, y, w, h float64) {
	return frame.BoxUnits(d, b)
}

// Fill returns the first fill rectangle of the given 0..1 color inside box,
// or nil. Fills returns them all, left to right.
func Fill(d *Display, box Box, color [3]float64) *DisplayOp {
	return frame.Fill(d, box, color)
}

// Fills returns the fill rectangles of the given color inside box, ordered
// left to right, or nil.
func Fills(d *Display, box Box, color [3]float64) []*DisplayOp {
	return frame.Fills(d, box, color)
}

// Text returns the first text operation inside box, or nil. A caller can
// change its Text field to show a new value on the next frame.
func Text(d *Display, box Box) *DisplayOp {
	return frame.Text(d, box)
}
