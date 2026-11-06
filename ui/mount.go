package ui

import (
	"context"
	"fmt"

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
// parsing, cascading, and laying out the whole document again. Mount turns on
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
// buffer and the map capacity. The incremental paths build the registry while
// rendering the snapshot or walking the rebuilt tree.
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
			return v.patch(ctx, root, d.patches)
		}
	}
	return v.renderFull(ctx, root)
}

// patch applies one incremental update. renderDocument builds the source
// snapshot and the callback registry the fresh tree would publish; the batch
// then edits the retained elements. Load with AdoptTree records that snapshot
// as the current source without parsing it, so Page.HTML stays in step.
//
// A refused batch rolls its tree edits back. If the batch or the adopt fails,
// the same snapshot goes through the ordinary ReplaceCurrent load, which
// reparses the whole document, so the page still ends on the new tree.
func (v *View) patch(ctx context.Context, root *Element, patches []ownframe.ElementPatch) error {
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
