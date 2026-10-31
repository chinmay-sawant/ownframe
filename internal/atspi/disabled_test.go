package atspi

import (
	"testing"

	"github.com/chinmay-sawant/ownframe/internal/host"
)

func TestMapDisabledClearsEnabled(t *testing.T) {
	tree := host.AccessibilityTree{
		Document: testDoc, Revision: testRev, Root: "id:x",
		Nodes: []host.AccessibilityNode{{
			ID: "id:x", Role: host.AccessibilityButton, Disabled: true,
		}},
	}
	obj := byID(t, Map(tree), "id:x")
	if obj.State.Enabled || obj.State.Sensitive {
		t.Fatalf("%+v", obj.State)
	}
}
