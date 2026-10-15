package page

import "github.com/chinmay-sawant/blinkless/markup"

type accessibilityState struct {
	source   string
	root     *markup.Node
	document uint64
	retire   bool
}

func (p *Page) commitAccessibility(source string) {
	if p.a11yState() == nil {
		p.setA11yState(&accessibilityState{})
	}
	s := p.a11yState()
	if s.document == 0 || s.retire {
		s.document++
		s.retire = false
	}
	if s.source != source {
		s.root = nil
	}
	s.source = source
}

func (p *Page) invalidateAccessibilityDocument() {
	if p.a11yState() != nil {
		p.a11yState().retire = true
	}
}
