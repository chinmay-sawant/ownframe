package flappy

import "github.com/chinmay-sawant/ownframe"

// paintBird places the six bird fills for the current pose.
func (a *App) paintBird(d *ownframe.Display) {
	pose := a.game.birdRects()

	for i := range birdCount {
		a.setFill(d, a.parts.bird[i], pose[i], true)
	}
}
