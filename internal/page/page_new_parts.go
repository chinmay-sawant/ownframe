package page

import (
	"html/template"
	"image"
	"os"

	"github.com/chinmay-sawant/blinkless/css"
	"github.com/chinmay-sawant/ownframe/internal/element"
)

// assemblePage builds the Page shell; New clamps the size after it.
func assemblePage(cfg Config, tpl *template.Template, theme *css.Sheet, themeSrc, source string, htmlInfo, themeInfo os.FileInfo, minW, minH int, schemes map[string]struct{}) *Page {
	title := cfg.Title
	if title == "" {
		title = "ownframe"
	}
	return &Page{
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
		nav:             newNavHistory(source),
		deepLinkSchemes: schemes,
		devtools:        cfg.DevTools,
		perf:            cfg.Perf,
		lockView:        cfg.LockView,
		scrollbar:       cfg.Scrollbar,
		allowTouchZoom:  !cfg.DisableTouchZoom,
		allowScroll:     true,
		touchScroll:     TouchScrollOptions{Sensitivity: 1, Deceleration: 5},
		watch:           watchFor(cfg, []byte(source), htmlInfo, []byte(themeSrc), themeInfo),
		elements:        element.NewRegistry(),
		pending:         map[string]bool{},
		last:            map[string]image.Rectangle{},
	}
}
