// Package pet implements a cyberpunk virtual companion screen.
// It manages pet vitals, animated face displays, interactions,
// and state transitions using ownframe.
package pet

import (
	"math/rand/v2"
	"sync"
	"time"

	"github.com/chinmay-sawant/ownframe"
)

const (
	DefaultWidth  = 520
	DefaultHeight = 640
	MinWidth      = 420
	MinHeight     = 500
	MaxWidth      = 800
	MaxHeight     = 900
)

// App is the cyber pet host and controller.
type App struct {
	page         *ownframe.Page
	state        State
	view         View
	now          func() time.Time
	rng          *rand.Rand
	last         time.Time
	lastPetDrag  time.Time
	draggingFace bool
	pm           parts
	bnd          uint64
	mu           sync.Mutex
}

// New creates and initializes the cyber pet application.
func New() (*App, error) {
	p, err := ownframe.New(ownframe.Config{
		Title:     "Cyber Pet",
		HTML:      buildHTML(),
		Width:     DefaultWidth,
		Height:    DefaultHeight,
		MinWidth:  MinWidth,
		MinHeight: MinHeight,
		MaxWidth:  MaxWidth,
		MaxHeight: MaxHeight,
	})
	if err != nil {
		return nil, err
	}

	now := time.Now
	sd := uint64(now().UnixNano())
	a := &App{
		page:  p,
		now:   now,
		rng:   rand.New(rand.NewPCG(sd, sd^0xa0761d6478bd642f)),
		state: newState(now),
	}

	a.refreshView(now())

	p.Handle(ownframe.Handlers{
		Click:     a.onClick,
		KeyDown:   a.onKeyDown,
		DragStart: a.onDragStart,
		DragMove:  a.onDragMove,
		DragEnd:   a.onDragEnd,
	})
	p.SetTick(a.Tick)

	return a, nil
}

// Page returns the underlying ownframe page.
func (a *App) Page() *ownframe.Page { return a.page }

// State returns a copy of current pet vitals.
func (a *App) State() State {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.state
}
