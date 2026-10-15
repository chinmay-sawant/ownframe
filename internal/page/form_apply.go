package page

func (p *Page) syncForm(body string) string {
	if p.form == nil || p.form.doc != p.nav.doc() {
		p.form = &formState{
			byID:    map[string]Control{},
			doc:     p.nav.doc(),
			blinkOn: true,
		}
	}

	body = rewriteButtons(body)
	spans := scanControls(body)
	byID, order := mergeControls(p.form.byID, spans)
	for id, ctrl := range byID {
		if ctrl.Bind == "" {
			continue
		}

		if c, ok := bindRead(p, ctrl); ok {
			byID[id] = c
		}
	}

	p.form.byID = byID
	p.form.order = order
	if _, ok := byID[p.form.focusID]; !ok {
		p.form.focusID = ""
		p.clearRange()
	}
	p.clampRange()

	return emojiPass(rewriteControlsState(body, spans, byID, p.caretOf()))
}
