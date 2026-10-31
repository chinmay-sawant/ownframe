package ui

import (
	"context"
	"fmt"

	"github.com/chinmay-sawant/ownframe"
)

// NewPage constructs an ordinary OwnFrame page from a declarative builder
// tree. It passes generated HTML through OwnFrame's existing Blinkless path.
// Existing template applications continue using ownframe.New unchanged.
//
// NewPage owns cfg.HTML. A caller may still set cfg.Theme, size, and the
// other regular OwnFrame options; cfg.File and cfg.HTML must be left empty.
//
// OnClick callbacks use OwnFrame's existing action dispatch and are invoked on
// the owning page loop. No DOM, JS runtime, or retained UI reconciliation is
// added by this experimental adapter.
func NewPage(cfg ownframe.Config, root *Element) (*ownframe.Page, error) {
	if cfg.File != "" || cfg.HTML != "" {
		return nil, fmt.Errorf("ui: NewPage requires empty Config.HTML and Config.File")
	}
	if root == nil {
		return nil, fmt.Errorf("ui: NewPage requires a non-nil component tree")
	}

	r := newRenderer()
	if err := r.visit(root, 0); err != nil {
		return nil, err
	}
	// The built tree is a template *source*. escape() also escapes template
	// braces so text nodes cannot become Go template expressions.
	cfg.HTML = "<!doctype html><html><head><meta charset=\"utf-8\"><title>" +
		escape(cfg.Title) + "</title></head><body>" + r.output.String() + "</body></html>"

	page, err := ownframe.New(cfg)
	if err != nil {
		return nil, err
	}
	if len(r.actions) != 0 {
		// The page is newly constructed; this doesn't overwrite any
		// preexisting click handler. Other handlers may be added later.
		page.AddHandlers(ownframe.Handlers{
			Click: func(ctx context.Context, box ownframe.Box) error {
				if onClick := r.actions[box.Action]; onClick != nil {
					return onClick(ctx)
				}
				return nil
			},
		})
	}
	return page, nil
}
