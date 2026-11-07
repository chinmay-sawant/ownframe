package ui

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/chinmay-sawant/ownframe"
)

// View owns a builder document and its current click callbacks. Update and
// Handle must run on the page loop, just like Page.Load and Page.Handle.
// Directly replacing Page's click handler bypasses builder callbacks.
type View struct {
	Page  *ownframe.Page
	title string
	// html is the source the current page was mounted or last updated with.
	// Update compares it with Page.HTML so a page that navigated away from
	// the builder document is rebuilt instead of being left untouched.
	html    string
	last    *Element
	actions map[string]func(context.Context) error
}

// NewPage constructs a builder page. Use Mount when the tree will be updated.
// Config.HTML and Config.File must be empty; Config.Theme is supported.
func NewPage(cfg ownframe.Config, root *Element) (*ownframe.Page, error) {
	view, err := mountPage(cfg, root)
	if err != nil {
		return nil, err
	}
	return view.Page, nil
}

// Mount constructs a page and retains the callback registry for later updates.
//
// Mount also turns on Config.RetainState: View.Update's incremental path needs
// the retained layout state PatchElements reads. NewPage, which returns a page
// without a View, leaves the caller's value alone.
func Mount(cfg ownframe.Config, root *Element) (*View, error) {
	cfg.RetainState = true
	return mountPage(cfg, root)
}

func mountPage(cfg ownframe.Config, root *Element) (*View, error) {
	if cfg.File != "" || cfg.HTML != "" {
		return nil, fmt.Errorf("ui: Mount requires empty Config.HTML and Config.File")
	}
	r, html, err := renderDocument(cfg.Title, root)
	if err != nil {
		return nil, err
	}
	cfg.HTML = html
	p, err := ownframe.New(cfg)
	if err != nil {
		putRenderer(r)
		return nil, err
	}
	view := &View{Page: p, title: cfg.Title, html: html, last: root, actions: r.actions}
	r.actions = make(map[string]func(context.Context) error)
	putRenderer(r)
	view.Handle(ownframe.Handlers{})
	return view, nil
}

// Handle installs application handlers while keeping builder OnClick dispatch.
// A builder callback handles its own action; other clicks reach handlers.Click.
func (v *View) Handle(handlers ownframe.Handlers) {
	fallback := handlers.Click
	handlers.Click = func(ctx context.Context, box ownframe.Box) error {
		if callback := v.actions[box.Action]; callback != nil {
			return callback(ctx)
		}
		if fallback != nil {
			return fallback(ctx, box)
		}
		return nil
	}
	v.Page.Handle(handlers)
}

// Update replaces the current document and callback registry without navigating.
// Stable control IDs and types keep live values, caret, selection and focus;
// removed controls lose that state. To change a live value, use Page.Form.
// Validation failures keep the previous source and callback registry. Draws
// obey Page redraw batching: a batch flush can report a later rendering error.
//
// When the rebuilt tree has the same structure as the retained one, Update
// patches the elements whose text, attributes, or styles changed instead of
// parsing, cascading, and laying out the whole document again, and rewrites
// the retained source snapshot with those edits instead of re-serializing the
// tree. Mount turns on
// Config.RetainState for the retained layout state PatchElements reads. A
// structural change, a changed node without a stable id, an attribute that
// disappeared, more than maxPatches changed elements, or a refused batch still
// take the full render, so the page is never left showing stale content.
//
// Update retains the tree it is given until the next call. Build a new tree
// for each update; editing the retained tree in place is not observed by the
// diff.
//
// The full-render path swaps callback maps with the pooled renderer: the view
// adopts the fresh callbacks on success and the previous map rides back into
// the pool for the next render, so repeated updates reuse both the output
// buffer and the map capacity. The fast patch path reuses the registry the
// diff built while walking the rebuilt tree, without rendering.
func (v *View) Update(ctx context.Context, root *Element) error {
	if root == nil {
		return fmt.Errorf("ui: a non-nil component tree is required")
	}
	if v.last != nil && ctx != nil && ctx.Err() == nil && v.Page.HTML() == v.html {
		d := diffTree(v.last, root)
		switch {
		case d.ok && len(d.patches) == 0:
			// Nothing visible changed. Adopt the rebuilt callbacks without
			// touching the page.
			v.actions = d.actions
			v.last = root
			return nil
		case d.ok:
			return v.patch(ctx, root, d.patches, d.actions)
		}
	}
	return v.renderFull(ctx, root)
}

// patch applies one incremental update. The fast path rewrites the retained
// source snapshot with the batch's edits instead of re-serializing the whole
// tree: patchSnapshot applies the edits to v.html directly, the batch then
// edits the retained elements, and Load with AdoptTree records that snapshot
// as the current source without parsing it, so Page.HTML stays in step.
//
// A refused batch, a failed adopt, an oversize snapshot, or a snapshot the
// locator pass cannot prove identical falls back to the ordinary ReplaceCurrent
// load below, which reparses the whole document, so the page still ends on
// the new tree.
func (v *View) patch(ctx context.Context, root *Element, patches []ownframe.ElementPatch, actions map[string]func(context.Context) error) error {
	if html, ok := patchSnapshot(v.html, patches); ok &&
		!textPatchUnderSynthesizedID(root, patches) &&
		len(html) <= ownframe.MaxSourceSize {
		if err := v.Page.PatchElements(ctx, patches); err == nil {
			if err := v.Page.Load(ctx, html, ownframe.LoadOptions{AdoptTree: true}); err == nil {
				v.actions = actions
				v.last, v.html = root, html
				return nil
			}
		}
	}
	r, html, err := renderDocument(v.title, root)
	if err != nil {
		return err
	}
	if len(html) > ownframe.MaxSourceSize {
		// Load refuses the snapshot before it touches the page. Check first
		// so a patch cannot land on a source the page will not adopt.
		putRenderer(r)
		return v.Page.Load(ctx, html, ownframe.LoadOptions{ReplaceCurrent: true})
	}
	fresh := r.actions
	previous := v.actions
	if err := v.Page.PatchElements(ctx, patches); err == nil {
		if err := v.Page.Load(ctx, html, ownframe.LoadOptions{AdoptTree: true}); err == nil {
			v.actions = fresh
			r.actions = previous
			putRenderer(r)
			v.last, v.html = root, html
			return nil
		}
	}
	v.actions = fresh
	if err := v.Page.Load(ctx, html, ownframe.LoadOptions{ReplaceCurrent: true}); err != nil {
		v.actions = previous
		putRenderer(r)
		return err
	}
	r.actions = previous
	putRenderer(r)
	v.last, v.html = root, html
	return nil
}

// patchSnapshot rewrites the retained source with the batch's edits instead
// of re-serializing the whole tree. It reports ok=false when the batch is not
// expressible as string edits, and the caller takes the full render.
//
// Invariant: on ok=true the result is byte-identical to a full render of the
// rebuilt tree. The diff guarantees the preconditions this relies on: equal
// tags in equal order (element positions stable), unchanged ids (locators
// stable), no removed attributes, text changes only on childless elements,
// and at most maxPatches entries. Attribute order in the snapshot stays
// sorted because value replacements never move a name and additions land in
// their sorted slot. Values and text escape with the same escape the renderer
// uses. Anything unexpected (missing or duplicated locators, malformed tags,
// void text, style removals, id renames, edits the patcher cannot mirror)
// falls back instead of guessing.
//
// What still forces a full serialize: a refused batch or adopt, an oversize
// snapshot, a text edit under a synthesized control id (see
// textPatchUnderSynthesizedID), and any snapshot the locator pass cannot
// prove identical.
//
// The batch resolves every edit against the original string, sorts the
// resulting spans, and emits the result once through a sized Builder: one
// allocation total, no per-edit copies of the snapshot.
func patchSnapshot(prev string, patches []ownframe.ElementPatch) (string, bool) {
	if len(patches) == 0 {
		return prev, true
	}
	seen := make(map[string]struct{}, len(patches))
	for _, p := range patches {
		if p.ID == "" {
			return "", false
		}
		if _, dup := seen[p.ID]; dup {
			return "", false
		}
		seen[p.ID] = struct{}{}
	}
	total := 0
	for _, p := range patches {
		total += len(p.Edits)
	}
	spans := make([]snapshotSpan, 0, total)
	for _, p := range patches {
		for _, e := range p.Edits {
			sp, has, ok := snapshotEditSpan(prev, p.ID, e)
			if !ok {
				return "", false
			}
			if has {
				spans = append(spans, sp)
			}
		}
	}
	if len(spans) == 0 {
		return prev, true
	}
	slices.SortFunc(spans, func(a, b snapshotSpan) int {
		if a.start < b.start {
			return -1
		}
		if a.start > b.start {
			return 1
		}
		if a.end < b.end {
			return -1
		}
		if a.end > b.end {
			return 1
		}
		return strings.Compare(a.name, b.name)
	})
	delta := 0
	pos := 0
	for _, sp := range spans {
		if sp.start < 0 || sp.end < sp.start || sp.end > len(prev) {
			return "", false
		}
		if sp.start < pos {
			// Overlapping replacements: the batch disagrees with the
			// string about what lives here. Decline to the full render.
			return "", false
		}
		pos = sp.end
		delta += len(sp.repl) - (sp.end - sp.start)
	}
	var sb strings.Builder
	sb.Grow(len(prev) + delta)
	pos = 0
	for _, sp := range spans {
		sb.WriteString(prev[pos:sp.start])
		sb.WriteString(sp.repl)
		pos = sp.end
	}
	sb.WriteString(prev[pos:])
	return sb.String(), true
}

// snapshotSpan replaces prev[start:end] with repl. A zero-length span is an
// insertion. name orders insertions sharing one offset so same-tag additions
// land in sorted attribute order however the edits arrived.
type snapshotSpan struct {
	start int
	end   int
	repl  string
	name  string
}

// snapshotEditSpan resolves one ElementEdit against s to a span, without
// copying the snapshot. An empty edit id targets the patch's element, the
// same rule PatchElements follows. An edit that sets both text and an
// attribute resolves the text, matching patch.Apply; an empty edit resolves
// to no span. ok=false declines exactly where applySnapshotEdit declines.
func snapshotEditSpan(s, patchID string, e ownframe.ElementEdit) (snapshotSpan, bool, bool) {
	target := e.ID
	if target == "" {
		target = patchID
	}
	if target == "" {
		return snapshotSpan{}, false, false
	}
	if e.Text != nil {
		span, spanOK := snapshotTextSpan(s, target, *e.Text)
		return span, spanOK, spanOK
	}
	if e.Attribute == "" {
		return snapshotSpan{}, false, true
	}
	if e.Attribute == "id" {
		// Renames never reach the patch path: the diff refuses them.
		return snapshotSpan{}, false, false
	}
	if e.Attribute == "style" && e.Value == "" {
		// An emptied style drops the attribute on a full render; a patched
		// string that keeps style="" would diverge, so rebuild instead.
		return snapshotSpan{}, false, false
	}
	span, spanOK := snapshotAttrSpan(s, target, e.Attribute, e.Value)
	return span, spanOK, spanOK
}

// snapshotAttrSpan resolves one attribute edit to a span. An absent
// attribute inserts at its sorted slot, so the tag stays byte-identical to a
// full render, which sorts attribute names.
func snapshotAttrSpan(s, id, name, value string) (snapshotSpan, bool) {
	_, tagEnd, attrs, ok := locateSnapshotTag(s, id)
	if !ok {
		return snapshotSpan{}, false
	}
	esc := escape(value)
	for _, a := range attrs {
		if a.name == name {
			return snapshotSpan{start: a.valueStart, end: a.valueEnd, repl: esc, name: name}, true
		}
	}
	insert := tagEnd
	for _, a := range attrs {
		if a.name > name {
			insert = a.nameStart
			return snapshotSpan{start: insert, end: insert, repl: name + `="` + esc + `" `, name: name}, true
		}
	}
	return snapshotSpan{start: insert, end: insert, repl: " " + name + `="` + esc + `"`, name: name}, true
}

// snapshotTextSpan resolves one text edit to a span. The first "<" after the
// opening tag must open its closing tag; anything else means the element
// holds children the diff promised it does not hold.
func snapshotTextSpan(s, id, text string) (snapshotSpan, bool) {
	tagStart, tagEnd, _, ok := locateSnapshotTag(s, id)
	if !ok {
		return snapshotSpan{}, false
	}
	j := tagStart + 1
	for j < tagEnd && s[j] != ' ' && s[j] != '>' {
		j++
	}
	tag := s[tagStart+1 : j]
	if isVoidTag(tag) {
		return snapshotSpan{}, false
	}
	content := tagEnd + 1
	rel := strings.IndexByte(s[content:], '<')
	if rel < 0 || !strings.HasPrefix(s[content+rel:], "</"+tag+">") {
		return snapshotSpan{}, false
	}
	return snapshotSpan{start: content, end: content + rel, repl: escape(text)}, true
}

// applySnapshotEdit applies one ElementEdit to the snapshot string. An empty
// edit id targets the patch's element, the same rule PatchElements follows.
// An edit that sets both text and an attribute applies the text, matching
// patch.Apply; an empty edit is a no-op in both.
func applySnapshotEdit(s, patchID string, e ownframe.ElementEdit) (string, bool) {
	sp, has, ok := snapshotEditSpan(s, patchID, e)
	if !ok {
		return "", false
	}
	if !has {
		return s, true
	}
	var sb strings.Builder
	sb.Grow(len(s) + len(sp.repl) - (sp.end - sp.start))
	sb.WriteString(s[:sp.start])
	sb.WriteString(sp.repl)
	sb.WriteString(s[sp.end:])
	return sb.String(), true
}

// snapshotAttr is one name="value" pair inside an opening tag, with byte
// offsets into the snapshot for a value replacement or a sorted insertion.
type snapshotAttr struct {
	name       string
	nameStart  int
	valueStart int
	valueEnd   int
}

// patchSnapshotAttr sets one attribute on the located element. An absent
// attribute is inserted at its sorted slot, so the tag stays byte-identical
// to a full render, which sorts attribute names.
func patchSnapshotAttr(s, id, name, value string) (string, bool) {
	sp, ok := snapshotAttrSpan(s, id, name, value)
	if !ok {
		return "", false
	}
	var sb strings.Builder
	sb.Grow(len(s) + len(sp.repl) - (sp.end - sp.start))
	sb.WriteString(s[:sp.start])
	sb.WriteString(sp.repl)
	sb.WriteString(s[sp.end:])
	return sb.String(), true
}

// patchSnapshotText replaces the text of a childless element. The first "<"
// after the opening tag must open its closing tag; anything else means the
// element holds children the diff promised it does not hold.
func patchSnapshotText(s, id, text string) (string, bool) {
	sp, ok := snapshotTextSpan(s, id, text)
	if !ok {
		return "", false
	}
	var sb strings.Builder
	sb.Grow(len(s) + len(sp.repl) - (sp.end - sp.start))
	sb.WriteString(s[:sp.start])
	sb.WriteString(sp.repl)
	sb.WriteString(s[sp.end:])
	return sb.String(), true
}

// locateSnapshotTag finds the opening tag carrying id="..." and parses its
// attributes. The id name must start the match: attribute values and text are
// escaped, so a literal quote only delimits attributes, but a longer name
// such as data-id holds the needle as a suffix. Candidates without a space
// before the name are skipped; more than one real locator refuses, since the
// batch and the string would no longer agree on which element to edit.
func locateSnapshotTag(s, id string) (tagStart, tagEnd int, attrs []snapshotAttr, ok bool) {
	needle := `id="` + escape(id) + `"`
	offset := 0
	found := -1
	for {
		rel := strings.Index(s[offset:], needle)
		if rel < 0 {
			break
		}
		at := offset + rel
		if at > 0 && s[at-1] == ' ' {
			if found >= 0 {
				return 0, 0, nil, false
			}
			found = at
		}
		offset = at + 1
	}
	if found < 0 {
		return 0, 0, nil, false
	}
	tagStart = strings.LastIndex(s[:found], "<")
	if tagStart < 0 || strings.IndexByte(s[tagStart:found], '>') >= 0 {
		return 0, 0, nil, false
	}
	rel := strings.IndexByte(s[found:], '>')
	if rel < 0 {
		return 0, 0, nil, false
	}
	tagEnd = found + rel
	j := tagStart + 1
	for j < tagEnd && s[j] != ' ' && s[j] != '>' {
		j++
	}
	if j == tagStart+1 {
		return 0, 0, nil, false
	}
	k := j
	for k < tagEnd {
		if s[k] != ' ' {
			return 0, 0, nil, false
		}
		k++
		ns := k
		for k < tagEnd && s[k] != '=' {
			k++
		}
		if k+1 >= tagEnd || s[k] != '=' || s[k+1] != '"' {
			return 0, 0, nil, false
		}
		name := s[ns:k]
		k += 2
		vs := k
		for k < tagEnd && s[k] != '"' {
			k++
		}
		if k >= tagEnd {
			return 0, 0, nil, false
		}
		for _, a := range attrs {
			if a.name == name {
				return 0, 0, nil, false
			}
		}
		attrs = append(attrs, snapshotAttr{name: name, nameStart: ns, valueStart: vs, valueEnd: k})
		k++
	}
	for _, a := range attrs {
		if a.name == "id" && s[a.valueStart:a.valueEnd] == escape(id) {
			return tagStart, tagEnd, attrs, true
		}
	}
	return 0, 0, nil, false
}

// textPatchUnderSynthesizedID reports whether a text edit lands inside the
// subtree of an element whose id is synthesized: an action (OnClick or custom
// data-action) without an explicit id. The control id hashes the element's
// whole subtree text, so a full render would mint a new id for that ancestor
// while the retained display keeps the old one; only the full render keeps
// the source, the display, and Page.HTML in step. Attribute edits cannot move
// those hashes: an ancestor folds only its own action-ish attributes, which a
// patch on an explicitly-identified descendant never renames.
func textPatchUnderSynthesizedID(root *Element, patches []ownframe.ElementPatch) bool {
	textIDs := make(map[string]struct{})
	for _, p := range patches {
		for _, e := range p.Edits {
			if e.Text == nil {
				continue
			}
			id := e.ID
			if id == "" {
				id = p.ID
			}
			if id != "" {
				textIDs[id] = struct{}{}
			}
		}
	}
	if len(textIDs) == 0 {
		return false
	}
	// Plain recursion, no memo map and no path slice: the diff refuses
	// cyclic trees before the patch path, so every node is visited at most a
	// few times and the walk allocates nothing. A shared subtree reached
	// twice yields the same answer twice, so repeat visits stay correct; the
	// old memo table cost a map entry per node (~0.3MB on a 1000-row page).
	var visit func(e *Element, synthAbove bool) bool
	visit = func(e *Element, synthAbove bool) bool {
		if e == nil {
			return false
		}
		if id := e.attrs.value("id"); id != "" {
			if _, patched := textIDs[id]; patched && synthAbove {
				return true
			}
		}
		synth := synthAbove
		if e.attrs.value("id") == "" && (e.onClick != nil || e.attrs.value("data-action") != "") {
			synth = true
		}
		for _, c := range e.children {
			if visit(c, synth) {
				return true
			}
		}
		return false
	}
	return visit(root, false)
}

// renderFull renders the whole document and replaces the current source. It is
// the fallback for structural changes and for updates the patch path cannot
// express.
func (v *View) renderFull(ctx context.Context, root *Element) error {
	r, html, err := renderDocument(v.title, root)
	if err != nil {
		return err
	}
	fresh := r.actions
	previous := v.actions
	v.actions = fresh
	if err := v.Page.Load(ctx, html, ownframe.LoadOptions{ReplaceCurrent: true}); err != nil {
		v.actions = previous
		putRenderer(r)
		return err
	}
	r.actions = previous
	putRenderer(r)
	v.last, v.html = root, html
	return nil
}

// renderDocument renders root straight into a complete builder page: the
// envelope, the escaped title, the tree, and the tail share one buffer, so
// the final String copies the payload once instead of copying the fragment
// and then the document. Output is byte-identical to Document of the title
// and the rendered fragment.
func renderDocument(title string, root *Element) (*renderer, string, error) {
	if root == nil {
		return nil, "", fmt.Errorf("ui: a non-nil component tree is required")
	}
	r := getRenderer()
	r.output.Grow(len(docHead) + len(title) + len(docMid) + estimateSize(root) + len(docTail))
	r.output.WriteString(docHead)
	writeEscaped(&r.output, title)
	r.output.WriteString(docMid)
	if err := r.visit(root, 0); err != nil {
		putRenderer(r)
		return nil, "", err
	}
	r.output.WriteString(docTail)
	return r, r.output.String(), nil
}

func renderTree(root *Element) (*renderer, error) {
	if root == nil {
		return nil, fmt.Errorf("ui: a non-nil component tree is required")
	}
	r := getRenderer()
	if err := r.visit(root, 0); err != nil {
		putRenderer(r)
		return nil, err
	}
	return r, nil
}
