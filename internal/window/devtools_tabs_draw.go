package window

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// devTabUnderline is the active tab underline thickness and rise.
const devTabUnderline = 2.0

// drawDevTabs paints the tab strip. The active tab gets the blue underline.
func (s *shell) drawDevTabs(screen *ebiten.Image) {
	r := s.dev.panel
	y := r.Y + devHeaderH
	vector.FillRect(screen, float32(r.X), float32(y), float32(r.W), devTabsH, devHeaderBg, false)
	vector.FillRect(screen, float32(r.X), float32(y+devTabsH-1), float32(r.W), 1, devBorder, false)

	x := r.X + devPanelPad
	top := y + (devTabsH-devFaceHeight())/devCenterDiv

	for i, name := range devTabNames {
		ink := devDim
		if devTab(i) == s.dev.tab {
			ink = devFg
		}

		devTextAt(screen, name, x, top, ink)

		w, _ := text.Measure(name, badgeFace, 0)
		if devTab(i) == s.dev.tab {
			vector.FillRect(screen, float32(x), float32(y+devTabsH-devTabUnderline), float32(w), devTabUnderline, devAccent, false)
		}

		x += w + devTabPad
	}
}
