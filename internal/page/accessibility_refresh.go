package page

import (
	"strings"

	"github.com/chinmay-sawant/blinkless/markup"
)

// refreshAccessibility puts the patch's edits into the accessibility tree and
// retires the document, so an old tree is not served as current. A tree no
// client asked for yet is built from the source when one does.
func (p *Page) refreshAccessibility(id string, edits []ElementEdit) {
	s := p.a11yState()
	if s == nil {
		return
	}
	if s.root != nil {
		b := newAccessibilityBuilder(p, s.root)
		for _, e := range edits {
			if e.ID == "" {
				e.ID = id
			}
			for _, n := range b.names.IDs[e.ID] {
				if e.Text != nil {
					n.Children = []*markup.Node{{Type: markup.TypeText, Text: *e.Text}}
				}
				if e.Attribute != "" {
					if n.Attrs == nil {
						n.Attrs = map[string]string{}
					}
					n.Attrs[strings.ToLower(e.Attribute)] = e.Value
				}
			}
		}
	}
	p.invalidateAccessibilityDocument()
	p.commitAccessibility(s.source)
}
