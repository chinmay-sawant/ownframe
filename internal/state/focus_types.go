package state

// FocusState captures focus, caret, and selection for one field.
type FocusState struct {
	ID             string
	Caret          int
	SelectionStart int
	SelectionEnd   int
}

// ScrollState captures scroll offsets.
type ScrollState struct {
	X int
	Y int
}

// CaretState is an alias for FocusState.
type CaretState = FocusState

// SelectionState is an alias for FocusState.
type SelectionState = FocusState

// UIState groups retained focus and scroll state.
type UIState struct {
	Focus  FocusState
	Scroll ScrollState
}

// PageUIState is an alias.
type PageUIState = UIState

func focusMatches(a, b FocusState) bool { return a.ID == b.ID }

// KeepFocusState returns next with focus preserved if ID unchanged.
func KeepFocusState(prev, next FocusState) FocusState {
	if focusMatches(prev, next) || next.ID == "" {
		if next.ID == "" {
			next = prev
		} else {
			next.Caret = prev.Caret
			next.SelectionStart = prev.SelectionStart
			next.SelectionEnd = prev.SelectionEnd
		}
	}
	return next
}

// PreserveCaret returns prev caret when field identity unchanged.
func PreserveCaret(prev, next FocusState) FocusState {
	if prev.ID != "" && prev.ID == next.ID {
		next.Caret = prev.Caret
		next.SelectionStart = prev.SelectionStart
		next.SelectionEnd = prev.SelectionEnd
	}
	return next
}

// KeepScrollState returns prev scroll when next is zeroed.
func KeepScrollState(prev, next ScrollState) ScrollState {
	if next.X == 0 && next.Y == 0 && (prev.X != 0 || prev.Y != 0) {
		return prev
	}
	return next
}

// CombineState bundles focus and scroll.
func CombineState(f FocusState, sc ScrollState) UIState {
	return UIState{Focus: f, Scroll: sc}
}
