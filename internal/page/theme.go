package page

import (
	"strings"

	"github.com/chinmay-sawant/blinkless/css"
)

// SetTheme stores an extra stylesheet the next Redraw applies after the
// page's own style elements. Any property the engine supports restyles the
// page, and a theme rule wins a tie with a template rule. An empty source
// removes the theme. SetTheme does not draw. source is trusted application
// CSS; an oversized source is rejected.
func (p *Page) SetTheme(source string) error {
	sheet, err := parseTheme(source)
	if err != nil {
		return err
	}

	p.theme = sheet
	p.themeSrc = source
	p.invalidateCache()
	p.markFull()

	return nil
}

// parseTheme parses theme CSS. Empty or blank source means no theme.
func parseTheme(source string) (*css.Sheet, error) {
	if len(source) > MaxSourceSize {
		return nil, sourceTooLarge("theme", int64(len(source)))
	}

	if strings.TrimSpace(source) == "" {
		return nil, nil
	}

	return css.Parse(source)
}
