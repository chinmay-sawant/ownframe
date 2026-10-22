package window

import (
	"fmt"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	devMobileHeader     = 48.0
	devMobileSafeBottom = 32.0
	devMobileFraction   = 0.4
)

func devMobileRect(w, h int, expanded bool) devRect {
	available := math.Max(0, float64(h)-devMobileSafeBottom)
	height := math.Min(available, devMobileHeader)
	if expanded {
		height = math.Min(available, math.Max(devMobileHeader, float64(h)*devMobileFraction))
	}
	return devRect{Y: available - height, W: math.Max(0, float64(w)), H: height}
}

func (s *shell) devMobileLayout() {
	s.dev.panel = devMobileRect(s.screenW, s.screenH, s.dev.mobileExpanded)
	r := s.dev.panel
	s.dev.content = devRect{X: devPanelPad, Y: r.Y + devMobileHeader, W: math.Max(0, r.W-2*devPanelPad), H: math.Max(0, r.H-devMobileHeader)}
	s.dev.rows = nil
	if s.dev.mobileExpanded {
		s.dev.rows = s.devMobileRows(s.dev.content.W)
	}
	s.devClampScroll()
	closeWidth := math.Min(r.W, devMobileHeader)
	headerHeight := math.Min(r.H, devMobileHeader)
	s.dev.hits = []devHit{
		{rect: devRect{X: r.X + r.W - closeWidth, Y: r.Y, W: closeWidth, H: headerHeight}, act: devActClose},
		{rect: devRect{X: r.X, Y: r.Y, W: r.W - closeWidth, H: headerHeight}, act: devActMobileExpand},
	}
}

func (s *shell) devMobileRows(width float64) []devRow {
	rows := []devRow{}
	header := func(label string) { rows = append(rows, devRow{line: devText(label, devDim)}) }
	value := func(label string, n any) { rows = append(rows, devFrameRow(width, label, fmt.Sprint(n), devNumInk)) }
	value("FPS", fmt.Sprintf("%.1f", ebiten.ActualFPS()))
	value("TPS", fmt.Sprintf("%.1f", ebiten.ActualTPS()))
	s.addPerfRows(&rows, width, header, value)
	s.addMemoryRows(&rows, width, header, value)
	header("Pipeline")
	value("Redraws", s.dev.stats.Redraws)
	value("Layouts", s.dev.stats.Layouts)
	value("Repaints", s.dev.stats.Repaints)
	s.addPipelinePerfRows(&rows, width, header, value, s.dev.stats)
	return rows
}

func (s *shell) drawDevMobile(screen *ebiten.Image) {
	r := s.dev.panel
	vector.FillRect(screen, float32(r.X), float32(r.Y), float32(r.W), float32(r.H), devPanelBg, false)
	vector.FillRect(screen, float32(r.X), float32(r.Y), float32(r.W), float32(math.Min(r.H, devMobileHeader)), devHeaderBg, false)
	label := "Metrics  |  tap to expand"
	if s.dev.mobileExpanded {
		label = "Metrics  |  tap to collapse"
	}
	devTextAt(screen, label, r.X+devPanelPad, r.Y+(devMobileHeader-devFaceHeight())/devCenterDiv, devFg)
	devTextAt(screen, "X", r.X+r.W-devMobileHeader/2-devTextWidth("X")/2, r.Y+(devMobileHeader-devFaceHeight())/devCenterDiv, devFg)
	if s.dev.mobileExpanded {
		s.drawDevContent(screen)
		s.drawDevScrollbar(screen)
	}
	s.frameNote(frameRect(r.X, r.Y, r.W, r.H))
}
