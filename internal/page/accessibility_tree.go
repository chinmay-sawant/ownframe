package page

import (
	"github.com/chinmay-sawant/blinkless/markup"
	"github.com/chinmay-sawant/ownframe/internal/host"
)

// AccessibilityTree returns current semantics in document order. Call it on
// the page's owning loop; native adapters must queue actions back to that loop.
func (p *Page) AccessibilityTree() (host.AccessibilityTree, error) {
	s := p.accessibility
	if s == nil || s.document == 0 {
		return host.AccessibilityTree{}, host.ErrAccessibilityUnavailable
	}
	if s.root == nil {
		root, err := markup.Parse([]byte(s.source))
		if err != nil {
			return host.AccessibilityTree{}, err
		}
		s.root = root
	}
	return p.buildAccessibilityTree(s.root)
}

func (p *Page) buildAccessibilityTree(root *markup.Node) (host.AccessibilityTree, error) {
	b := newAccessibilityBuilder(p, root)
	b.walk(root, "document", "")
	b.tree.Document = p.accessibility.document
	b.tree.Revision = p.generation
	b.tree.Root = "document"
	b.tree.Focused = b.focused
	return b.tree, nil
}
