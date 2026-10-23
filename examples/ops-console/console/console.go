// Package console is the five-screen yard ops sample.
// One HTML template, one View, and data-action clicks drive login, the
// ticket list, a ticket detail, settings, and a status report.
// ownframe opens the window. This package does not.
package console

import (
	_ "embed"

	"github.com/chinmay-sawant/ownframe"
)

//go:embed shell.html
var shellHTML string

const (
	// DefaultWidth and DefaultHeight are the size of a newly opened window.
	DefaultWidth  = 960
	DefaultHeight = 640

	// MinWidth and MinHeight are the smallest frame the screen will draw.
	MinWidth  = 640
	MinHeight = 480
)

const (
	screenLogin    = "login"
	screenList     = "list"
	screenDetail   = "detail"
	screenSettings = "settings"
	screenReport   = "report"
)

const (
	statusOpen   = "open"
	statusHold   = "hold"
	statusClosed = "closed"
	filterAll    = "all"
)

const (
	sitePier  = "Pier 9"
	siteAnnex = "Annex"
)

// App is the yard console screen.
type App struct {
	page     *ownframe.Page
	view     View
	tickets  []Ticket
	offsetY  int
	viewH    int
	winStart int
	winCount int
}
