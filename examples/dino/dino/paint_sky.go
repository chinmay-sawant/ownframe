package dino

import "github.com/chinmay-sawant/ownframe"

// paintClouds places the two clouds from their drifting positions.
func (a *App) paintClouds(d *ownframe.Display) {
	for i := range a.clouds {
		c := a.clouds[i]
		base := i * opsPerCloud

		a.setFill(d, a.parts.clouds[base], rect{c.x, c.y + 8, 26, 10}, true)
		a.setFill(d, a.parts.clouds[base+1], rect{c.x + 8, c.y, 34, 14}, true)
		a.setFill(d, a.parts.clouds[base+2], rect{c.x + 30, c.y + 8, 22, 10}, true)
	}
}

// paintPebbles places the six ground marks.
func (a *App) paintPebbles(d *ownframe.Display) {
	for i := range a.pebbles {
		a.setFill(d, a.parts.pebbles[i], rect{a.pebbles[i], 242, 8, 4}, true)
	}
}
