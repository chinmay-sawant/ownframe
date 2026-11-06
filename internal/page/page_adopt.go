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

	// One FNV-1a fold serves every consumer: nav stores this hash instead
	// of rescanning the source, the source memo reuses it instead of
	// rehashing, and the sync memo pins it with the body so a later clean
	// redraw never folds the same bytes again.
	hash := draw.HashString(source)

	p.nav.RetargetHashed(source, hash)
	p.dataMu.Lock()
	p.source = source
	p.lastSyncBody = source
	p.lastBody = hashedBody{source: source, hash: hash}
	p.syncClean = true
	p.formDirty = false
	p.dataMu.Unlock()
	p.staticSource = true
	p.tpl = nil
	p.tplProbeTpl, p.tplProbeHas = nil, false
	p.noteSourceHashed(source, hash)
	p.noteCacheSource()
	p.remember(p.renderState())
	p.invalidateAccessibilityDocument()

	return nil
}
