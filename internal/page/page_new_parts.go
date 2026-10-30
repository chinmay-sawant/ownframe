package page

import (
	"html/template"
	"os"

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
		perf:             cfg.Perf,
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
	return p
}
