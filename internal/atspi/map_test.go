package atspi

import (
	"testing"

	"github.com/chinmay-sawant/ownframe/internal/host"
)

func TestMapPreservesOrderAndRedactsPassword(t *testing.T) {
	app := Map(sampleTree())
	if app.Name != ApplicationName || app.RootPath != accessibleRootPath || app.Document != testDoc || app.Revision != testRev {
		t.Fatalf("app: %+v", app)
	}
	if len(app.Objects) != 6 {
		t.Fatalf("objects %d", len(app.Objects))
	}
	appRoot := app.Objects[0]
	if appRoot.Role != RoleApplication || appRoot.Path != accessibleRootPath || len(appRoot.Children) != 1 || !appRoot.State.Active {
		t.Fatalf("app root: %+v", appRoot)
	}
	root := byID(t, app, "document")
	if root.Role != RoleDocumentWeb || len(root.Children) != 4 || root.Parent != accessibleRootPath || !root.State.Active {
		t.Fatalf("root: %+v", root)
	}
	name := byID(t, app, "id:name")
	if name.Role != RoleEntry || name.Value != "Ada" || !name.State.Focused || !name.State.Editable {
		t.Fatalf("name: %+v", name)
	}
	if got := name.Actions; len(got) != 2 || got[0] != ActionGrabFocus || got[1] != ActionSetText {
		t.Fatalf("name actions %v", got)
	}
	pw := byID(t, app, "id:pw")
	if pw.Role != RolePasswordText || pw.Value != "" || !pw.Password {
		t.Fatalf("password: %+v", pw)
	}
	box := byID(t, app, "id:on")
	if box.Role != RoleCheckBox || !box.State.Checked || !box.State.Checkable {
		t.Fatalf("checkbox: %+v", box)
	}
	if byID(t, app, "id:go").Role != RolePushButton {
		t.Fatal("button")
	}
}

func byID(t *testing.T, app Application, id host.AccessibilityNodeID) Object {
	t.Helper()
	for _, obj := range app.Objects {
		if obj.ID == id {
			return obj
		}
	}
	t.Fatalf("missing %s", id)
	return Object{}
}

func TestMapPathsSurviveInsertion(t *testing.T) {
	tree := sampleTree()
	before := Map(tree)
	tree.Nodes = append([]host.AccessibilityNode{{ID: "id:inserted", Role: host.AccessibilityText}}, tree.Nodes...)
	after := Map(tree)
	seen := map[string]bool{}
	for _, obj := range after.Objects {
		if seen[obj.Path] {
			t.Fatal("duplicate object path")
		}
		seen[obj.Path] = true
	}
	for _, obj := range before.Objects {
		if byID(t, after, obj.ID).Path != obj.Path {
			t.Fatal("insertion changed an existing object path")
		}
	}
	password := byID(t, after, "id:pw")
	if password.State.Focused {
		t.Fatal("password unexpectedly focused")
	}
	if !password.State.Editable {
		t.Fatal("password should stay editable while hiding its value")
	}
	if password.Value != "" {
		t.Fatal("password value leaked")
	}
	for _, action := range password.Actions {
		if action == ActionSetText {
			t.Fatal("password advertises set-text")
		}
	}
}
