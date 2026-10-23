package console

import "github.com/chinmay-sawant/ownframe"

// View is the data the shell template prints.
type View struct {
	Screen     string
	Notice     string
	Operator   string
	Name       string
	PIN        ownframe.Secret
	Ready      bool
	Error      string
	Tickets    []Ticket
	ListTop    int
	ListBottom int
	Empty      bool
	Current    Ticket
	Note       string
	Compact    bool
	Night      bool
	Site       string
	Filter     string
	OpenN      int
	HoldN      int
	ClosedN    int
	Nav        []NavItem
}

// Ticket is one yard incident row.
type Ticket struct {
	ID     string
	Title  string
	Bay    string
	Status string
	Note   string
	Action string
}

// NavItem is one rail link.
type NavItem struct {
	Label  string
	Action string
	On     bool
}
