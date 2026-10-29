package window

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// drawDevPanel paints the dock: the background, the header, the tab strip,
// the content, the scrollbar, and the footer hint.
func (s *shell) drawDevPanel(screen *ebiten.Image) {
	if s.dev.mobile {
		s.drawDevMobile(screen)
		return
	}
	r := s.dev.panel
	if r.W <= 0 || r.H <= 0 {
		return
	}

	vector.FillRect(screen, float32(r.X), float32(r.Y), float32(r.W), float32(r.H), devPanelBg, false)
	vector.FillRect(screen, float32(r.X), float32(r.Y), 1, float32(r.H), devBorder, false)
	s.frameNote(frameRect(r.X, r.Y, r.W, r.H))

	s.drawDevHeader(screen)
	s.drawDevTabs(screen)
	s.drawDevContent(screen)
	s.drawDevScrollbar(screen)
	s.drawDevFooter(screen)
}

// drawDevHeader paints the title bar, the picker toggle, and the close key hint.
func (s *shell) drawDevHeader(screen *ebiten.Image) {
	r := s.dev.panel
	vector.FillRect(screen, float32(r.X), float32(r.Y), float32(r.W), devHeaderH, devHeaderBg, false)

	top := r.Y + (devHeaderH-devFaceHeight())/devCenterDiv
	devTextAt(screen, "Developer Tools", r.X+devPanelPad, top, devFg)

	label := "Picker"
	ink := devDim
	if s.devPickerOn() {
		ink = devAccent
	}
	cls, px := devHeaderRight(r)
	devTextAt(screen, label, px, top, ink)

	devTextAt(screen, "F12", cls.X, top, devDim)
	xX := r.X + r.W - devPanelPad - devTextWidth("X")
	devTextAt(screen, "X", xX, top, devDim)
}

// drawDevFooter paints the hint line at the bottom of the dock.
func (s *shell) drawDevFooter(screen *ebiten.Image) {
	r := s.dev.panel
	y := r.Y + r.H - devFooterH
	vector.FillRect(screen, float32(r.X), float32(y), float32(r.W), 1, devBorder, false)
	devTextAt(screen, devTabHint(s.dev.tab), r.X+devPanelPad, y+(devFooterH-devFaceHeight())/devCenterDiv, devDim)
}
