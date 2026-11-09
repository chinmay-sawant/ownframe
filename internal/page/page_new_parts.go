package page

import (
	"html/template"
	"os"
	"runtime"

	"github.com/chinmay-sawant/blinkless/css"
	"github.com/chinmay-sawant/ownframe/internal/element"
	"github.com/chinmay-sawant/ownframe/internal/motion"
	"github.com/chinmay-sawant/ownframe/internal/page/nav"
	"github.com/chinmay-sawant/ownframe/internal/page/scroll"
	"github.com/chinmay-sawant/ownframe/internal/render"
)

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
		hoverPatch:       render.RelayoutElement,
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
	p.imageResolver = p.imageBytes
	// Evict the per-page box-cache entry when the page is collected. Benches
	// and apps that churn pages (e.g. a fresh page per cold build) otherwise
	// pin every discarded page through the boxCaches map key until the
	// process exits. The finalizer only deletes the map entry; lookup
	// callers already treat a missing entry as a page with no boxes.
	runtime.SetFinalizer(p, func(q *Page) { q.forgetBoxes() })
	return p
}
