package page

import (
	"context"
	"errors"
	"fmt"
	"net/url"
)

func (p *Page) openLinkAt(ctx context.Context, x, y float64) (bool, error) {
	target, found := p.linkTargetAt(x, y)
	if !found {
		return false, nil
	}
	link, err := parseOutboundLink(target)
	if err != nil || link == nil {
		return err != nil, err
	}
	if err := p.handlers.OpenLink(ctx, link); err != nil {
		return true, err
	}
	return true, p.Redraw(ctx)
}

func parseOutboundLink(target string) (*url.URL, error) {
	link, err := classifyLink(target)
	if err != nil {
		if errors.Is(err, errRelativeLink) {
			return nil, nil
		}
		return nil, err
	}
	switch link.Scheme {
	case "http", "https", "mailto":
		return link, nil
	default:
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedLinkScheme, link.Scheme)
	}
}

func (p *Page) linkTargetAt(x, y float64) (string, bool) {
	d := p.draw.Display
	if d == nil || d.PointsPerPixel <= 0 {
		return "", false
	}
	pt := d.PointsPerPixel
	xPt, yPt := x*pt, y*pt
	for i := range d.Ops {
		op := &d.Ops[i]
		if op.Kind == DisplayOpLinkURI && xPt >= op.X && xPt <= op.X+op.W && yPt >= op.Y && yPt <= op.Y+op.H {
			return op.LinkURI(), true
		}
	}
	return "", false
}
