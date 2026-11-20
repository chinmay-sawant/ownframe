package inputlab

import "github.com/chinmay-sawant/ownframe"

// App is the combined input screen.
type App struct {
	page        *ownframe.Page
	view        View
	lastClip    string
	selectedAll bool
	lastEdit    string
	eUndo       []string
	eRedo       []string
	cUndo       []csnap
	cRedo       []csnap
	lUndo       []lsnap
	lRedo       []lsnap
}

// csnap saves the two clipboard fields.
type csnap struct {
	Left  string
	Right string
}

// lsnap saves the two sign-in fields.
type lsnap struct {
	Email string
	Pass  string
}
