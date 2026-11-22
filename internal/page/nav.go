package page

import (
	"context"
	"errors"
	"fmt"
	"github.com/chinmay-sawant/ownframe/internal/page/nav"
	"html/template"
	"net/url"
	"strings"
)

// ErrNoHistory means Back or Forward has no entry.
var ErrNoHistory = errors.New("ownframe: no history")

// HTML returns the source at the current history index.
func (p *Page) HTML() string { return p.nav.Current().HTML }

// Load adds a Go template to history and redraws. Routes use this default
// mode. Template actions can call methods on the page data.
func (p *Page) Load(ctx context.Context, html string) error {
	return LoadHTML(ctx, p, html, false)
}

// LoadHTML adds a source to history in the requested mode and redraws.
func LoadHTML(ctx context.Context, p *Page, html string, static bool) error {
	if err := useContext(ctx); err != nil {
		return err
	}
	source, tpl, identical, err := prepareNavigation(p, html, static)
	if err != nil {
		return err
	}
	p.nav.Push(source)
	p.tpl = tpl
	invalidateSourceState(p, !identical)
	markFull(p)
	return p.Redraw(ctx)
}

func prepareNavigation(p *Page, html string, static bool) (nav.Entry, *template.Template, bool, error) {
	if len(html) > MaxSourceSize {
		return nav.Entry{}, nil, false, sourceTooLarge("html", int64(len(html)))
	}
	if strings.TrimSpace(html) == "" {
		return nav.Entry{}, nil, false, ErrEmptyHTML
	}
	source := nav.Source(html, static)
	if source.SameSource(p.nav.Current()) {
		return source, p.tpl, true, nil
	}
	if err := checkDOMLimits(html); err != nil {
		return nav.Entry{}, nil, false, err
	}
	tpl, err := parseTemplate(source)
	return source, tpl, false, err
}

func validateHTML(html string) error {
	if strings.TrimSpace(html) == "" {
		return ErrEmptyHTML
	}
	return checkDOMLimits(html)
}

func parseTemplate(source nav.Entry) (*template.Template, error) {
	if source.Static {
		return nil, nil
	}
	return template.New("page").Parse(source.HTML)
}

// ReplaceHTML updates a document without adding or dropping history entries.
// Like file reload, stable control IDs keep live values and focus. To change
// a live value, use Page.Form.
func ReplaceHTML(ctx context.Context, p *Page, html string, static bool) error {
	if err := useContext(ctx); err != nil {
		return err
	}
	source, tpl, identical, err := prepareNavigation(p, html, static)
	if err != nil {
		return err
	}
	previous, previousSource := p.tpl, p.nav.Current()
	previousFrame, previousBoxes := p.draw, p.boxes
	previousBody, previousSynced := p.source, p.lastSyncBody
	previousForm := formState(p)
	previousRequest, previousHasRequest := p.ff.requestFocus, p.ff.hasFocusRequest
	if previousForm != nil {
		next := *previousForm
		setFormState(p, &next)
	}
	p.tpl = tpl
	p.nav.Retarget(source)
	invalidateSourceState(p, !identical)
	if err := p.Redraw(ctx); err != nil {
		p.tpl = previous
		p.draw, p.boxes = previousFrame, previousBoxes
		p.source, p.lastSyncBody = previousBody, previousSynced
		p.nav.Retarget(previousSource)
		setFormState(p, previousForm)
		p.ff.requestFocus, p.ff.hasFocusRequest = previousRequest, previousHasRequest
		invalidateSourceState(p, !identical)
		return err
	}
	// The success path needs no rollback: the rollback state above is dead
	// from here, so the old sources release before the state resolve.
	if p.redrawBatch == 0 {
		resolveStates(p)
	}
	return nil
}

// Back draws the previous template source.
func (p *Page) Back(ctx context.Context) error {
	if err := useContext(ctx); err != nil {
		return err
	}
	if !p.nav.CanBack() {
		return ErrNoHistory
	}
	return drawAt(p, ctx, p.nav.BackIndex())
}

// Forward draws the next template source.
func (p *Page) Forward(ctx context.Context) error {
	if err := useContext(ctx); err != nil {
		return err
	}
	if !p.nav.CanForward() {
		return ErrNoHistory
	}
	return drawAt(p, ctx, p.nav.ForwardIndex())
}

// Route maps a data-action value to HTML for Click to load. html is
// trusted author content parsed as a Go template at load time; ownframe
// does not sandbox template execution. An empty action is ignored. The
// same action replaces its HTML.
func (p *Page) Route(action, html string) error {
	if action == "" {
		return nil
	}
	if len(html) > MaxSourceSize {
		return sourceTooLarge("route:"+action, int64(len(html)))
	}
	if err := checkDOMLimits(html); err != nil {
		return err
	}
	p.nav.Route(action, html)
	return nil
}

func drawAt(p *Page, ctx context.Context, at int) error {
	if err := useContext(ctx); err != nil {
		return err
	}
	target := p.nav.At(at)
	identical := target.SameSource(p.nav.Current())
	if !identical {
		tpl, err := parseTemplate(target)
		if err != nil {
			return err
		}
		p.tpl = tpl
	}
	p.nav.GoTo(at)
	invalidateSourceState(p, !identical)
	markFull(p)
	return p.Redraw(ctx)
}

func openLinkAt(p *Page, ctx context.Context, x, y float64) (bool, error) {
	target, found := linkTargetAt(p, x, y)
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

func linkTargetAt(p *Page, x, y float64) (string, bool) {
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

// errRelativeLink marks a target without an absolute scheme. Outbound links
// treat it as in-page fallthrough; deep links reject it as invalid.
var errRelativeLink = errors.New("ownframe: relative link")

// classifyLink parses target once and enforces the shared trust policy:
// surrounding or inner whitespace is invalid, blocked schemes are denied,
// empty-host http is invalid, and empty mailto is invalid. Each caller keeps
// only its allowlist line.
func classifyLink(target string) (*url.URL, error) {
	if strings.TrimSpace(target) != target || strings.ContainsAny(target, "\r\n\t ") {
		return nil, fmt.Errorf("%w: %q", ErrInvalidLink, target)
	}
	link, err := url.Parse(target)
	if err != nil {
		return nil, fmt.Errorf("%w: %q", ErrInvalidLink, target)
	}
	if !link.IsAbs() || link.Scheme == "" {
		return nil, fmt.Errorf("%w: %q", errRelativeLink, target)
	}
	link.Scheme = strings.ToLower(link.Scheme)
	if blockedLinkScheme(link.Scheme) {
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedLinkScheme, link.Scheme)
	}
	switch link.Scheme {
	case "http", "https":
		if link.Hostname() == "" {
			return nil, fmt.Errorf("%w: %q", ErrInvalidLink, target)
		}
	case "mailto":
		if link.Opaque == "" && link.Path == "" {
			return nil, fmt.Errorf("%w: %q", ErrInvalidLink, target)
		}
	}
	if link.Opaque == "" && link.Host == "" && link.Path == "" {
		return nil, fmt.Errorf("%w: %q", ErrInvalidLink, target)
	}
	return link, nil
}

// DispatchDeepLink validates a host-delivered URL and calls the app handler.
// Call it on the page's owning event loop. It does not draw.
func (p *Page) DispatchDeepLink(ctx context.Context, target string) error {
	if err := useContext(ctx); err != nil {
		return err
	}
	if p.handlers.DeepLink == nil {
		return ErrNoDeepLinkHandler
	}
	link, err := parseDeepLink(target, p.deepLinkSchemes)
	if err != nil {
		return err
	}
	return p.handlers.DeepLink(ctx, link)
}

func parseDeepLink(target string, schemes map[string]struct{}) (*url.URL, error) {
	link, err := classifyLink(target)
	if err != nil {
		if errors.Is(err, errRelativeLink) {
			return nil, fmt.Errorf("%w: %q", ErrInvalidLink, target)
		}
		return nil, err
	}
	if _, ok := schemes[link.Scheme]; !ok {
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedLinkScheme, link.Scheme)
	}
	return link, nil
}

// ErrInvalidLink means a URL is malformed or relative.
var ErrInvalidLink = errors.New("ownframe: invalid link")

// ErrUnsupportedLinkScheme means the URL scheme is not allowlisted.
var ErrUnsupportedLinkScheme = errors.New("ownframe: unsupported link scheme")

// ErrBadLinkScheme means Config contains an invalid or blocked scheme.
var ErrBadLinkScheme = errors.New("ownframe: bad link scheme")

// ErrNoDeepLinkHandler means no DeepLink callback is registered.
var ErrNoDeepLinkHandler = errors.New("ownframe: no deep link handler")

func acceptedDeepLinkSchemes(input []string) (map[string]struct{}, error) {
	schemes := make(map[string]struct{}, len(input))
	for _, raw := range input {
		scheme := strings.ToLower(raw)
		if !validLinkScheme(scheme) || blockedLinkScheme(scheme) {
			return nil, fmt.Errorf("%w: %q", ErrBadLinkScheme, raw)
		}
		schemes[scheme] = struct{}{}
	}
	return schemes, nil
}

func validLinkScheme(scheme string) bool {
	if len(scheme) == 0 || !letter(scheme[0]) {
		return false
	}
	for i := 1; i < len(scheme); i++ {
		c := scheme[i]
		if !letter(c) && (c < '0' || c > '9') && c != '+' && c != '-' && c != '.' {
			return false
		}
	}
	return true
}

func letter(c byte) bool { return c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' }

func blockedLinkScheme(scheme string) bool {
	switch scheme {
	case "javascript", "data", "vbscript", "file", "blob", "about":
		return true
	default:
		return false
	}
}

// ValidateOpenTarget allows only http, https, and mailto targets.
// It wraps parseOutboundLink so clicks and helpers share one gate.
// ownframe never invokes an OS opener itself; the app opens the
// returned link after an explicit user action such as a click.
func ValidateOpenTarget(target string) (*url.URL, error) {
	link, err := parseOutboundLink(target)
	if err != nil {
		return nil, err
	}
	if link == nil {
		return nil, ErrInvalidLink
	}
	return link, nil
}
