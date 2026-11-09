package render

import "github.com/chinmay-sawant/blinkless/css"

// State carries the runtime pointer and focus state the engine's stateful
// pseudo-classes match: the ids of the focused, hovered, and pressed elements.
// An empty id matches none of them.
type State struct {
	Focus  string
	Hover  string
	Active string
	// Theme is an extra stylesheet applied after the page's own styles, so
	// it can override any rule they set. Nil adds nothing.
	Theme *css.Sheet
	// Images returns encoded image bytes for one src, such as an <img src>
	// value or a CSS background-image url(...) target. Nil resolves none.
	Images func(src string) ([]byte, error)
	// RetainState keeps the layout state an element patch needs: the
	// resolved styles of the document and the box tree the placement
	// produced. It is off by default, so a host that never patches an
	// element does not pay for the retained state.
	RetainState bool
}

// options builds the screen cascade options for a viewport.
func (s State) options(width, height int) css.Options {
	return css.Options{
		WidthPx:  width,
		HeightPx: height,
		Media:    "screen",
		Extra:    s.extra(),
		Focus:    s.Focus,
		Hover:    s.Hover,
		Active:   s.Active,
	}
}

// extra returns the theme as the engine's extra-sheet list.
func (s State) extra() []*css.Sheet {
	if s.Theme == nil {
		return nil
	}

	return []*css.Sheet{s.Theme}
}

// SameStyleState reports that moving from prev to s cannot change which
// rules match: every pointer or focus id that differs names a pseudo-class
// the sheets do not use. The caller may reuse the styled document built for
// prev instead of re-cascading. Only the stateful pseudo-classes are
// compared here; the viewport, the source, the theme, and the image inputs
// are the caller's to check.
func (s State) SameStyleState(prev State, pseudos StatePseudos) bool {
	if pseudos.Hover && s.Hover != prev.Hover {
		return false
	}

	if pseudos.Active && s.Active != prev.Active {
		return false
	}

	if pseudos.Focus && s.Focus != prev.Focus {
		return false
	}

	return true
}
