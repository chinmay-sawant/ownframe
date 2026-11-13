package page

// ExecutedSource returns the last executed template source with live form
// values merged in. It is the same source the PNG path replays, so a browser
// shell can render the actual tags instead of a picture. Empty before the
// first Redraw.
func (p *Page) ExecutedSource() string {
	if p == nil {
		return ""
	}

	return p.source
}

// ThemeSource returns the raw theme stylesheet SetTheme stored, or "" when
// no theme is set.
func (p *Page) ThemeSource() string {
	if p == nil {
		return ""
	}

	return p.themeSrc
}

// SelectOption is one choice of a select control for the web preview.
type SelectOption struct {
	Value    string
	Label    string
	Selected bool
	Disabled bool
}

// PlaceCaret collapses the caret of the focused field to a rune offset.
// The web preview sends the browser caret with every edit so the server
// applies backspaces and typing where the browser put them. Out-of-range
// offsets clamp. It does nothing when id is not focused. It does not draw;
// the edit that follows it does.
func (p *Page) PlaceCaret(id string, caret int) {
	if p == nil {
		return
	}
	fs := p.formState()
	if fs == nil || fs.FocusID != id {
		return
	}
	if _, ok := p.focusedEditable(); !ok {
		return
	}
	p.tickMutated.Store(true)
	n := 0
	if c, ok := p.focusedEditable(); ok {
		n = runeLen(c.Value)
	}
	fs.SetCaret(caret, n, false)
}

// SelectControls returns the option lists of the live select controls, keyed
// by control id. The executed source carries only the current label, so the
// browser shell needs this to fill real dropdowns. Nil when the page has no
// form yet.
func (p *Page) SelectControls() map[string][]SelectOption {
	if p == nil {
		return nil
	}
	fs := p.formState()
	if fs == nil {
		return nil
	}
	out := make(map[string][]SelectOption)
	for id, c := range fs.ByID {
		if c.Tag != "select" {
			continue
		}
		opts := make([]SelectOption, 0, len(c.Options))
		for _, o := range c.Options {
			opts = append(opts, SelectOption{
				Value:    o.Value,
				Label:    o.Label,
				Selected: o.Selected,
				Disabled: o.Disabled,
			})
		}
		out[id] = opts
	}

	return out
}
