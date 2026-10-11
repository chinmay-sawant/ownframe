package page

import (
	"context"
	"fmt"
	"net/url"
	"strings"
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
	if strings.TrimSpace(target) != target || strings.ContainsAny(target, "\r\n\t ") {
		return nil, fmt.Errorf("%w: %q", ErrInvalidLink, target)
	}
	link, err := url.Parse(target)
	if err != nil {
		return nil, fmt.Errorf("%w: %q", ErrInvalidLink, target)
	}
	if !link.IsAbs() || link.Scheme == "" {
		return nil, nil
	}
	link.Scheme = strings.ToLower(link.Scheme)
	switch link.Scheme {
	case "http", "https":
		if link.Hostname() == "" {
			return nil, fmt.Errorf("%w: %q", ErrInvalidLink, target)
		}
	case "mailto":
		if link.Opaque == "" && link.Path == "" {
			return nil, fmt.Errorf("%w: %q", ErrInvalidLink, target)
		}
	default:
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedLinkScheme, link.Scheme)
	}
	return link, nil
}

func (p *Page) linkTargetAt(x, y float64) (string, bool) {
	d := p.display
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
