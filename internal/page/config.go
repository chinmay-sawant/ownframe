package page

import (
	"io"
	"strings"
	"time"

	"github.com/chinmay-sawant/blinkless/css"
	"github.com/chinmay-sawant/ownframe/internal/host"
	"github.com/chinmay-sawant/ownframe/internal/page/draw"
	"github.com/chinmay-sawant/ownframe/internal/render"
)

// Config is the template, the optional theme, and the frame size for a new
// page. Theme is an extra stylesheet applied after the template's own styles;
// empty means no theme.
// Width and Height are the first frame, in CSS pixels.
// MinWidth and MinHeight are the smallest frame. Zero means 1.
// MaxWidth and MaxHeight ask the host to cap the window. Zero means no cap.
// They do not cap the picture, so the layout follows the window.
type Config struct {
	Title string
	HTML  string
	// File reads the page source from disk at New and watches it. Setting
	// File and HTML together is an error.
	File string
	// StaticHTML renders HTML or File without executing Go template actions.
	// Reloads of File use the same mode.
	StaticHTML bool
	// Render writes trusted author HTML directly for each redraw, bypassing
	// html/template reflection. Setting Render with HTML or File is an error.
	Render func(w io.Writer, data any) error
	Theme  string
	// ThemeFile reads the theme stylesheet from disk and watches it.
	ThemeFile string
	// DisableHotReload turns the file watch off. It is ignored without File
	// or ThemeFile.
	DisableHotReload bool
	// RetainState keeps the layout state a single element patch needs. It is
	// off by default, so a page that never patches pays nothing. With it on,
	// Page.PatchElement can re-lay out one element.
	RetainState bool
	// DevTools starts the window overlay on and enables performance sampling
	// while it is open.
	DevTools bool
	// Perf records pipeline timing, dirty counts, and allocs in Stats. It
	// is off by default unless DevTools is enabled; SetPerf opts in separately.
	Perf bool
	// LockView scales the page to the window and turns off scroll and zoom.
	LockView bool
	// Scrollbar selects thumb visibility. Auto shows thumbs except on
	// Android and iOS. Visible and Hidden override that platform default.
	Scrollbar host.ScrollbarMode
	// DisableTouchZoom keeps two-finger touch gestures from zooming the page.
	// Wheel and keyboard zoom remain available unless LockView,
	// DisableWheelZoom, or DisableKeyZoom is also set.
	DisableTouchZoom bool
	// DisableWheelZoom keeps Ctrl+wheel and Command+wheel from scaling the
	// page. The wheel scrolls. Keyboard zoom is unchanged.
	DisableWheelZoom bool
	// DisableKeyZoom keeps Ctrl+=, Ctrl++, Ctrl+-, and Ctrl+0 from scaling
	// the page. The window still consumes those chords. Wheel zoom is
	// unchanged.
	DisableKeyZoom bool
	// DeepLinkSchemes lists URL schemes accepted by DispatchDeepLink.
	DeepLinkSchemes []string
	Width           int
	Height          int
	MinWidth        int
	MinHeight       int
	MaxWidth        int
	MaxHeight       int
}

// renderState is the input, theme, and image state Redraw and PNG share.
func renderState(p *Page) render.State {
	focus := ""
	if st := formState(p); st != nil {
		focus = st.FocusID
	}

	return draw.RenderState(p.HoverID(), activeID(p), focus, p.theme, p.imageResolver, p.retainState)
}

// allocSampleGap bounds ReadMemStats on a Perf redraw to about twice a second.
// Stage timing still records every redraw.
const allocSampleGap = 500 * time.Millisecond

// SetPerf turns pipeline timing, dirty counters, and alloc tracking on or
// off independently of DevTools. Sampling stays on while DevTools is open.
// Each call clears the allocator sample clock, so the next tracked redraw
// samples immediately.
func (p *Page) SetPerf(on bool) {
	defer notifyFrame(p)
	p.perfOptIn = on
	setPerfEnabled(p, on || p.devtools)
}

func setPerfEnabled(p *Page, on bool) {
	p.perf = on
	p.allocAt = time.Time{}
}

// allocDue reports whether this redraw should read the allocator.
// A zero clock is due. A later call inside allocSampleGap is not and does
// not move the clock. A call at the gap is due.
func allocDue(p *Page, now time.Time) bool {
	if !p.allocAt.IsZero() && now.Sub(p.allocAt) < allocSampleGap {
		return false
	}
	p.allocAt = now

	return true
}

// Perf reports whether pipeline timing and dirty counters are recorded.
func (p *Page) Perf() bool { return p.perf }

// clearPerf zeroes the opt-in Stats fields after an untracked Redraw, so a
// page that opts out never shows stale developer numbers.
func clearPerf(p *Page) {
	p.stats.lastTemplate = 0
	p.stats.styleTime = 0
	p.stats.layoutTime = 0
	p.stats.paintTime = 0
	p.stats.otherTime = 0
	p.stats.dirtyOps = 0
	p.stats.dirtyRegions = 0
	p.stats.dirtyArea = 0
	p.stats.changedOps = 0
	p.stats.allocFrame = 0
}

// pageStats counts the stages of every Redraw. The devtools overlay reads
// them through Stats.
type pageStats struct {
	fullLayouts      uint64
	hoverPatches     uint64
	hoverFallbacks   uint64
	hoverPatchErrors uint64

	redraws    uint64
	parses     uint64
	cascades   uint64
	layouts    uint64
	repaints   uint64
	reloads    uint64
	reloadErr  string
	lastRedraw time.Duration
	lastDraw   time.Duration
	frameDraw  time.Duration
	// lastTemplate, styleTime, layoutTime, and paintTime split the
	// last Redraw into stages. dirtyOps, dirtyRegions, and changedOps
	// describe its dirty region. allocFrame is the bytes allocated during
	// the last sampled Redraw. Sampling runs about twice a second.
	// Unsampled redraws keep the previous value.
	lastTemplate time.Duration
	styleTime    time.Duration
	layoutTime   time.Duration
	paintTime    time.Duration
	otherTime    time.Duration
	dirtyOps     int
	dirtyRegions int
	dirtyArea    int
	changedOps   int
	allocFrame   uint64
}

// DevTools reports whether the window devtools overlay starts on.
func (p *Page) DevTools() bool {
	return p.devtools
}

// SetDevTools turns the window devtools overlay on or off.
func (p *Page) SetDevTools(on bool) {
	defer notifyFrame(p)
	p.devtools = on
	setPerfEnabled(p, on || p.perfOptIn)
}

// SetDrawTime records how long the window spent drawing the last frame, so
// Stats reports it as LastDraw. The window calls this through its own
// interface; host.Inspector keeps its shape.
func (p *Page) SetDrawTime(d time.Duration) {
	p.stats.frameDraw = d
}

var _ host.Inspector = (*Page)(nil)

// SetTheme stores an extra stylesheet the next Redraw applies after the
// page's own style elements. Any property the engine supports restyles the
// page, and a theme rule wins a tie with a template rule. An empty source
// removes the theme. SetTheme does not draw. source is trusted application
// CSS; an oversized source is rejected.
func (p *Page) SetTheme(source string) error {
	sheet, err := parseTheme(source)
	if err != nil {
		return err
	}

	p.theme = sheet
	p.themeSrc = source
	invalidateCache(p)
	markFull(p)

	return nil
}

// parseTheme parses theme CSS. Empty or blank source means no theme.
func parseTheme(source string) (*css.Sheet, error) {
	if len(source) > MaxSourceSize {
		return nil, sourceTooLarge("theme", int64(len(source)))
	}

	if strings.TrimSpace(source) == "" {
		return nil, nil
	}

	return css.Parse(source)
}
