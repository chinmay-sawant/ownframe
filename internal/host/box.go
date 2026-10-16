package host

import "github.com/chinmay-sawant/blinkless/layout"

// Boxer lists the clickable boxes on a screen, innermost last. web/areas
// takes only this so its fakes stay one method.
type Boxer interface {
	Boxes() []layout.Box
}
