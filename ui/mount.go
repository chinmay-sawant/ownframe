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
	r, err := renderTree(root)
	if err != nil {
		return nil, err
	}
	cfg.HTML = document(cfg.Title, r.output.String())
	p, err := ownframe.New(cfg)
	if err != nil {
		return nil, err
	}
	view := &View{Page: p, title: cfg.Title, actions: r.actions}
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
func (v *View) Update(ctx context.Context, root *Element) error {
	r, err := renderTree(root)
	if err != nil {
		return err
	}
	previous := v.actions
	v.actions = r.actions
	if err := v.Page.Load(ctx, document(v.title, r.output.String()), ownframe.LoadOptions{ReplaceCurrent: true}); err != nil {
		v.actions = previous
		return err
	}
	return nil
}

func renderTree(root *Element) (*renderer, error) {
	if root == nil {
		return nil, fmt.Errorf("ui: a non-nil component tree is required")
	}
	r := newRenderer()
	if err := r.visit(root, 0); err != nil {
		return nil, err
	}
	return r, nil
}

func document(title, fragment string) string {
	return "<!doctype html><html><head><meta charset=\"utf-8\"><title>" +
		escape(title) + "</title><style>html,body{margin:0;padding:0}</style></head><body>" + fragment + "</body></html>"
}
