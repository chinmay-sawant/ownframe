package window

import (
	"bytes"
	"image/color"
	"math"
	"runtime"
	"sync/atomic"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font/gofont/gomedium"
)

const titlebarHeight = 36
const titlebarButtonWidth = 46

// Caption button numbers, left to right, as the switch statements see them.
// titlebarButtonCount is how many caption buttons the titlebar draws.
const (
	buttonMaximize      = 2
	buttonClose         = 3
	titlebarButtonCount = 3
	// halfDivisor halves a size to find its center.
	halfDivisor = 2
)

// Title layout: font size in points, the smallest header height a caller can
// ask for, the padding around the title, and the room kept free at its end.
// midpointDivisor halves the search range when fitting the title to its width.
const (
	headerFontSize  = 13
	minHeaderHeight = 24
	midpointDivisor = 2
	titleInsetX     = 14
	titleEndReserve = 28
)

var contentTop atomic.Int32 //nolint:gochecknoglobals // ContentCursorPosition is a package-level func with no receiver, so the offset it reads must be package state

// storeContentTop records the titlebar inset that page content starts below.
// Insets are never negative and are far below MaxInt32, so the clamp only
// makes the int32 conversion provably safe.
func storeContentTop(inset int) {
	contentTop.Store(int32(min(max(inset, 0), math.MaxInt32)))
}

// ContentCursorPosition returns the mouse position in page coordinates.
func ContentCursorPosition() (int, int) {
	x, y := ebiten.CursorPosition()
	return x, y - int(contentTop.Load())
}

type titlebarState struct {
	style       headerStyle
	enabled     bool
	starting    bool
	fixed       bool
	pressed     int
	down        bool
	canvas      *ebiten.Image
	title       string
	fittedTitle string
	titleWidth  int
}

type headerStyle struct {
	title                                     string
	height                                    int
	font                                      text.Face
	background, foreground, hover, closeHover color.Color
}

func newHeaderFont() (text.Face, error) {
	source, err := text.NewGoTextFaceSource(bytes.NewReader(gomedium.TTF))
	if err != nil {
		return nil, err
	}
	return &text.GoTextFace{Source: source, Size: headerFontSize}, nil
}

func fitHeaderTitle(title string, face text.Face, width int) string {
	if width <= 0 {
		return ""
	}
	if w, _ := text.Measure(title, face, 0); w <= float64(width) {
		return title
	}
	const ellipsis = "…"
	if w, _ := text.Measure(ellipsis, face, 0); w > float64(width) {
		return ""
	}
	runes := []rune(title)
	lo, hi := 0, len(runes)
	for lo < hi {
		mid := (lo + hi + 1) / midpointDivisor
		w, _ := text.Measure(string(runes[:mid])+ellipsis, face, 0)
		if w <= float64(width) {
			lo = mid
		} else {
			hi = mid - 1
		}
	}
	return string(runes[:lo]) + ellipsis
}

func headerHeight(o Options) int {
	if o.HeaderHeight <= 0 {
		return titlebarHeight
	}
	return max(minHeaderHeight, o.HeaderHeight)
}

func (s *shell) placeTitlebarWindow() {
	if !s.titlebar.starting {
		return
	}
	rect, ok := nativeWindowGeometry()
	if !ok {
		return
	}
	mon, ok := monitorForWindow(rect.x, rect.y, rect.w, rect.h, monitorLayout())
	if !ok {
		return
	}
	chromeMu.Lock()
	area, ok := wslgHostWorkareaLocked(mon)
	chromeMu.Unlock()
	if !ok {
		return
	}
	x := max(area.x, min(rect.x, area.x+area.w-rect.w))
	y := max(area.y, min(rect.y, area.y+area.h-rect.h))
	if x != rect.x || y != rect.y {
		setWindowRect(x, y, rect.w, rect.h)
		windowTracef("startup fitting default titlebar window monitor=%q target=(%d,%d %dx%d)", mon.name, x, y, rect.w, rect.h)
	}
	s.titlebar.starting = false
}

func usesTitlebar(o Options) bool {
	return runtime.GOOS != "js" && runtime.GOOS != "android" && runtime.GOOS != "ios" &&
		!o.NativeHeader && !o.CustomHeader && !o.Borderless && !o.Transparent &&
		!o.MousePassthrough && o.Interactive == nil
}

func (s *shell) titlebarInset() int {
	if !s.titlebar.enabled || s.fullscreenNow() {
		return 0
	}
	if s.titlebar.style.height > 0 {
		return s.titlebar.style.height
	}
	return titlebarHeight
}

func titlebarButton(x, width int) int {
	if x < max(0, width-3*titlebarButtonWidth) || x >= width {
		return 0
	}
	return (x-(width-3*titlebarButtonWidth))/titlebarButtonWidth + 1
}

func (s *shell) titlebarPointer(x, y int, down bool) bool {
	if s.titlebarInset() == 0 {
		return false
	}
	width, _ := ebiten.WindowSize()
	hit := 0
	if y < 0 && y >= -s.titlebarInset() {
		hit = titlebarButton(x, width)
	}
	pressed := s.titlebar.pressed
	if down && !s.titlebar.down {
		s.titlebar.pressed = hit
	}
	s.titlebar.down = down
	if !down && pressed != 0 {
		s.titlebar.pressed = 0
		if hit == pressed {
			switch hit {
			case 1:
				Minimize()
			case buttonMaximize:
				if !s.titlebar.fixed {
					ToggleMaximize()
				}
			case buttonClose:
				RequestClose()
			}
		}
		return true
	}
	return hit != 0 || s.titlebar.pressed != 0
}

func (s *shell) titlebarCanvas(screen *ebiten.Image) (*ebiten.Image, func()) {
	inset := s.titlebarInset()
	if inset == 0 {
		return screen, func() {}
	}
	w, h := screen.Bounds().Dx(), max(1, screen.Bounds().Dy()-inset)
	canvas := s.titlebar.canvas
	if canvas == nil || canvas.Bounds().Dx() != w || canvas.Bounds().Dy() != h {
		if canvas != nil {
			canvas.Deallocate()
		}
		canvas = ebiten.NewImage(w, h)
		s.titlebar.canvas = canvas
	}
	return canvas, func() {
		var op ebiten.DrawImageOptions
		op.GeoM.Translate(0, float64(inset))
		screen.DrawImage(canvas, &op)
		s.drawTitlebar(screen)
	}
}

func (s *shell) drawTitlebar(screen *ebiten.Image) {
	width := screen.Bounds().Dx()
	// The caption colors are RGBA channel tables; one const per channel would hide the color.
	var bg color.Color = color.NRGBA{R: 235, G: 231, B: 218, A: 255} //nolint:mnd // RGBA palette table
	var ink color.Color = color.NRGBA{R: 31, G: 41, B: 55, A: 255}   //nolint:mnd // RGBA palette table
	page := s.pageBackground()
	r, g, b, _ := page.RGBA()
	if r+g+b < 3*0x8000 {
		bg = color.NRGBA{R: 35, G: 38, B: 44, A: 255}     //nolint:mnd // RGBA palette table
		ink = color.NRGBA{R: 235, G: 237, B: 240, A: 255} //nolint:mnd // RGBA palette table
	}
	style := s.titlebar.style
	if style.background != nil {
		bg = style.background
	}
	if style.foreground != nil {
		ink = style.foreground
	}
	height := float32(s.titlebarInset())
	vector.FillRect(screen, 0, 0, float32(width), height, bg, false)
	x, y := s.cursorPosition()
	for button := 1; button <= 3; button++ {
		left := width - (titlebarButtonCount+1-button)*titlebarButtonWidth
		if y < 0 && y >= -s.titlebarInset() && titlebarButton(x, width) == button {
			var hover color.Color = color.NRGBA{R: 128, G: 128, B: 128, A: 90} //nolint:mnd // RGBA palette table
			if style.hover != nil {
				hover = style.hover
			}
			if button == buttonClose {
				hover = color.NRGBA{R: 210, G: 40, B: 40, A: 255} //nolint:mnd // RGBA palette table
				if style.closeHover != nil {
					hover = style.closeHover
				}
			}
			vector.FillRect(screen, float32(left), 0, titlebarButtonWidth, height, hover, false)
		}
		cx := float32(left) + titlebarButtonWidth/halfDivisor
		cy := height / halfDivisor
		switch button {
		case 1:
			vector.StrokeLine(screen, cx-5, cy+3, cx+5, cy+3, 1, ink, false) //nolint:mnd // caption glyph pixel offsets
		case buttonMaximize:
			if Maximized() {
				vector.StrokeRect(screen, cx-3, cy-5, 8, 8, 1, ink, false) //nolint:mnd // caption glyph pixel offsets
			}
			vector.StrokeRect(screen, cx-5, cy-3, 8, 8, 1, ink, false) //nolint:mnd // caption glyph pixel offsets
		case buttonClose:
			vector.StrokeLine(screen, cx-4, cy-4, cx+4, cy+4, 1, ink, false) //nolint:mnd // caption glyph pixel offsets
			vector.StrokeLine(screen, cx-4, cy+4, cx+4, cy-4, 1, ink, false) //nolint:mnd // caption glyph pixel offsets
		}
	}
	title := s.app.Title()
	if style.title != "" {
		title = style.title
	}
	available := max(0, width-3*titlebarButtonWidth-titleEndReserve)
	if title != s.titlebar.title || available != s.titlebar.titleWidth {
		s.titlebar.title = title
		s.titlebar.titleWidth = available
		s.titlebar.fittedTitle = fitHeaderTitle(title, style.font, available)
	}
	metrics := style.font.Metrics()
	var op text.DrawOptions
	op.GeoM.Translate(titleInsetX, (float64(height)-metrics.HAscent-metrics.HDescent)/halfDivisor)
	op.ColorScale.ScaleWithColor(ink)
	text.Draw(screen, s.titlebar.fittedTitle, style.font, &op)
}
