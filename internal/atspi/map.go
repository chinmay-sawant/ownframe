package atspi

import (
	"github.com/chinmay-sawant/ownframe/internal/host"
)

// Map converts a detached page tree into AT-SPI objects. Parent and
// child order follow the snapshot. The application root is a synthetic
// wrapper with role application; the page document is its only child.
// That is the shape Orca expects when it looks for an active window.
func Map(tree host.AccessibilityTree) Application {
	paths := make(map[host.AccessibilityNodeID]string, len(tree.Nodes))
	for _, n := range tree.Nodes {
		paths[n.ID] = objectPath(n.ID)
	}
	objects := make([]Object, 0, len(tree.Nodes)+1)
	for _, n := range tree.Nodes {
		objects = append(objects, mapNode(n, paths, tree.Focused))
	}
	rootPath := paths[tree.Root]

	name := ApplicationName
	for _, n := range tree.Nodes {
		if n.ID == tree.Root && n.Name != "" {
			name = n.Name
			break
		}
	}
	appRoot := Object{
		Path:     accessibleRootPath,
		Role:     RoleApplication,
		Name:     name,
		Children: nil,
		State:    State{Enabled: true, Sensitive: true, Active: true},
	}
	if rootPath != "" {
		appRoot.Children = []string{rootPath}
	}
	for i := range objects {
		if objects[i].Path == rootPath {
			objects[i].Parent = accessibleRootPath
			objects[i].State.Active = true
			break
		}
	}
	objects = append([]Object{appRoot}, objects...)
	return Application{
		Name:     ApplicationName,
		RootPath: accessibleRootPath,
		Document: tree.Document,
		Revision: tree.Revision,
		Focused:  tree.Focused,
		Objects:  objects,
	}
}
