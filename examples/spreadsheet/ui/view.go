package ui

import (
	"html/template"
)

// View is the template data for one frame. All coordinates are content
// pixels: the chrome bands follow the scroll offset, the cells sit at their
// natural position.
type View struct {
	Dark    bool
	ViewW   int
	ViewH   int
	TotalW  int
	TotalH  int
	ScrollX int
	ScrollY int
	Ref     string
	Formula string
	Status  string
	Toolbar []Box
	Grid    []Box
	SelRect *Box
	GridX   int
	GridY   int
	GridW   int
	GridH   int
	Cols    []Box
	Rows    []Box
	Cells   []Box
	Editor  *Editor
	Tabs    []Box
	Dialog  *Dialog
}

// Box is one absolutely positioned element with a click action. Action is
// typed as template.URL because html/template treats data-action as a URL
// and would replace a colon-bearing value with #ZgotmplZ.
type Box struct {
	Action template.URL
	Text   string
	Class  string
	X      int
	Y      int
	W      int
	H      int
}

// act builds a data-action token from a computed string.
func act(s string) template.URL { return template.URL(s) } //nolint:gosec // reason: action strings are internal constants like cell: and sheet:, never user HTML

// Editor is the active cell editor: the text before and after the caret.
type Editor struct {
	X      int
	Y      int
	W      int
	H      int
	Before string
	After  string
}

// Dialog is the CSV import or export overlay.
type Dialog struct {
	Title     string
	Prompt    string
	Field     string
	FieldType string
	FieldVal  string
	Value     string
	Rows      []string
	Note      string
	Warnings  []string
	X         int
	Y         int
	W         int
	H         int
	Buttons   []Box
}
