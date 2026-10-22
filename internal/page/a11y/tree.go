package a11y

import (
	"strings"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/blinkless/markup"
	"github.com/chinmay-sawant/ownframe/internal/host"
)

// View is the page data one tree reads.
type View struct {
	Title    string
	Boxes    []layout.Box
	Controls map[string]Control
	Focused  string
	Document uint64
	Revision uint64
}

// Tree returns the current snapshot. A missing state or a zero document
// is unavailable. The parsed root is cached on s.
func Tree(s *State, view View) (host.AccessibilityTree, error) {
	if s == nil || s.document == 0 {
		return host.AccessibilityTree{}, host.ErrAccessibilityUnavailable
	}
	root, err := s.ensureRoot()
	if err != nil {
		return host.AccessibilityTree{}, err
	}
	view.Document = s.document
	return build(root, view), nil
}

type builder struct {
	controls map[string]Control
	focusID  string
	tree     host.AccessibilityTree
	focused  host.AccessibilityNodeID
	names    Names
	boxes    map[string]layout.Box
	counts   map[string]int
	nodeBox  map[*markup.Node]layout.Box
	nodeAt   map[host.AccessibilityNodeID]int
}

func newBuilder(root *markup.Node, view View) *builder {
	names := NewNames()
	names.Index(root, nil)
	boxes, counts := UniqueBoxes(view.Boxes)
	b := &builder{
		controls: view.Controls,
		focusID:  view.Focused,
		names:    names,
		boxes:    boxes,
		counts:   counts,
		nodeBox:  NodeBoxes(root, view.Boxes),
		nodeAt:   map[host.AccessibilityNodeID]int{"document": 0},
	}
	b.tree.Nodes = []host.AccessibilityNode{{
		ID:   "document",
		Role: host.AccessibilityDocument,
		Name: view.Title,
	}}
	return b
}

func build(root *markup.Node, view View) host.AccessibilityTree {
	b := newBuilder(root, view)
	b.walk(root, "document", "")
	b.tree.Document = view.Document
	b.tree.Revision = view.Revision
	b.tree.Root = "document"
	b.tree.Focused = b.focused
	return b.tree
}

func (b *builder) walk(n *markup.Node, parent host.AccessibilityNodeID, path string) {
	if n == nil {
		return
	}
	if n.Type == markup.TypeText {
		b.appendText(n, parent, path)
		return
	}
	if n.Type != markup.TypeElement || Hidden(n) {
		return
	}
	role, ok := Role(n)
	if ok {
		id := ID(n, path, b.names.IDs)
		node := b.makeNode(
			n,
			id,
			parent,
			role,
		)
		b.tree.Nodes = append(b.tree.Nodes, node)
		b.nodeAt[id] = len(b.tree.Nodes) - 1
		parentIndex := b.nodeAt[parent]
		b.tree.Nodes[parentIndex].Children = append(b.tree.Nodes[parentIndex].Children, id)
		if node.Focused {
			b.focused = id
		}
		parent = id
	}
	for i, child := range n.Children {
		b.walk(child, parent, ChildPath(path, i))
	}
}

func (b *builder) appendText(n *markup.Node, parent host.AccessibilityNodeID, path string) {
	name := strings.Join(strings.Fields(n.Text), " ")
	if name == "" || TextNamedByParent(b.tree.Nodes[b.nodeAt[parent]].Role) {
		return
	}
	id := host.AccessibilityNodeID("path:" + path)
	node := host.AccessibilityNode{
		ID:     id,
		Parent: parent,
		Role:   host.AccessibilityText,
		Name:   name,
	}
	b.nodeAt[id] = len(b.tree.Nodes)
	b.tree.Nodes = append(b.tree.Nodes, node)
	i := b.nodeAt[parent]
	b.tree.Nodes[i].Children = append(b.tree.Nodes[i].Children, id)
}
