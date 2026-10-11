package page

import (
	"github.com/chinmay-sawant/blinkless/markup"
	"github.com/chinmay-sawant/ownframe/internal/host"
)

type accessibilityBuilder struct {
	page    *Page
	tree    host.AccessibilityTree
	focused host.AccessibilityNodeID
	ids     map[string][]*markup.Node
	labels  map[string][]*markup.Node
	parents map[*markup.Node]*markup.Node
	boxes   map[string]Box
	counts  map[string]int
	nodeBox map[*markup.Node]Box
	nodeAt  map[host.AccessibilityNodeID]int
}

func newAccessibilityBuilder(p *Page, root *markup.Node) *accessibilityBuilder {
	b := &accessibilityBuilder{page: p, ids: map[string][]*markup.Node{}, labels: map[string][]*markup.Node{}, parents: map[*markup.Node]*markup.Node{}, nodeAt: map[host.AccessibilityNodeID]int{"document": 0}}
	b.index(root, nil)
	b.boxes, b.counts = uniqueAccessibilityBoxes(p.boxes)
	b.nodeBox = accessibilityNodeBoxes(root, p.boxes)
	b.tree.Nodes = []host.AccessibilityNode{{ID: "document", Role: host.AccessibilityDocument, Name: p.title}}
	return b
}

func (b *accessibilityBuilder) index(n, parent *markup.Node) {
	if n == nil {
		return
	}
	b.parents[n] = parent
	if id := n.Attrs["id"]; id != "" {
		b.ids[id] = append(b.ids[id], n)
	}
	if n.Name == "label" && n.Attrs["for"] != "" {
		b.labels[n.Attrs["for"]] = append(b.labels[n.Attrs["for"]], n)
	}
	for _, child := range n.Children {
		b.index(child, n)
	}
}

func uniqueAccessibilityBoxes(boxes []Box) (map[string]Box, map[string]int) {
	byID, counts := map[string]Box{}, map[string]int{}
	for _, box := range boxes {
		if box.ID != "" {
			byID[box.ID] = box
			counts[box.ID]++
		}
	}
	return byID, counts
}
