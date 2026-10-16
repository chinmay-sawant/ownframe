package page

import "context"

// SetIMEComposition paints temporary IME text without storing or notifying
// the field value. start and end select bytes within text.
func (p *Page) SetIMEComposition(ctx context.Context, text string, start, end int) error {
	if err := useContext(ctx); err != nil {
		return err
	}
	if p.formState() == nil {
		return nil
	}
	if text == "" {
		if p.formState().preedit == "" {
			return nil
		}
		p.formState().preeditID, p.formState().preedit = "", ""
		p.formState().preeditStart, p.formState().preeditEnd = 0, 0
		return p.Redraw(ctx)
	}
	if _, ok := p.typingTarget(); !ok {
		return nil
	}
	if p.formState().preeditID == p.formState().focusID && p.formState().preedit == text &&
		p.formState().preeditStart == start && p.formState().preeditEnd == end {
		return nil
	}
	p.formState().preeditID, p.formState().preedit = p.formState().focusID, text
	p.formState().preeditStart, p.formState().preeditEnd = start, end
	return p.Redraw(ctx)
}
