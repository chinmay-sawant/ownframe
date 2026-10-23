package form

// Span is one control and where it sits in the executed HTML.
// Start is the index of '<'. End is the index after the element.
type Span struct {
	Control Control
	Start   int
	End     int
}
