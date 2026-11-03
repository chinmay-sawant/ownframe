package window

import (
	"time"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/hajimehoshi/ebiten/v2"
)

const (
	devPanelRefreshGap = time.Second
	devPanelFastGap    = 250 * time.Millisecond
	devPanelSlowGap    = 5 * time.Second
)

// The panel owns one texture. Metrics refresh at 1 Hz by default; input and geometry
// changes invalidate it immediately. Page outlines are drawn separately.
type devPanelCache struct {
	img   *ebiten.Image
	key   devPanelKey
	next  time.Time
	valid bool
}

type devPanelKey struct {
	width, height    int
	dockWidth        float64
	tab              devTab
	scroll           float64
	mobile, expanded bool
	picker, ops      bool
	haveOp, havePin  bool
	opPick           int
	cursorX, cursorY int
	gen              uint64
	display          *layout.Display
	selection        devElementJSON
	paused           bool
	refreshGap       time.Duration
}

func (s *shell) devPanelKey() devPanelKey {
	key := devPanelKey{
		paused: s.dev.paused, refreshGap: s.dev.refreshGap,
		width: s.screenW, height: s.screenH, dockWidth: s.devDockWidth(),
		tab: s.dev.tab, scroll: s.dev.scroll,
		mobile: s.dev.mobile, expanded: s.dev.mobileExpanded,
		picker: s.dev.picker, ops: s.dev.ops,
		haveOp: s.dev.haveOp, havePin: s.dev.havePin, opPick: s.dev.opPick,
	}
	if !s.dev.mobile && s.dev.tab != devTabFrame {
		key.gen, key.display = s.devGeneration(), s.display
	}
	if devInRect(s.dev.panel, float64(s.cursorX), float64(s.cursorY)) {
		key.cursorX, key.cursorY = s.cursorX, s.cursorY
	}
	if box, ok := s.devBox(); ok {
		key.selection = devElementJSON{
			Tag: box.Tag, ID: box.ID, Action: box.Action, Text: box.Text,
			Rect: devElementRect{X: box.X, Y: box.Y, Width: box.W, Height: box.H},
		}
	}
	return key
}

func (s *shell) refreshDevPanel(now time.Time) bool {
	c := &s.dev.panelCache
	key := s.devPanelKey()
	if c.valid && c.key == key && (c.next.IsZero() || now.Before(c.next)) {
		return false
	}
	s.devLayout()
	c.key = s.devPanelKey()
	c.next = time.Time{}
	if (s.dev.mobile || s.dev.tab == devTabFrame) && !s.dev.paused {
		gap := s.dev.refreshGap
		if gap == 0 {
			gap = devPanelRefreshGap
		}
		c.next = now.Add(gap)
	}
	c.valid = true
	return true
}

func (s *shell) drawCachedDevPanel(screen *ebiten.Image, refresh bool) {
	r := s.dev.panel
	bounds := frameRect(r.X, r.Y, r.W, r.H).Intersect(screen.Bounds())
	if bounds.Empty() {
		return
	}
	c := &s.dev.panelCache
	if c.img == nil || c.img.Bounds() != bounds {
		if c.img != nil {
			c.img.Deallocate()
		}
		c.img = ebiten.NewImageWithOptions(bounds, &ebiten.NewImageOptions{Unmanaged: true})
		refresh = true
	}
	if refresh {
		c.img.Clear()
		s.drawDevPanel(c.img)
	}
	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Translate(float64(bounds.Min.X), float64(bounds.Min.Y))
	screen.DrawImage(c.img, opts)
	s.frameNote(bounds)
}

func (c *devPanelCache) close() {
	if c.img != nil {
		c.img.Deallocate()
	}
	*c = devPanelCache{}
}
