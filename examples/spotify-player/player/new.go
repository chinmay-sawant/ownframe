package player

import (
	"os"
	"time"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/examples/music"
)

// App is the player screen.
type App struct {
	page           *ownframe.Page
	view           View
	navHistory     []string
	navAt          int
	base           string
	seconds        int
	lastVolume     int
	audio          *music.Engine
	audioKey       string
	clipKey        string
	logger         *os.File
	navFocusStates map[string]navFocusState
	runtime        *RuntimeBehaviorHelper
	// vpos is the virtual playback position for engines that cannot
	// advance on their own. vlast is the last tick that sampled it,
	// vlastPos the last engine position seen, vfreeze counts frozen
	// frames while playing, and vengaged selects the virtual clock.
	vpos     time.Duration
	vlast    time.Time
	vlastPos time.Duration
	vfreeze  int
	vengaged bool
}

// New parses the player template, registers its images, and starts on the
// bundled view. Search remains available from the UI.
func New() (*App, error) {
	return NewAt(itunesBase)
}

// NewAt is New with a different Search API base URL. Tests pass a local
// server so they never reach iTunes.
func NewAt(base string) (*App, error) {
	return NewWith(base, music.NewEngine(bundleResolver{}))
}

// NewWith is NewAt with an injected audio engine, so tests never open the
// OS audio device.
func NewWith(base string, engine *music.Engine) (*App, error) {
	page, err := ownframe.New(ownframe.Config{
		DevTools:    true,
		Title:       "Spotify — Player",
		HTML:        buildHTML(),
		Theme:       responsiveTheme(),
		Width:       DefaultWidth,
		Height:      DefaultHeight,
		MinWidth:    MinWidth,
		MinHeight:   MinHeight,
		RetainState: true,
	})
	if err != nil {
		return nil, err
	}

	app := &App{
		page:       page,
		view:       DefaultView(),
		navHistory: []string{"home"},
		base:       base,
		audio:      engine,
		logger:     openSpotifyLog(),
		runtime:    NewRuntimeBehaviorHelper(),
	}
	app.updateNavButtons()
	app.seconds = lengthSeconds(app.view.Now.Length)
	engine.SetVolume(float64(defaultVolume) / percentScale)
	registerImages(page)
	page.Handle(ownframe.Handlers{Click: app.onClick})
	page.SetTick(app.Tick)
	page.SetData(app.view)

	return app, nil
}
