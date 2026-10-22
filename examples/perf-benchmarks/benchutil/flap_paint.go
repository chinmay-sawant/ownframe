package benchutil

import (
	"context"
	"strconv"
	"time"

	"github.com/chinmay-sawant/ownframe"
)

// Tick advances the game and moves the retained ops. It never redraws: the
// score text op changes in place and a Redraw rebinds the op pointers.
func (f *Flap) Tick(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	now := time.Now()
	dt := 1.0 / flapFrameRate

	if !f.last.IsZero() {
		dt = min(max(now.Sub(f.last).Seconds(), 0), flapMaxDt)
	}

	f.last = now
	f.Step(dt)
	f.paint()

	return nil
}

// paint moves the cached ops to match the game state.
func (f *Flap) paint() {
	d := f.page.Probe().Display()
	if d == nil {
		return
	}

	if f.page.Probe().Generation() != f.bound {
		f.bind(d)
	}

	place(d, f.bird, flapBirdX, f.y, flapBirdW, flapBirdH)

	for i := range f.tops {
		if i < len(f.pipes) {
			top := f.pipes[i].gapY - flapHalfGap
			bot := f.pipes[i].gapY + flapHalfGap
			place(d, f.tops[i], f.pipes[i].x, 0, flapPipeW, top)
			place(d, f.bots[i], f.pipes[i].x, bot, flapPipeW, flapSceneH-bot)

			continue
		}

		hide(f.tops[i])
		hide(f.bots[i])
	}

	if f.text != nil && f.shown != f.score {
		f.shown = f.score
		f.text.Text = strconv.Itoa(f.score)
	}
}

// bind re-finds the scene ops after a Redraw replaced the display list.
func (f *Flap) bind(d *ownframe.Display) {
	boxes := f.page.Boxes()
	f.bird = fillIn(d, boxes, "bird")

	for i := range f.tops {
		id := "p" + strconv.Itoa(i)
		f.tops[i] = fillIn(d, boxes, id+"t")
		f.bots[i] = fillIn(d, boxes, id+"b")
	}

	f.text = textIn(d, boxes, "score")
	f.bound = f.page.Probe().Generation()
}
