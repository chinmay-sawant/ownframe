package page

import (
	"context"
	"slices"
	"strings"

	"github.com/chinmay-sawant/ownframe/internal/host"
	"github.com/chinmay-sawant/ownframe/internal/page/form"
)

var _ host.WebDOM = (*Page)(nil)

// WebControl is one live control for the browser document.
// Value is the stored value, including a password. Never the painted bullets.
type WebControl = host.WebControl

// WebControls returns the live controls, one per id, sorted by ID.
// A nil page or a page with no form yet returns nil.
// Options is nil unless the control is a select.
func (p *Page) WebControls() []WebControl {
	if p == nil {
		return nil
	}

	fs := formState(p)
	if fs == nil || len(fs.ByID) == 0 {
		return nil
	}

	out := make([]WebControl, 0, len(fs.ByID))
	for _, c := range fs.ByID {
		control := webControlFrom(c)
		control.Focused = fs.FocusID == c.ID
		if control.Focused && supportsBrowserSelection(c) {
			control.SelectionStart, control.SelectionEnd = fs.Bounds(form.RuneLen(c.Value))
		}
		out = append(out, control)
	}

	slices.SortFunc(out, func(a, b WebControl) int {
		return strings.Compare(a.ID, b.ID)
	})

	return out
}

func supportsBrowserSelection(c Control) bool {
	if c.Tag == "textarea" {
		return true
	}
	if c.Tag != "input" {
		return false
	}

	switch strings.ToLower(c.Type) {
	case "", "text", "password", "search", "tel", "url":
		return true
	default:
		return false
	}
}

func webControlFrom(c Control) WebControl {
	out := WebControl{
		ID:       c.ID,
		Tag:      c.Tag,
		Type:     c.Type,
		Name:     c.Name,
		Value:    c.Value,
		Checked:  c.Checked,
		Disabled: c.Disabled,
	}
	if c.Tag != "select" {
		return out
	}

	out.Options = make([]SelectOption, len(c.Options))
	for i, opt := range c.Options {
		out.Options[i] = SelectOption{
			Value:    opt.Value,
			Label:    opt.Label,
			Selected: opt.Selected,
			Disabled: opt.Disabled,
		}
	}

	return out
}

// ApplyWebText stores value on id the way a native edit does.
// A select takes an option value. A checkbox or radio is left alone.
// A nil page, a missing control, or a disabled control returns nil.
func (p *Page) ApplyWebText(ctx context.Context, id, value string) error {
	if p == nil {
		return nil
	}

	if err := useContext(ctx); err != nil {
		return err
	}

	c, ok := control(p, id)
	if !ok || c.Disabled {
		return nil
	}

	switch {
	case c.Tag == "select":
		return applyWebSelect(p, ctx, c, value)
	case c.Type == "checkbox", c.Type == "radio":
		return nil
	default:
		return applyWebField(p, ctx, id, value)
	}
}

func applyWebSelect(p *Page, ctx context.Context, c Control, value string) error {
	if c.Value == value {
		return nil
	}

	if err := p.Focus(ctx, c.ID); err != nil {
		return err
	}

	if err := beforeEdit(p, ctx, c.ID); err != nil {
		return err
	}

	p.SetFormValue(c.ID, value)

	return finishWebEdit(p, ctx, c.ID)
}

func applyWebField(p *Page, ctx context.Context, id, value string) error {
	if err := p.Focus(ctx, id); err != nil {
		return err
	}

	end := form.RuneLen(value)

	return editField(p, ctx, nil, func(c Control, _, _ int) (Control, int, int, bool) {
		changed := c.Value != value
		if changed {
			c.Value = value
		}

		return c, end, end, changed
	})
}

// ApplyWebChecked stores checked on id the way a native edit does.
// A nil page, a missing control, or a disabled control returns nil.
func (p *Page) ApplyWebChecked(ctx context.Context, id string, checked bool) error {
	if p == nil {
		return nil
	}

	if err := useContext(ctx); err != nil {
		return err
	}

	c, ok := control(p, id)
	if !ok || c.Disabled {
		return nil
	}

	if err := p.Focus(ctx, id); err != nil {
		return err
	}

	if c.Checked == checked {
		return nil
	}

	if err := beforeEdit(p, ctx, id); err != nil {
		return err
	}

	p.SetFormChecked(id, checked)

	return finishWebEdit(p, ctx, id)
}

// finishWebEdit writes the control just stored back into its bound field,
// fires Change, and redraws. A Change error skips the redraw.
func finishWebEdit(p *Page, ctx context.Context, id string) error {
	if c, ok := control(p, id); ok {
		bindWrite(p, c)
	}

	if err := change(p, ctx, id); err != nil {
		return err
	}

	return p.Redraw(ctx)
}
