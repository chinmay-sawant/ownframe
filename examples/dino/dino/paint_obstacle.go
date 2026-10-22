package dino

import "github.com/chinmay-sawant/ownframe"

// obstacleParts is the number of fill elements in one obstacle slot.
const partsPerSlot = len(obstacleParts)

// partBird, partSmall, and partBig are the first part index of each kind.
const (
	partBird  = 0
	partSmall = 4
	partBig   = 7
)

// paintObstacles places the active obstacles and parks the empty slots.
func (a *App) paintObstacles(d *ownframe.Display) {
	for slot := range slotMax {
		if slot < len(a.game.obstacles) {
			a.paintObstacle(d, &a.parts.slots[slot], a.game.obstacles[slot])

			continue
		}

		a.hideParts(d, &a.parts.slots[slot], 0, partsPerSlot)
	}
}

// paintObstacle places one obstacle. A part the kind does not use is
// hidden.
func (a *App) paintObstacle(d *ownframe.Display, p *[partsPerSlot]*ownframe.DisplayOp, ob obstacle) {
	top := groundY - ob.bottom - ob.h

	switch ob.kind {
	case bird:
		a.paintBird(d, p, ob, top)
		a.hideParts(d, p, partSmall, partBig)
		a.hideParts(d, p, partBig, partsPerSlot)
	case cactusBig:
		a.hideParts(d, p, partBird, partBig)
		a.setInk(d, p[partBig], rect{ob.x + 7, top, 16, 48})
		a.setInk(d, p[partBig+1], rect{ob.x, top + 10, 10, 8})
		a.setInk(d, p[partBig+2], rect{ob.x + 20, top + 20, 10, 8})
	default:
		a.hideParts(d, p, partBird, partSmall)
		a.setInk(d, p[partSmall], rect{ob.x + 5, top, 12, 32})
		a.setInk(d, p[partSmall+1], rect{ob.x, top + 6, 8, 6})
		a.setInk(d, p[partSmall+2], rect{ob.x + 14, top + 14, 8, 6})
		a.hideParts(d, p, partBig, partsPerSlot)
	}
}

// paintBird places the body, the beak, and the wing the flap cycle shows.
func (a *App) paintBird(d *ownframe.Display, p *[partsPerSlot]*ownframe.DisplayOp, ob obstacle, top float64) {
	up := int(ob.flap/flapPeriod)%flapCycle == 0

	a.setInk(d, p[0], rect{ob.x + 12, top + 10, 20, 10})
	a.setInkUp(d, p[1], rect{ob.x + 10, top, 18, 8}, up)
	a.setInkUp(d, p[2], rect{ob.x + 10, top + 20, 18, 8}, !up)
	a.setInk(d, p[3], rect{ob.x + 6, top + 12, 8, 4})
}
