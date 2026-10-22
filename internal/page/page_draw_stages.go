package page

import (
	"context"
	"strings"

	"github.com/chinmay-sawant/blinkless/css"
	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/page/draw"
	"github.com/chinmay-sawant/ownframe/internal/render"
)

// execTemplate runs the trusted template, timing it when tracked.
func (p *Page) execTemplate(track bool) (string, error) {
	ts := draw.Clock(track)
	p.applyScrollWindow()
	var body strings.Builder
	if len(p.source) > 0 {
		body.Grow(len(p.source))
	}
	if err := p.tpl.Execute(&body, p.data); err != nil {
		return "", err
	}
	executed := body.String()
	draw.Note(track, &p.stats.lastTemplate, ts)
	return executed, nil
}

// styleStaged styles the source, timing it when tracked.
func (p *Page) styleStaged(ctx context.Context, track bool, source string, state render.State) (*css.Document, error) {
	ls := draw.Clock(track)
	styled, err := p.styledDocument(ctx, source, state)
	draw.Note(track, &p.stats.layoutTime, ls)
	return styled, err
}

// drawStaged builds the list and commits either path.
func (p *Page) drawStaged(ctx context.Context, track bool, styled *css.Document, state render.State, executed string) error {
	drawStart := draw.Clock(true)
	ds := draw.Clock(track)
	display, err := render.DisplayListDocumentOptions(ctx, styled,
		layout.Options{Images: state.Images, RetainState: state.RetainState})
	draw.Note(track, &p.stats.displayListTime, ds)
	if err != nil {
		return err
	}
	p.paintCaret(display)
	if draw.Choose(display) == draw.PathList {
		p.commitDisplay(display, executed, drawStart, state)
		return nil
	}
	if err := p.paintFallback(display, drawStart, track); err != nil {
		return err
	}
	p.commitAccessibility(executed)
	return nil
}
