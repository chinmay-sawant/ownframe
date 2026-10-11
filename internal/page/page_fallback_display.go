package page

import "github.com/chinmay-sawant/blinkless/layout"

// FallbackDisplay returns the retained operations behind a bitmap page.
// Hosts use them to repaint fixed content correctly when the viewport scrolls.
// Callers must not mutate the returned display.
func (p *Page) FallbackDisplay() *layout.Display { return p.fallbackDisplay }
