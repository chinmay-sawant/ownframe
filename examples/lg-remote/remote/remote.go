// Package remote is the LG webOS remote page.
// The desktop build talks to the TV over Wi-Fi. The Android build can also
// send the same keys as a Bluetooth HID keyboard. The theme starts dark.
package remote

import (
	"github.com/chinmay-sawant/ownframe"
	"time"
)

const (
	// DefaultWidth and DefaultHeight fit a phone remote in a desktop window.
	DefaultWidth  = 420
	DefaultHeight = 800
)

// App is the remote screen.
type App struct {
	page           *ownframe.Page
	view           View
	phone          bool
	dark           bool
	typed          string
	focusNext      string
	focusTries     int
	host           string
	link           linker
	async          bool
	loaded         bool
	wantSearch     bool
	btSeen         uint64
	notes          chan update
	jobs           jobQueue
	wake           wakeState
	access         accessState
	pressedUntil   time.Time
	powerIntent    powerIntent
	motion         padMotion
	dragTarget     string
	sliderPosition float64
	repeat         repeatState
	nativeHeld     string
	touchDirty     bool
}

type update struct {
	status   string
	title    string
	host     string
	setPower bool
	powerOn  bool
	wakeHost string
	powerSeq uint64
}

type linker interface {
	Exec(host, spec string) (string, error)
	Scan() (string, error)
	Wake() (string, error)
	SavedHost() string
	SavedModel() string
}

// Page is the ownframe page.
func (a *App) Page() *ownframe.Page { return a.page }

// Status is the line under the title.
func (a *App) Status() string { return a.view.Status }

// Mode is "Wi-Fi" or "Bluetooth".
func (a *App) Mode() string { return a.view.Mode }

// ThemeLabel is the theme button caption.
func (a *App) ThemeLabel() string { return a.view.ThemeLabel }
