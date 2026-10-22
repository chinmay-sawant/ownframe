// Package ipc is the in-process IPC example.
// There is no socket, no second process, and no JavaScript. Send, Listen,
// Handle, and Request are calls inside this one process. The screen registers
// two listeners on demo.log and one handler on demo.double, then shows what
// each call does. ownframe opens the window; this package does not.
package ipc

import (
	_ "embed"

	"github.com/chinmay-sawant/ownframe"
)

//go:embed ipc.html
var ipcHTML string

// DefaultWidth and DefaultHeight are the size of a newly opened window.
const (
	DefaultWidth  = 640
	DefaultHeight = 700
	// maxEvents caps the message tape at the newest entries.
	maxEvents = 6
	// doubleFactor is the demo.double reply multiplier.
	doubleFactor = 2
)

// View is the state the template prints.
type View struct {
	Status   string   // outcome of the last click
	Bad      bool     // that outcome was an error
	Received string   // payloads the demo.log listeners saw
	Sends    int      // sends the counting listener saw
	Events   []string // message tape, newest first
	Live     bool     // registrations are active
}

// App is the IPC screen. The stop fields cancel the registrations from New.
type App struct {
	page       *ownframe.Page
	view       View
	stopLog    func()
	stopCount  func()
	stopDouble func()
}
