package page

import (
	"context"
	"slices"
	"strings"

	"github.com/chinmay-sawant/ownframe/internal/page/form"
)

// WebControl is one live control for the browser document.
// Value is the stored value, including a password. Never the painted bullets.
type WebControl struct {
	ID       string
	Tag      string
	Type     string
	Name     string
	Value    string
	Checked  bool
	Disabled bool
	Options  []SelectOption
}

// WebControls returns the live controls, one per id, sorted by ID.
// A nil page or a page with no form yet returns nil.
// Options is nil unless the control is a select.
func (p *Page) WebControls() []WebControl {
	if p == nil {
		return nil
	}

	fs := p.formState()
	if fs == nil || len(fs.ByID) == 0 {
		return nil
	}

	out := make([]WebControl, 0, len(fs.ByID))
	for _, c := range fs.ByID {
		out = append(out, webControlFrom(c))
	}

	slices.SortFunc(out, func(a, b WebControl) int {
		return strings.Compare(a.ID, b.ID)
	})

	return out
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

	c, ok := p.control(id)
	if !ok || c.Disabled {
		return nil
	}

	switch {
	case c.Tag == "select":
		return p.applyWebSelect(ctx, c, value)
	case c.Type == "checkbox", c.Type == "radio":
		return nil
	default:
		return p.applyWebField(ctx, id, value)
	}
}

func (p *Page) applyWebSelect(ctx context.Context, c Control, value string) error {
	if c.Value == value {
		return nil
	}

	if err := p.Focus(ctx, c.ID); err != nil {
		return err
	}

	if err := p.beforeEdit(ctx, c.ID); err != nil {
		return err
	}

	p.SetFormValue(c.ID, value)

	return p.finishWebEdit(ctx, c.ID)
}

func (p *Page) applyWebField(ctx context.Context, id, value string) error {
	if err := p.Focus(ctx, id); err != nil {
		return err
	}

	end := form.RuneLen(value)

	return p.editField(ctx, nil, func(c Control, _, _ int) (Control, int, int, bool) {
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

	c, ok := p.control(id)
	if !ok || c.Disabled {
		return nil
	}

	if err := p.Focus(ctx, id); err != nil {
		return err
	}

	if c.Checked == checked {
		return nil
	}

	if err := p.beforeEdit(ctx, id); err != nil {
		return err
	}

	p.SetFormChecked(id, checked)

	return p.finishWebEdit(ctx, id)
}

// finishWebEdit writes the control just stored back into its bound field,
// fires Change, and redraws. A Change error skips the redraw.
func (p *Page) finishWebEdit(ctx context.Context, id string) error {
	if c, ok := p.control(id); ok {
		bindWrite(p, c)
	}

	if err := p.change(ctx, id); err != nil {
		return err
	}

	return p.Redraw(ctx)
}
