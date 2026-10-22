package cat

import (
	"context"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/internal/web"
)

type preview struct {
	*ownframe.Page
	companion *Companion
}

// ServePreview applies inbox notifications before serializing each web frame.
// It keeps all page updates inside the web server's request lock.
func (c *Companion) ServePreview(ctx context.Context, addr string) error {
	if err := c.Page.Redraw(ctx); err != nil {
		return err
	}
	return web.ServeWithContext(ctx, &preview{Page: c.Page, companion: c}, addr)
}

func (p *preview) PollReload(ctx context.Context) (bool, error) {
	if !p.companion.receive() {
		return false, nil
	}
	return true, p.Redraw(ctx)
}

func (p *preview) Watching() bool { return true }
