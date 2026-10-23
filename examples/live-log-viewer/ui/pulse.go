package ui

import (
	"math"
	"time"

	"github.com/chinmay-sawant/ownframe"
)

// pulse colors match --unread in the light and dark sheets.
var (
	pulseLight = [3]float64{0x25 / 255.0, 0x63 / 255.0, 0xeb / 255.0}
	pulseDark  = [3]float64{0x6e / 255.0, 0xa8 / 255.0, 0xfe / 255.0}
)

// pulse is the paint-only half of the tick: it breathes the live dot by
// changing one retained operation's alpha, with no Redraw and no layout.
// A Redraw or a resize replaces the display list, so the pointer is
// rebound whenever the page generation changes; a bitmap-fallback page
// has no display list and the dot simply stays static.
func (a *App) pulse(now time.Time) {
	if !a.follow.Live() {
		return
	}

	d := a.page.Probe().Display()
	if d == nil {
		return
	}

	if a.pulseGen != a.page.Probe().Generation() {
		a.pulseGen = a.page.Probe().Generation()
		a.pulseDot = nil

		if box, ok := a.boxByID("pulse"); ok {
			color := pulseLight
			if a.dark {
				color = pulseDark
			}

			a.pulseDot = ownframe.Fill(d, box, color)
		}
	}

	if a.pulseDot == nil {
		return
	}

	phase := math.Sin(now.Sub(a.started).Seconds() * uiPulseFreq)
	a.pulseDot.Alpha = uiPulseBase + uiPulseAmp*(uiPulseHalf+uiPulseHalf*phase)
}

// boxByID returns the hit-test box with id.
func (a *App) boxByID(id string) (ownframe.Box, bool) {
	for _, b := range a.page.Boxes() {
		if b.ID == id {
			return b, true
		}
	}

	return ownframe.Box{}, false
}
