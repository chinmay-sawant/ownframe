package page

import (
	"html/template"

	"github.com/chinmay-sawant/blinkless/css"
	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/element"
	"github.com/chinmay-sawant/ownframe/internal/host"
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
	touchCancel uint32
	tickSlot
	nav             nav.History
	deepLinkSchemes map[string]struct{}
	ff              formFocusState
	picker          PickFunc
	cache           *render.Cache
	stats           pageStats
	devtools        bool
	perf            bool
	retainState     bool
	lockView        bool
	scrollbar       host.ScrollbarMode
	mobileViewport  bool
	allowTouchZoom  bool
	allowScroll     bool
	touchScroll     TouchScrollOptions
	watch           *watchState
	dirty           dirty.Tracker
	lastPatch       DisplayPatch
	changedProps    []string
	stateStore      *state.Store
	elements        *element.Registry
	frameState      // deadline: loop sets, window takes; lock: atomic
	pinZ            int
	index           map[string]Box
	hits            []hitItem
	pageContent     // paint rect: box cache writes; lock: page loop
	scroll          scroll.State
}
