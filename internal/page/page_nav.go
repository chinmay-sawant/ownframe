package page

import (
	"context"
	"errors"
	"html/template"
	"strings"

	"github.com/chinmay-sawant/ownframe/internal/page/nav"
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
func (p *Page) sourceReady() bool { return p.tpl != nil || p.staticSource }

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
	identical := p.sourceReady() && navIdentical(html, incomingHash, p.nav.CurrentHash(), p.nav.Current())
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
	p.tpl = tpl
	p.staticSource = static
	if static {
		// The envelope proves no delimiter, so seed the probe: the redraw
		// path returns nav.Current without scanning or executing it.
		p.tplProbeTpl, p.tplProbeHas = nil, false
	} else if !identical {
		// The delimiter verdict rides with the new template, so the redraw
		// probe reuses it instead of rescanning the source it just hashed.
		p.tplProbeTpl = tpl
		p.tplProbeHas = incomingDelim
	}
	p.markFormDirty()
	p.invalidateAccessibilityDocument()
	if !identical {
		p.invalidateCache()
	}
	p.markFull()
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
	identical := p.sourceReady() && navIdentical(source, incomingHash, p.nav.CurrentHash(), p.nav.Current())
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
	previousForm := p.formState()
	previousRequest, previousHasRequest := p.ff.requestFocus, p.ff.hasFocusRequest
	if previousForm != nil {
		next := *previousForm
		p.setFormState(&next)
	}
	p.tpl = tpl
	p.staticSource = static
	p.nav.Retarget(source)
	if static {
		p.tplProbeTpl, p.tplProbeHas = nil, false
	} else if !identical {
		p.tplProbeTpl = tpl
		p.tplProbeHas = incomingDelim
	}
	p.invalidateAccessibilityDocument()
	p.markFormDirty()
	if !identical {
		p.invalidateCache()
	}
	if err := p.Redraw(ctx); err != nil {
		p.tpl = previous
		p.staticSource = previousStatic
		p.tplProbeTpl, p.tplProbeHas = previousProbeTpl, previousProbeHas
		p.draw, p.boxes = previousFrame, previousBoxes
		p.source, p.lastSyncBody = previousBody, previousSynced
		p.nav.Retarget(previousSource)
		p.setFormState(previousForm)
		p.ff.requestFocus, p.ff.hasFocusRequest = previousRequest, previousHasRequest
		p.markFormDirty()
		if !identical {
			p.invalidateCache()
		}
		return err
	}
	// The success path needs no rollback: the rollback state above is dead
	// from here, so the old sources release before the state resolve.
	if p.redrawBatch == 0 {
		p.resolveStates()
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
	return p.drawAt(ctx, p.nav.BackIndex())
}

// Forward draws the next template source.
func (p *Page) Forward(ctx context.Context) error {
	if err := useContext(ctx); err != nil {
		return err
	}
	if !p.nav.CanForward() {
		return ErrNoHistory
	}
	return p.drawAt(ctx, p.nav.ForwardIndex())
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

func (p *Page) drawAt(ctx context.Context, at int) error {
	if err := useContext(ctx); err != nil {
		return err
	}
	target := p.nav.SourceAt(at)
	static := isBuilderDocument(target)
	// Both sides are stored entries: the hashes compare with no rehash of
	// either side, and the string compare guards the collision.
	identical := p.sourceReady() && navIdentical(target, p.nav.HashAt(at), p.nav.CurrentHash(), p.nav.Current())
	if !identical {
		if static {
			// A stored builder document needs no template, so the redraw
			// path returns it directly.
			p.tpl = nil
			p.staticSource = true
			p.tplProbeTpl, p.tplProbeHas = nil, false
		} else {
			tpl, err := template.New("page").Parse(target)
			if err != nil {
				return err
			}
			p.tpl = tpl
			p.staticSource = false
			// The stored delimiter verdict travels with the entry, so the
			// redraw probe never rescans the target.
			p.tplProbeTpl = tpl
			p.tplProbeHas = p.nav.DelimAt(at)
		}
	}
	p.nav.GoTo(at)
	p.markFormDirty()
	p.invalidateAccessibilityDocument()
	if !identical {
		p.invalidateCache()
	}
	p.markFull()
	return p.Redraw(ctx)
}
