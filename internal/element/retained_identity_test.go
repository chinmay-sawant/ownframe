package element

import "testing"

func TestRetainedIdentityAcrossMultipleUpdates(t *testing.T) {
	root := NewNode("root", "div")
	btn := NewNode("btn", "button")
	root.Append(btn)

	for i := 0; i < 4; i++ {
		next := NewNode("root", "div")
		nextButton := NewNode("btn", "button")
		nextButton.Attrs["label"] = "run"
		nextButton.Metadata["count"] = i
		next.Append(nextButton)

		res := Reconcile(root, next)
		root = res.Root
		if root == nil {
			t.Fatalf("iter %d: missing root", i)
		}
		got := root.Find("btn")
		if got != btn {
			t.Fatalf("iter %d: button identity changed", i)
		}
		if got.Attrs["label"] != "run" {
			t.Fatalf("iter %d: attrs not updated", i)
		}
		md, _ := got.MetadataValue("count")
		if md != i {
			t.Fatalf("iter %d: metadata not updated", i)
		}
		if !containsNode(res.Retained, btn) {
			t.Fatalf("iter %d: retained list missing button", i)
		}
	}
}
