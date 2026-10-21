package page

import (
	"context"
	"html/template"
	"image"

	"github.com/chinmay-sawant/blinkless/css"
	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/element"
	"github.com/chinmay-sawant/ownframe/internal/host"

	"github.com/chinmay-sawant/ownframe/internal/render"
	"github.com/chinmay-sawant/ownframe/internal/state"
)

// Page is one HTML template and its last picture; serialize mutations.
type Page struct {
	title           string
	tpl             *template.Template
	data            any
	theme           *css.Sheet
	themeSrc        string
	handlers        Handlers
	gesture         pageDrag
	images          map[string][]byte
	img             image.Image
	display         *layout.Display
	fallbackDisplay *layout.Display
	boxes           []layout.Box
	source          string
	png             []byte
	pageGeom        // size: New inits, SetSize writes; lock: page loop
	renderTrack     // generation and last frame key; lock: page loop
	touchCancel     uint32
	tick            func(ctx context.Context) error
	nav             navHistory
	deepLinkSchemes map[string]struct{}
	ff              formFocusState // owner: page_form_focus*.go; lock: page loop
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
	dirty           dirtyTracker
	lastPatch       DisplayPatch
	changedProps    []string
	stateStore      *state.Store
	elements        *element.Registry
	frameState      // deadline: loop sets, window takes; lock: atomic
	pinZ            int
	index           map[string]Box
	hits            []hitItem
	pageContent     // paint rect: box cache writes; lock: page loop
	pageScrollState // scroll: window writes via Set*; lock: loop
}
