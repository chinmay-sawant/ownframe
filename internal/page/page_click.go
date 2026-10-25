package page

import (
	"context"

	"github.com/chinmay-sawant/ownframe/internal/host"
	"github.com/chinmay-sawant/ownframe/internal/page/click"
)

// Click hit-tests the last picture and applies the click.
func (p *Page) Click(ctx context.Context, x, y float64) error {
	if box, ok := p.activationBox(x, y); ok && !p.triggerActivated(box) {
		if _, err := p.fireTriggers(ctx, box, true); err != nil {
			return err
		}
	}

	return click.Run(ctx, clickSurface{p}, x, y)
}

// triggerActivated reports whether the press that started this gesture already
// fired the entry actions for box, so the click does not repeat them.
func (p *Page) triggerActivated(box Box) bool {
	return p.triggerPress.armed && sameBox(p.triggerPress.box, box)
}

// clickSurface adapts Page to click.Surface.
// Each method delegates to the page method the sequence already used.
type clickSurface struct{ p *Page }

func (s clickSurface) HitBox(x, y float64) (Box, bool) {
	return s.p.hitBox(x, y)
}

func (s clickSurface) BoxIDAt(x, y float64) string {
	return s.p.boxIDAt(x, y)
}

func (s clickSurface) BoxByID(id string) Box {
	return s.p.boxByID(id)
}

func (s clickSurface) ConsumePress(id string) bool {
	return s.p.consumePress(id)
}

func (s clickSurface) FormControl(id string) bool {
	return s.p.formControl(id)
}

func (s clickSurface) Control(id string) (Control, bool) {
	return s.p.control(id)
}

func (s clickSurface) Lookup(action string) (string, bool) {
	return s.p.nav.Lookup(action)
}

func (s clickSurface) Load(ctx context.Context, html string) error {
	return s.p.Load(ctx, html)
}

func (s clickSurface) ClickControl(ctx context.Context, box Box) error {
	return s.p.clickControl(ctx, box)
}

func (s clickSurface) BlurForm() {
	s.p.blurForm()
}

func (s clickSurface) HasOpenLink() bool {
	return s.p.handlers.OpenLink != nil
}

func (s clickSurface) OpenLinkAt(ctx context.Context, x, y float64) (bool, error) {
	return s.p.openLinkAt(ctx, x, y)
}

func (s clickSurface) HasClick() bool {
	return s.p.handlers.Click != nil
}

func (s clickSurface) HandlerClick(ctx context.Context, box Box) error {
	return s.p.handlers.Click(ctx, box)
}

func (s clickSurface) ClickPainted(ctx context.Context, box Box) (bool, error) {
	return s.p.clickPainted(ctx, box)
}

func (s clickSurface) Redraw(ctx context.Context) error {
	return s.p.Redraw(ctx)
}

func (s clickSurface) RedrawIfChanged(ctx context.Context) error {
	return s.p.RedrawIfChanged(ctx)
}

var _ host.Screen = (*Page)(nil)

var _ click.Surface = clickSurface{}
