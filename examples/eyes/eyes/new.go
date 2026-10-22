package eyes

import (
	"math/rand/v2"
	"time"

	"github.com/chinmay-sawant/ownframe"
)

// pcgStream xors the time seed into the PCG stream selector.
const pcgStream = 0x9e3779b97f4a7c15

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
	p.SetImage("spiral", spiralPNG)
	sd := uint64(time.Now().UnixNano())
	a := &App{page: p,
		rng: rand.New(rand.NewPCG(sd, sd^pcgStream)), //nolint:gosec // reason: eye saccades are visual demo randomness, not security-sensitive.
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
