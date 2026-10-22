package atspi

import (
	"testing"

	"github.com/chinmay-sawant/ownframe/internal/host"
)

func TestMapPreservesOrderAndRedactsPassword(t *testing.T) {
	app := Map(sampleTree())
	if app.Name != DefaultName || app.Document != testDoc || app.Revision != testRev {
		t.Fatalf("app: %+v", app)
	}
	if len(app.Objects) != 5 {
		t.Fatalf("objects %d", len(app.Objects))
	}
	root := app.Objects[0]
	if root.Role != RoleDocumentWeb || len(root.Children) != 4 {
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
