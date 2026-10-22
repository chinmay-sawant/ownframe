package page

func (p *Page) syncForm(body string) string {
	if p.formState() == nil || p.formState().doc != p.nav.Doc() {
		p.setFormState(&formState{
			byID:    map[string]Control{},
			doc:     p.nav.Doc(),
			blinkOn: true,
		})
	}

	body = rewriteButtons(body)
	spans := scanControls(body)
	byID, order := mergeControls(p.formState().byID, spans)
	for id, ctrl := range byID {
		if ctrl.Bind == "" {
			continue
		}

		if c, ok := bindRead(p, ctrl); ok {
			byID[id] = c
		}
	}

	p.formState().byID = byID
	p.formState().order = order
	if _, ok := byID[p.formState().focusID]; !ok {
		p.formState().focusID = ""
		p.clearRange()
	}
	p.clampRange()

	return emojiPass(rewriteControlsState(body, spans, byID, p.caretOf()))
}
