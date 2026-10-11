package page

import (
	"context"
	"html/template"
	"image"

	"github.com/chinmay-sawant/blinkless/css"
	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/host"

	"github.com/chinmay-sawant/ownframe/internal/render"
)

// Page is one HTML template and the last picture it produced.
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
	width           int
	height          int
	minWidth        int
	minHeight       int
	maxWidth        int
	maxHeight       int
	generation      uint64
	touchCancel     uint32
	tick            func(ctx context.Context) error
	past            []string
	pastAt          int
	routes          map[string]string
	form            *formState
	picker          PickFunc
	hover           string
	hoverX          float64
	hoverY          float64
	active          string
	pressUsed       bool
	cache           *render.Cache
	stats           pageStats
	devtools        bool
	perf            bool
	lockView        bool
	mobileViewport  bool
	allowTouchZoom  bool
	allowScroll     bool
	touchScroll     TouchScrollOptions
	watch           *watchState
	dirty           image.Rectangle
	dirtyFull       bool
	frameDirty      bool
	pinZ            int
	pending         map[string]bool
	last            map[string]image.Rectangle
	index           map[string]Box
	content         image.Rectangle
	contentW        int
	contentH        int
	contentOK       bool
	scroll          host.Scroll
	hasScroll       bool
	windowing       windowingState
}
