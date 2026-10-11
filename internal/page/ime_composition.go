package page

import "context"

// SetIMEComposition paints temporary IME text without storing or notifying
// the field value. start and end select bytes within text.
func (p *Page) SetIMEComposition(ctx context.Context, text string, start, end int) error {
	if err := useContext(ctx); err != nil {
		return err
	}
	if p.form == nil {
		return nil
	}
	if text == "" {
		if p.form.preedit == "" {
			return nil
		}
		p.form.preeditID, p.form.preedit = "", ""
		p.form.preeditStart, p.form.preeditEnd = 0, 0
		return p.Redraw(ctx)
	}
	if _, ok := p.typingTarget(); !ok {
		return nil
	}
	if p.form.preeditID == p.form.focusID && p.form.preedit == text &&
		p.form.preeditStart == start && p.form.preeditEnd == end {
		return nil
	}
	p.form.preeditID, p.form.preedit = p.form.focusID, text
	p.form.preeditStart, p.form.preeditEnd = start, end
	return p.Redraw(ctx)
}
