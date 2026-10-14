package window

const (
	devPanelPad = 8.0
	devHeaderH  = 26.0
	devTabsH    = 28.0
	devFooterH  = 22.0
)

// devLayout places the dock, builds the active tab's rows, and records this
// frame's clickable regions.
func (s *shell) devLayout() {
	s.dev.panel = devDockRect(s.screenW, s.screenH, s.devDockWidth())
	s.dev.content = devRect{
		X: s.dev.panel.X + devPanelPad,
		Y: s.dev.panel.Y + devHeaderH + devTabsH + devPanelPad,
		W: s.dev.panel.W - 2*devPanelPad,
		H: s.dev.panel.H - devHeaderH - devTabsH - devFooterH - 2*devPanelPad,
	}

	s.dev.rows = s.devTabRows()
	s.devClampScroll()
	s.devHits()
}

// devHits records the tab strip and every actionable content row. Rows are
// placed at the scroll offset the last frame drew, so the pointer reads the
// same layout it sees.
func (s *shell) devHits() {
	lineH := devLineHeight()
	s.dev.hits = s.dev.hits[:0]

	x := s.dev.panel.X + devPanelPad

	s.devPickerHit()

	for i, name := range devTabNames {
		w := devTextWidth(name)
		s.dev.hits = append(s.dev.hits, devHit{
			rect: devRect{X: x, Y: s.dev.panel.Y + devHeaderH, W: w + 20, H: devTabsH},
			act:  devActTab,
			arg:  i,
		})
		x += w + 20
	}

	y := s.dev.content.Y - s.dev.scroll

	for _, row := range s.dev.rows {
		rect := devRect{X: s.dev.panel.X, Y: y, W: s.dev.panel.W, H: lineH}
		y += lineH

		if row.act == devActNone || rect.Y+lineH < s.dev.content.Y || rect.Y > s.dev.content.Y+s.dev.content.H {
			continue
		}

		s.dev.hits = append(s.dev.hits, devHit{rect: rect, act: row.act, arg: row.arg, key: row.key})
	}
}

// devTabRows builds the content of the active tab.
func (s *shell) devTabRows() []devRow {
	switch s.dev.tab {
	case devTabFrame:
		return s.devFrameRows(s.dev.content.W)
	case devTabOps:
		return s.devOpsRows()
	default:
		return s.devElementRows()
	}
}
