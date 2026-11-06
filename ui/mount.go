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
	Page    *ownframe.Page
	title   string
	actions map[string]func(context.Context) error
}

// NewPage constructs a builder page. Use Mount when the tree will be updated.
// Config.HTML and Config.File must be empty; Config.Theme is supported.
func NewPage(cfg ownframe.Config, root *Element) (*ownframe.Page, error) {
	view, err := Mount(cfg, root)
	if err != nil {
		return nil, err
	}
	return view.Page, nil
}

// Mount constructs a page and retains the callback registry for later updates.
func Mount(cfg ownframe.Config, root *Element) (*View, error) {
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
	view := &View{Page: p, title: cfg.Title, actions: r.actions}
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
// The view keeps one callback map for its lifetime: the pooled renderer
// arrives with a cleared map, the view adopts the fresh callbacks on success,
// and the previous map rides back into the pool for the next render, so
// repeated updates reuse both the output buffer and the map capacity.
func (v *View) Update(ctx context.Context, root *Element) error {
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
