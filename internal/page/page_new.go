package page

import (
	"html/template"
	"strings"
)

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
	schemes, err := acceptedDeepLinkSchemes(cfg.DeepLinkSchemes)
	if err != nil {
		return nil, err
	}
	minW, minH := cfg.mins()
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
	page := assemblePage(cfg, tpl, theme, themeSrc, source, htmlInfo, themeInfo, minW, minH, schemes)
	page.width, page.height = page.Clamp(cfg.Width, cfg.Height)
	return page, nil
}
