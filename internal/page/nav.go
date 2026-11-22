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

// navIdentical reports whether the incoming source matches the current
// history entry. The incoming hash compares against the stored hash, so an
// unequal source fails fast without rehashing the stored side or memcmping;
// the string compare guards the hash collision.
func navIdentical(next string, nextHash, curHash uint64, cur string) bool {
	if nextHash != curHash {
		return false
	}
	return next == cur
}

// builderDocHead, builderDocMid, and builderDocTail envelope every page the
// ui builder assembles: keep them byte-identical to ui.Document's pieces.
// The builder escapes every '{' it emits and only writes supported tags, so
// an enveloped document provably carries no template delimiter; the nav
// seam skips the CheckDOM rescan and pre-seeds the delimiter probe for it.
// Any other document takes the full guards, so hand-written sources keep
// their exact verdicts.
const (
	builderDocHead = "<!doctype html><html><head><meta charset=\"utf-8\"><title>"
	builderDocMid  = "</title><style>html,body{margin:0;padding:0}</style></head><body>"
	builderDocTail = "</body></html>"
)

// isBuilderDocument reports whether s carries the ui.Document envelope: a
// fixed head and tail plus the style marker the builder writes between the
// escaped title and the escaped fragment. The head and tail compares are
// O(1); the marker sits right after the title, so the scan ends near the
// start instead of walking the whole document twice.
func isBuilderDocument(s string) bool {
	if !strings.HasPrefix(s, builderDocHead) || !strings.HasSuffix(s, builderDocTail) {
		return false
	}
	return strings.Contains(s, builderDocMid)
}

// HTML returns the template source at the current history index.
func (p *Page) HTML() string { return p.nav.Current() }

// sourceReady reports that the current nav entry is loaded: either its
// template is parsed, or it is a builder document that needs no template.
// The identical-source checks require it so a page with no loaded source
// never claims a match.
func sourceReady(p *Page) bool { return p.tpl != nil || p.staticSource }

// Load parses html, drops later history, and draws that page. html is
// trusted author content: Go template actions in it may read fields and
// methods of the page data. ownframe does not sandbox template execution.
func (p *Page) Load(ctx context.Context, html string) error {
	if err := useContext(ctx); err != nil {
		return err
	}
	if len(html) > MaxSourceSize {
		return sourceTooLarge("html", int64(len(html)))
	}
	if strings.TrimSpace(html) == "" {
		return ErrEmptyHTML
	}
	// One fused pass hashes the incoming source and probes its delimiter;
	// the stored side is never rehashed. A builder document provably holds
	// no delimiter, so it skips the parse and keeps tpl nil.
	static := isBuilderDocument(html)
	incomingHash, incomingDelim := nav.Scan(html)
	identical := sourceReady(p) && navIdentical(html, incomingHash, p.nav.CurrentHash(), p.nav.Current())
	if !identical && !static {
		if err := checkDOMLimits(html); err != nil {
			return err
		}
	}
	var tpl *template.Template
	if !static {
		if identical {
			tpl = p.tpl
		} else {
			var err error
			tpl, err = template.New("page").Parse(html)
			if err != nil {
				return err
			}
		}
	}
	p.nav.Push(html)
	setSourceTemplate(p, tpl, static, incomingDelim, identical)
	invalidateSourceState(p, !identical)
	markFull(p)
	return p.Redraw(ctx)
}

// ReplaceHTML updates a document without adding or dropping history entries.
// Like file reload, stable control IDs keep live values and focus. To change
// a live value, use Page.Form.
func ReplaceHTML(ctx context.Context, p *Page, source string) error {
	if err := useContext(ctx); err != nil {
		return err
	}
	if len(source) > MaxSourceSize {
		return sourceTooLarge("html", int64(len(source)))
	}
	if strings.TrimSpace(source) == "" {
		return ErrEmptyHTML
	}
	// A builder document carries no template action, so it needs no parse;
	// any other source keeps the full guards and the parsed template.
	static := isBuilderDocument(source)
	incomingHash, incomingDelim := nav.Scan(source)
	identical := sourceReady(p) && navIdentical(source, incomingHash, p.nav.CurrentHash(), p.nav.Current())
	var tpl *template.Template
	if !static {
		if identical {
			tpl = p.tpl
		} else {
			if err := checkDOMLimits(source); err != nil {
				return err
			}
			var err error
			tpl, err = template.New("page").Parse(source)
			if err != nil {
				return err
			}
		}
	}
	previous, previousSource := p.tpl, p.nav.Current()
	previousStatic := p.staticSource
	previousProbeTpl, previousProbeHas := p.tplProbeTpl, p.tplProbeHas
	previousFrame, previousBoxes := p.draw, p.boxes
	previousBody, previousSynced := p.source, p.lastSyncBody
	previousForm := formState(p)
	previousRequest, previousHasRequest := p.ff.requestFocus, p.ff.hasFocusRequest
	if previousForm != nil {
		next := *previousForm
		setFormState(p, &next)
	}
	setSourceTemplate(p, tpl, static, incomingDelim, identical)
	p.nav.Retarget(source)
	invalidateSourceState(p, !identical)
	if err := p.Redraw(ctx); err != nil {
		p.tpl = previous
		p.staticSource = previousStatic
		p.tplProbeTpl, p.tplProbeHas = previousProbeTpl, previousProbeHas
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
	target := p.nav.SourceAt(at)
	static := isBuilderDocument(target)
	// Both sides are stored entries: the hashes compare with no rehash of
	// either side, and the string compare guards the collision.
	identical := sourceReady(p) && navIdentical(target, p.nav.HashAt(at), p.nav.CurrentHash(), p.nav.Current())
	if !identical {
		if static {
			// A stored builder document needs no template, so the redraw
			// path returns it directly.
			setSourceTemplate(p, nil, true, false, false)
		} else {
			tpl, err := template.New("page").Parse(target)
			if err != nil {
				return err
			}
			// The stored delimiter verdict travels with the entry, so the
			// redraw probe never rescans the target.
			setSourceTemplate(p, tpl, false, p.nav.DelimAt(at), false)
		}
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
