package page

import (
	"html/template"
	"time"

	"github.com/chinmay-sawant/blinkless/css"
	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/element"
	"github.com/chinmay-sawant/ownframe/internal/motion"
	"github.com/chinmay-sawant/ownframe/internal/page/dirty"
	"github.com/chinmay-sawant/ownframe/internal/page/draw"
	"github.com/chinmay-sawant/ownframe/internal/page/nav"
	"github.com/chinmay-sawant/ownframe/internal/page/scroll"
	"github.com/chinmay-sawant/ownframe/internal/render"
	"github.com/chinmay-sawant/ownframe/internal/state"
)

// Page is one template and its picture.
type Page struct {
	title       string
	tpl         *template.Template
	data        any
	theme       *css.Sheet
	themeSrc    string
	handlers    Handlers
	gesture     pageDrag
	images      map[string][]byte
	draw        draw.Frame
	boxes       []layout.Box
	source      string
	pageGeom    // size: New inits, SetSize writes; lock: page loop
	renderTrack // generation and last frame key; lock: page loop
	// layerGen keys the compositor's retained element layers. It moves when a
	// publish or an element patch changes display content, and stays put when
	// a re-projection pass only rewrites operation matrices in place.
	layerGen    uint64
	touchCancel uint32
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
	motion           *motion.Runtime
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
	painted *Display
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
}
