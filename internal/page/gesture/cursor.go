package gesture

import "github.com/chinmay-sawant/ownframe/internal/host"

// Cursor is the hit under the pointer, already resolved by the page.
type Cursor struct {
	HoverID  string
	Control  bool
	Editable bool
	Tag      string
	Action   string
	Link     bool
}

// CursorShape returns the shape for the hovered box: an I-beam over a text
// control, a hand over a link, a button, or another control, and the default
// arrow everywhere else.
func CursorShape(in Cursor) host.Shape {
	if in.HoverID != "" {
		if in.Control {
			if in.Editable {
				return host.ShapeText
			}

			return host.ShapePointer
		}

		switch in.Tag {
		case "a", "button":
			return host.ShapePointer
		}

		if in.Action != "" {
			return host.ShapePointer
		}
	}

	if in.Link {
		return host.ShapePointer
	}

	return host.ShapeDefault
}
