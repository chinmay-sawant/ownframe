package a11y

import (
	"strconv"

	"github.com/chinmay-sawant/blinkless/markup"
	"github.com/chinmay-sawant/ownframe/internal/host"
)

// ID is a stable node id: the element id when unique, else a path token.
func ID(n *markup.Node, path string, ids map[string][]*markup.Node) host.AccessibilityNodeID {
	if id := n.Attrs["id"]; id != "" && len(ids[id]) == 1 {
		return host.AccessibilityNodeID("id:" + id)
	}
	return host.AccessibilityNodeID("path:" + path)
}

// ChildPath is the tree path of the child at at.
func ChildPath(parent string, at int) string {
	part := strconv.Itoa(at)
	if parent == "" {
		return part
	}
	return parent + "." + part
}
