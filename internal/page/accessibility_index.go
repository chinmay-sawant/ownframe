package page

import (
	"github.com/chinmay-sawant/blinkless/markup"
	"github.com/chinmay-sawant/ownframe/internal/host"
	"github.com/chinmay-sawant/ownframe/internal/page/a11y"
)

type accessibilityBuilder struct {
	page    *Page
	tree    host.AccessibilityTree
	focused host.AccessibilityNodeID
	names   a11y.Names
	boxes   map[string]Box
	counts  map[string]int
	nodeBox map[*markup.Node]Box
	nodeAt  map[host.AccessibilityNodeID]int
}

func newAccessibilityBuilder(p *Page, root *markup.Node) *accessibilityBuilder {
	names := a11y.NewNames()
	names.Index(root, nil)
	boxes, counts := a11y.UniqueBoxes(p.boxes)
	b := &accessibilityBuilder{
		page: p, names: names, boxes: boxes, counts: counts,
		nodeBox: a11y.NodeBoxes(root, p.boxes),
		nodeAt:  map[host.AccessibilityNodeID]int{"document": 0},
	}
	b.tree.Nodes = []host.AccessibilityNode{{
		ID: "document", Role: host.AccessibilityDocument, Name: p.title,
	}}
	return b
}
