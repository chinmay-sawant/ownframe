package page

import (
	"time"

	"github.com/chinmay-sawant/ownframe/internal/host"
)

// DevTools reports whether the window devtools overlay starts on.
func (p *Page) DevTools() bool {
	return p.devtools
}

// SetDevTools turns the window devtools overlay on or off.
func (p *Page) SetDevTools(on bool) {
	p.devtools = on
	p.setPerfEnabled(on || p.perfOptIn)
}

// SetDrawTime records how long the window spent drawing the last frame, so
// Stats reports it as LastDraw. The window calls this through its own
// interface; host.Inspector keeps its shape.
func (p *Page) SetDrawTime(d time.Duration) {
	p.stats.frameDraw = d
}

var _ host.Inspector = (*Page)(nil)
