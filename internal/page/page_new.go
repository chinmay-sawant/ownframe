package page

import (
	"html/template"
	"strings"
)

// New parses html as an html/template page.
func New(cfg Config) (*Page, error) {
	source, htmlInfo, err := sourceHTML(cfg)
	if err != nil {
		return nil, err
	}

	if strings.TrimSpace(source) == "" {
		return nil, ErrEmptyHTML
	}
	deepLinkSchemes, err := acceptedDeepLinkSchemes(cfg.DeepLinkSchemes)
	if err != nil {
		return nil, err
	}

	minWidth := cfg.MinWidth
	if minWidth <= 0 {
		minWidth = 1
	}

	minHeight := cfg.MinHeight
	if minHeight <= 0 {
		minHeight = 1
	}

	maxWidth := cfg.MaxWidth
	maxHeight := cfg.MaxHeight

	if cfg.Width <= 0 || cfg.Height <= 0 {
		return nil, ErrBadSize
	}

	if maxWidth > 0 && minWidth > maxWidth {
		return nil, ErrBadSize
	}

	if maxHeight > 0 && minHeight > maxHeight {
		return nil, ErrBadSize
	}

	tpl, err := template.New("page").Parse(source)
	if err != nil {
		return nil, err
	}

	themeSrc, themeInfo, err := sourceTheme(cfg)
	if err != nil {
		return nil, err
	}

	theme, err := parseTheme(themeSrc)
	if err != nil {
		return nil, err
	}

	title := cfg.Title
	if title == "" {
		title = "ownframe"
	}

	page := &Page{
		title:           title,
		tpl:             tpl,
		theme:           theme,
		themeSrc:        themeSrc,
		minWidth:        minWidth,
		minHeight:       minHeight,
		maxWidth:        maxWidth,
		maxHeight:       maxHeight,
		past:            []string{source},
		pastAt:          0,
		deepLinkSchemes: deepLinkSchemes,
		devtools:        cfg.DevTools,
		perf:            cfg.Perf,
		lockView:        cfg.LockView,
		scrollbar:       cfg.Scrollbar,
		allowTouchZoom:  !cfg.DisableTouchZoom,
		allowScroll:     true,
		touchScroll:     TouchScrollOptions{Sensitivity: 1, Deceleration: 5},
		watch:           watchFor(cfg, []byte(source), htmlInfo, []byte(themeSrc), themeInfo),
	}
	page.width, page.height = page.Clamp(cfg.Width, cfg.Height)

	return page, nil
}
