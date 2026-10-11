package page

import "github.com/chinmay-sawant/ownframe/internal/host"

// ScrollbarMode reports the configured page thumb visibility mode.
func (p *Page) ScrollbarMode() host.ScrollbarMode {
	if p == nil {
		return host.ScrollbarAuto
	}

	return p.scrollbar
}
