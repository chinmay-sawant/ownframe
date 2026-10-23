package atspi

import "github.com/chinmay-sawant/ownframe/internal/host"

// Map converts a detached page tree into AT-SPI objects. Parent and
// child order follow the snapshot. Password values are cleared.
func Map(tree host.AccessibilityTree) Application {
	paths := make(map[host.AccessibilityNodeID]string, len(tree.Nodes))
	for _, n := range tree.Nodes {
		paths[n.ID] = objectPath(n.ID)
	}
	objects := make([]Object, 0, len(tree.Nodes))
	for _, n := range tree.Nodes {
		objects = append(objects, mapNode(n, paths, tree.Focused))
	}
	return Application{
		Name:     DefaultName,
		RootPath: paths[tree.Root],
		Document: tree.Document,
		Revision: tree.Revision,
		Focused:  tree.Focused,
		Objects:  objects,
	}
}
