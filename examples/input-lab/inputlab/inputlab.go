// Package inputlab is the combined form and input example.
// One window covers the old forms, controls, bind, bind-hooks, editing,
// clipboard, input, login, and states demos: plain controls, two-way
// binding, a locked bound field, edit undo, OS clipboard, scroll list,
// sign-in, and CSS pseudo-classes. ownframe opens the window, not this package.
package inputlab

import _ "embed"

//go:embed inputlab.html
var inputlabHTML string

const (
	DefaultWidth  = 640
	DefaultHeight = 900
	MinWidth      = 320
	MinHeight     = 400
	undoLimit     = 64
)

// View is the data the template prints. Field values live in the
// page; bound fields mirror into B/L pairs through data-bind.
type View struct {
	BName     string
	BEmail    string
	BAgree    bool
	BPlan     string
	BColor    string
	BStatus   string
	LName     string
	LEmail    string
	LAgree    bool
	LPlan     string
	LStatus   string
	Status    string
	Stamp     int
	Ready     bool
	LoginOk   string
	LoginEr   string
	MLReadout string
}
