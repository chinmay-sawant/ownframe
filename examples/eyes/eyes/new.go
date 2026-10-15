package eyes

import (
	"math/rand/v2"
	"time"

	"github.com/chinmay-sawant/ownframe"
)

// New parses the eyes page and registers input and the frame tick.
func New() (*App, error) {
	p, err := ownframe.New(ownframe.Config{
		Title: "Eyes", HTML: buildHTML(),
		Width: DefaultWidth, Height: DefaultHeight,
		MinWidth: MinWidth, MinHeight: MinHeight,
		MaxWidth: MaxWidth, MaxHeight: MaxHeight})
	if err != nil {
		return nil, err
	}
	sd := uint64(time.Now().UnixNano())
	a := &App{page: p,
		rng: rand.New(rand.NewPCG(sd, sd^0x9e3779b97f4a7c15)),
		now: time.Now}
	a.st = newState(a.now)
	p.Handle(ownframe.Handlers{KeyDown: a.onKeyDown,
		Click: a.onTap, Swipe: a.onSwipe,
		Press: a.onPress, DragStart: a.onDragStart,
		DragMove: a.onDragMove, DragEnd: a.onDragEnd})
	p.SetTick(a.Tick)
	return a, nil
}

// Page returns the ownframe page Run and Serve display.
func (a *App) Page() *ownframe.Page { return a.page }
