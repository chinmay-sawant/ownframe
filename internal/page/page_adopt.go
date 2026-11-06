package page

import (
	"errors"

	"github.com/chinmay-sawant/ownframe/internal/page/draw"
)

// ErrAdoptSource reports a source that cannot be adopted after a tree patch.
var ErrAdoptSource = errors.New("page: adopted source is not a builder document")

// AdoptPatchedSource records source as the current document after the caller
// has patched the cached tree to match it. It never parses, executes a
// template, or draws. The builder uses it after an incremental update so
// Page.HTML and the rendered tree stay in step without paying a full rebuild.
//
// The source must be a builder document: the builder escapes delimiters, so it
// carries no template actions and the patched tree is the whole truth. Any
// other source is refused instead of silently diverging from a template.
func AdoptPatchedSource(p *Page, source string) error {
	if p == nil {
		return ErrNilPage
	}
	if len(source) > MaxSourceSize {
		return sourceTooLarge("html", int64(len(source)))
	}
	if !isBuilderDocument(source) {
		return ErrAdoptSource
	}

	p.nav.Retarget(source)
	p.dataMu.Lock()
	p.source = source
	p.lastSyncBody = source
	p.syncClean = true
	p.formDirty = false
	p.dataMu.Unlock()
	p.bodyHash = draw.HashString(source)
	p.bodyLen = len(source)
	p.staticSource = true
	p.tpl = nil
	p.tplProbeTpl, p.tplProbeHas = nil, false
	p.noteSource(source)
	p.noteCacheSource()
	p.remember(p.renderState())
	p.invalidateAccessibilityDocument()

	return nil
}
