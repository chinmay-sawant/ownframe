// Package page turns an HTML template into a picture and hit boxes.
package page

import (
	"context"
	"errors"
	"github.com/chinmay-sawant/blinkless/css"
	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/element"
	"github.com/chinmay-sawant/ownframe/internal/fetch"
	"github.com/chinmay-sawant/ownframe/internal/host"
	"github.com/chinmay-sawant/ownframe/internal/motion"
	"github.com/chinmay-sawant/ownframe/internal/page/dirty"
	"github.com/chinmay-sawant/ownframe/internal/page/draw"
	"github.com/chinmay-sawant/ownframe/internal/page/nav"
	"github.com/chinmay-sawant/ownframe/internal/page/scroll"
	"github.com/chinmay-sawant/ownframe/internal/render"
	"github.com/chinmay-sawant/ownframe/internal/state"
	"html/template"
	"image"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// Box is one element a click can land on.
// X, Y, W, H are CSS pixels, origin at the top left of the picture.
type Box = layout.Box

var (
	// ErrEmptyHTML means New was given a blank template.
	ErrEmptyHTML = errors.New("ownframe: empty html")

	// ErrBadSource means a source file could not be read at New, or a
	// Config set both HTML and File, or both Theme and ThemeFile.
	ErrBadSource = errors.New("ownframe: bad source")

	// ErrBadSize means a width or a height is unusable.
	ErrBadSize = errors.New("ownframe: bad size")

	// ErrNilPage means Run, Serve, or BindMobile was called without a page.
	ErrNilPage = errors.New("ownframe: nil page")

	// ErrPatchRefused means PatchElement could not apply the edit.
	ErrPatchRefused = errors.New("ownframe: element patch refused")

	errNilContext = fetch.ErrNilContext

	// errNoImage is the resolver answer for a src SetImage does not hold.
	errNoImage = errors.New("ownframe: no image for src")
)

// Page is one template and its picture.
type Page struct {
	title string
	tpl   *template.Template
	// staticSource marks the current nav entry as a builder document, whose
	// envelope proves it carries no template action. Such a page keeps tpl
	// nil: the redraw path returns nav.Current instead of parsing or
	// executing it, and a Load of any other source parses again.
	staticSource bool
	// dataMu guards data and the form-sync handoff below it. SetData
	// takes it to publish new data; the redraw path takes it to run the
	// template (shared) and to check, merge, and store the sync flags
	// (exclusive), so a concurrent SetData never races a Redraw.
	dataMu sync.RWMutex
	// boxCache holds the lazy kid index and motion declaration verdict for
	// the committed box list (dirty.go). It lives on the page so a
	// discarded page releases it with everything else.
	boxCache      boxCache
	data          any
	theme         *css.Sheet
	themeSrc      string
	handlers      Handlers
	gesture       pageDrag
	images        map[string][]byte
	draw          draw.Frame
	boxes         []layout.Box
	source        string
	surf          draw.RedrawSurface
	imageResolver func(string) ([]byte, error)
	// lastSyncBody is the template output syncForm last merged, and
	// formDirty forces the next Redraw to merge again even when the
	// output is unchanged (data binds may still have moved). Both are
	// read and written only while holding dataMu; outside the redraw
	// path use markFormDirty to set the flag.
	lastSyncBody string
	formDirty    bool
	// syncClean reports that the last syncForm call took the fast path:
	// same template output, nothing dirty, source byte-identical to the
	// last committed sync. The redraw wrapper then skips the source
	// rehash, since the memoized hash, length, and revision are still
	// valid. Any slow-path merge clears it. Page-loop only, like the
	// sync flags above.
	syncClean   bool
	pageGeom    // size: New inits, SetSize writes; lock: page loop
	renderTrack // generation and last frame key; lock: page loop
	// layerGen keys the compositor's retained element layers. It moves when a
	// publish or an element patch changes display content, and stays put when
	// a re-projection pass only rewrites operation matrices in place.
	layerGen    uint64
	touchCancel uint32
	tickMutated atomic.Bool
	tickActive  bool
	tickSnap    *opBuffer
	tickSlot
	nav             nav.History
	deepLinkSchemes map[string]struct{}
	ff              formFocusState
	picker          PickFunc
	// hoverPatch re-lays one element's subtree for a hover-only change.
	// Production pages use the engine patch; tests swap it per page to
	// observe which roots a hover change patches.
	hoverPatch       hoverPatchFunc
	cache            *render.Cache
	stats            pageStats
	devtools         bool
	perf             bool
	perfOptIn        bool
	motion           *motion.Runtime
	motionIndex      *motion.Index
	animationsOff    bool
	transitionsOff   bool
	triggerPress     triggerPressState
	allocAt          time.Time // last allocator sample; zero means the next Perf redraw is due
	retainState      bool
	restampRedraw    bool
	layoutAdopting   bool
	lockView         bool
	mobileViewport   bool
	allowTouchZoom   bool
	disableWheelZoom bool
	disableKeyZoom   bool
	watch            *watchState
	dirty            dirty.Tracker
	lastPatch        DisplayPatch
	// painted is the display the host last took damage for. A redraw that
	// finds damage already marked full can diff this display against the new
	// one, so a small change made after a full mark keeps its own region
	// instead of repainting the page.
	painted           *Display
	stickyViewport    *scrollViewport
	hasViewportSticky bool
	// commitRoots scopes the current commit to the patched element ids: the
	// hover patch roots, one PatchElement id, or a PatchElements id list. A
	// nil slice means a full layout, whose change can reach any box.
	// captureTransitions consumes and clears it, so a scoped skip only ever
	// applies to the commit that produced the scope.
	commitRoots []string
	// fullGen is the generation whose commit last marked the whole frame
	// dirty. A mark from an earlier generation is covered by the display
	// diff against painted; one from the current generation may describe an
	// in-place tick change no display diff can see.
	fullGen      uint64
	stateBinding state.Binding
	elements     *element.Registry
	frameState   // deadline: loop sets, window takes; lock: atomic
	index        map[string]*Box
	hits         []hitItem
	pageContent  // paint rect: box cache writes; lock: page loop
	scroll       scroll.State
	audio        audioSlot // web preview clip; lock: audioSlot.mu
}

// New parses html as an html/template page. The source is trusted author
// content: Go template actions in it may read fields and methods of the
// data passed to SetData. ownframe does not sandbox template execution, so
// never pass user-controlled or remote HTML.
func New(cfg Config) (*Page, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	source, htmlInfo, err := sourceHTML(cfg)
	if err != nil {
		return nil, err
	}
	if len(source) > MaxSourceSize {
		path := cfg.File
		if path == "" {
			path = "html"
		}
		return nil, sourceTooLarge(path, int64(len(source)))
	}
	if strings.TrimSpace(source) == "" {
		return nil, ErrEmptyHTML
	}
	static := isBuilderDocument(source)
	if !static {
		if err := checkDOMLimits(source); err != nil {
			return nil, err
		}
	}
	schemes, err := acceptedDeepLinkSchemes(cfg.DeepLinkSchemes)
	if err != nil {
		return nil, err
	}
	minW, minH := cfg.mins()
	var tpl *template.Template
	if !static {
		tpl, err = template.New("page").Parse(source)
		if err != nil {
			return nil, err
		}
	}
	themeSrc, themeInfo, err := sourceTheme(cfg)
	if err != nil {
		return nil, err
	}
	theme, err := parseTheme(themeSrc)
	if err != nil {
		return nil, err
	}
	page := assemblePage(cfg, tpl, theme, themeSrc, source, htmlInfo, themeInfo, minW, minH, schemes)
	if static {
		page.staticSource = true
		page.tplProbeTpl, page.tplProbeHas = nil, false
	}
	page.width, page.height = page.Clamp(cfg.Width, cfg.Height)
	return page, nil
}

// assemblePage builds the Page shell; New clamps the size after it.
func assemblePage(cfg Config, tpl *template.Template, theme *css.Sheet, themeSrc, source string, htmlInfo, themeInfo os.FileInfo, minW, minH int, schemes map[string]struct{}) *Page {
	title := cfg.Title
	if title == "" {
		title = "ownframe"
	}
	p := &Page{
		title:    title,
		tpl:      tpl,
		theme:    theme,
		themeSrc: themeSrc,
		pageGeom: pageGeom{
			minWidth:  minW,
			minHeight: minH,
			maxWidth:  cfg.MaxWidth,
			maxHeight: cfg.MaxHeight,
		},
		nav:              nav.New(source),
		deepLinkSchemes:  schemes,
		devtools:         cfg.DevTools,
		perf:             cfg.Perf || cfg.DevTools,
		perfOptIn:        cfg.Perf,
		motion:           motion.New(),
		retainState:      cfg.RetainState,
		lockView:         cfg.LockView,
		allowTouchZoom:   !cfg.DisableTouchZoom,
		disableWheelZoom: cfg.DisableWheelZoom,
		disableKeyZoom:   cfg.DisableKeyZoom,
		scroll:           scroll.New(cfg.Scrollbar),
		watch:            watchFor(cfg, []byte(source), htmlInfo, []byte(themeSrc), themeInfo),
		elements:         element.NewRegistry(),
		formDirty:        true,
	}
	p.surf = redrawSurface{p}
	p.imageResolver = func(src string) ([]byte, error) { return imageBytes(p, src) }
	return p
}

// ViewLocked reports that the window should scale this page to the screen
// and ignore scroll and pinch zoom.
func (p *Page) ViewLocked() bool {
	if p == nil {
		return false
	}

	return p.lockView
}

var (
	_ host.Focuser         = (*Page)(nil)
	_ host.ContextMenu     = (*Page)(nil)
	_ host.CursorShape     = (*Page)(nil)
	_ host.ScrollRequester = (*Page)(nil)
	_ host.LongPresser     = (*Page)(nil)
	_ host.Swiper          = (*Page)(nil)
	_ host.Accessible      = (*Page)(nil)
)

// Fronter is the front-half surface every sink needs.
type Fronter = host.Fronter

// Front is one display-else-image snapshot.
type Front = host.Front

// SelectFront prefers the display list, then the picture bounds, then the
// frame size. It calls only Display, Image, Size, and Generation, so the
// window and the browser share one preference and keep only sink delivery.
func SelectFront(s Fronter) Front { return host.SelectFront(s) }

// pageContent caches the painted rect. Embedded in Page so promotion
// keeps p.content working and page_types.go stays in the size limit.
// Writer: the box cache stores and invalidates. Lock: the page loop
// serializes all writes; readers run on the same loop.
type pageContent struct {
	content   image.Rectangle
	contentW  int
	contentH  int
	contentOK bool
}

// pageGeom is the frame size in CSS pixels. It is embedded in Page so
// promotion keeps p.width working and page_types.go stays in the file
// size limit. Writer: New once, then SetSize. Lock: the page loop
// serializes all writes; readers run on the same loop.
type pageGeom struct {
	width     int
	height    int
	minWidth  int
	minHeight int
	maxWidth  int
	maxHeight int
}

// Handle registers the input functions. A later call replaces the whole
// set; use AddHandlers to merge one callback while keeping the rest.
func (p *Page) Handle(h Handlers) {
	p.handlers = h
}

// SetData stores the value the template prints on the next Redraw.
func (p *Page) SetData(data any) {
	p.dataMu.Lock()
	p.data = data
	p.formDirty = true
	p.dataMu.Unlock()
}

// Title returns the window title.
func (p *Page) Title() string {
	return p.title
}

// Size returns the frame size in CSS pixels.
func (p *Page) Size() (int, int) {
	return p.width, p.height
}

// MinSize returns the configured desktop minimum. Responsive mobile pages
// follow the native viewport instead.
func (p *Page) MinSize() (int, int) {
	return p.minWidth, p.minHeight
}

// MaxSize returns the largest window the page asks its host for. The
// window clamps to it; the layout does not.
func (p *Page) MaxSize() (int, int) {
	return p.maxWidth, p.maxHeight
}

// Clamp applies the desktop minimum or the native mobile viewport. LockView
// retains its canvas minimum on mobile. The maximum only bounds the window.
func (p *Page) Clamp(width, height int) (int, int) {
	if p.mobileViewport && !p.lockView {
		return clampMin(width, 1), clampMin(height, 1)
	}
	return clampMin(width, p.minWidth), clampMin(height, p.minHeight)
}

// SetSize stores the frame size used by the next Redraw.
// Clamp selects the minimum for the current host and canvas policy.
func (p *Page) SetSize(width, height int) {
	w, h := p.Clamp(width, height)
	if w == p.width && h == p.height {
		return
	}
	p.width, p.height = w, h
	// A viewport callback can update template data even when it declines a
	// scroll redraw. Pages without one only need the new layout dimensions.
	if p.scroll.Windowing.On && p.scroll.Windowing.Fn != nil {
		markFormDirty(p)
	}
}

// Generation increases by one on every successful Redraw.
func (p *Page) Generation() uint64 {
	return p.generation
}

// LayerGeneration keys the compositor's retained element layers. It changes
// when a publish or an element patch changes display content, and stays put
// through a re-projection pass, because that pass rewrites operation matrices
// in place. A layer texture reuses its bake across a 3D animation; the layer
// signature still catches a colour or text change.
func (p *Page) LayerGeneration() uint64 {
	return p.layerGen
}

func clampMin(v, min int) int {
	if v < min {
		return min
	}

	return v
}

// SetViewport receives the host's actual CSS viewport. Desktop minimums
// constrain the OS window, not responsive document layout inside that window.
func (p *Page) SetViewport(width, height int) {
	if p.lockView {
		p.SetSize(width, height)
		return
	}
	p.width, p.height = clampMin(width, 1), clampMin(height, 1)
}

// Prepare draws the page when it has not been drawn yet.
func Prepare(ctx context.Context, page *Page) error {
	if ctx == nil {
		return errNilContext
	}

	if page == nil {
		return ErrNilPage
	}

	if page.Image() == nil && page.Display() == nil {
		return page.Redraw(ctx)
	}

	return nil
}

// PrepareMobile prepares a page for its native mobile viewport.
func PrepareMobile(ctx context.Context, p *Page) error {
	if ctx == nil {
		return errNilContext
	}
	if p == nil {
		return ErrNilPage
	}
	p.mobileViewport = true
	return Prepare(ctx, p)
}

// PrepareMobile prepares this page for its native mobile viewport.
func (p *Page) PrepareMobile(ctx context.Context) error {
	return PrepareMobile(ctx, p)
}

// ErrAdoptSource reports a source that cannot be adopted after a tree patch.
var ErrAdoptSource = errors.New("page: adopted source is not a builder document")

// AdoptPatchedSource records source as the current document after the caller
// has patched the cached tree to match it. It never parses, executes a
// template, or draws. The builder uses it after an incremental update so
// Page.HTML and the rendered tree stay in step without paying a full rebuild.
//
// The source must be a builder document: the builder escapes delimiters, so it
// carries no template actions and the patched tree is the whole truth. Any
// other source is refused instead of silently diverging from a template.
func AdoptPatchedSource(p *Page, source string) error {
	if p == nil {
		return ErrNilPage
	}
	if len(source) > MaxSourceSize {
		return sourceTooLarge("html", int64(len(source)))
	}
	if !isBuilderDocument(source) {
		return ErrAdoptSource
	}

	// One FNV-1a fold serves every consumer: nav stores this hash instead
	// of rescanning the source, the source memo reuses it instead of
	// rehashing, and the sync memo pins it with the body so a later clean
	// redraw never folds the same bytes again.
	hash := draw.HashString(source)

	p.nav.RetargetHashed(source, hash)
	p.dataMu.Lock()
	p.source = source
	p.lastSyncBody = source
	p.lastBody = hashedBody{source: source, hash: hash}
	p.syncClean = true
	p.formDirty = false
	p.dataMu.Unlock()
	p.staticSource = true
	p.tpl = nil
	p.tplProbeTpl, p.tplProbeHas = nil, false
	noteSourceHashed(p, source, hash)
	noteCacheSource(p)
	rememberPatched(p, renderState(p))
	invalidateAccessibilityDocument(p)

	return nil
}

// SanitizeBidi strips bidi controls from untrusted text. Call it on
// chat, notification, and other app-external strings before SetData.
func SanitizeBidi(s string) string { return draw.SanitizeBidi(s) }

// CheckIdentifier enforces the precis nickname profile on handles,
// room names, and other identifiers rendered from untrusted input.
func CheckIdentifier(s string) error { return draw.CheckIdentifier(s) }

// emojiPass swaps supported emoji text for images and adds the sizing rule
// when anything changed. It runs after the control rewrite, so field
// values and painted text share one pass and caret spans are untouched.
func emojiPass(source string) string { return draw.EmojiPass(source) }

// Display is the retained vector list behind a replayable page.
type Display = layout.Display

// DisplayOp is one operation in a Display. A frame callback may change its
// paint fields; see Page.SetTick.
type DisplayOp = layout.DisplayOp

// DisplayOpFillRect, DisplayOpText, and DisplayOpLinkURI are the operation
// kinds the page and frame helpers look for.
const (
	DisplayOpFillRect = layout.DisplayOpFillRect
	DisplayOpText     = layout.DisplayOpText
	DisplayOpLinkURI  = layout.DisplayOpLinkURI
)
