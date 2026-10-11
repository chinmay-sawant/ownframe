package page

import "github.com/chinmay-sawant/blinkless/markup"

type accessibilityState struct {
	source   string
	root     *markup.Node
	document uint64
	retire   bool
}

func (p *Page) commitAccessibility(source string) {
	if p.accessibility == nil {
		p.accessibility = &accessibilityState{}
	}
	s := p.accessibility
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
	if p.accessibility != nil {
		p.accessibility.retire = true
	}
}
