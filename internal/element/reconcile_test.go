package element

import "testing"

func TestKeepsInstance(t *testing.T) {
	old := NewNode("root", "div")
	old.Append(NewNode("btn", "button"))
	newRoot := NewNode("root", "div")
	newRoot.Append(NewNode("btn", "button"))
	res := Reconcile(old, newRoot)
	if res.Root != old || old.Find("btn") == nil {
		t.Fatal("not retained")
	}
}

func TestDelete(t *testing.T) {
	old := NewNode("root", "div")
	old.Append(NewNode("a", "span"))
	newRoot := NewNode("root", "div")
	res := Reconcile(old, newRoot)
	if !res.Patches.HasOp(OpDelete) {
		t.Fatal("missing delete")
	}
}

func TestMove(t *testing.T) {
	old := NewNode("root", "div")
	old.Append(NewNode("a", "span"))
	old.Append(NewNode("b", "span"))
	newRoot := NewNode("root", "div")
	newRoot.Append(NewNode("b", "span"))
	newRoot.Append(NewNode("a", "span"))
	res := Reconcile(old, newRoot)
	if !res.Patches.HasOp(OpMove) {
		t.Fatal("missing move")
	}
}

func TestUpdate(t *testing.T) {
	old := NewNode("root", "div")
	old.Attrs["color"] = "red"
	newRoot := NewNode("root", "div")
	newRoot.Attrs["color"] = "blue"
	res := Reconcile(old, newRoot)
	if !res.Patches.HasOp(OpUpdate) || old.Attrs["color"] != "blue" {
		t.Fatal("update broken")
	}
}

func TestCreate(t *testing.T) {
	old := NewNode("root", "div")
	newRoot := NewNode("root", "div")
	added := NewNode("child", "span")
	added.Handlers["click"] = func(Event) bool { return true }
	newRoot.Append(added)
	res := Reconcile(old, newRoot)
	if !res.Patches.HasOp(OpCreate) || old.Find("child") != added {
		t.Fatal("create broken")
	}
}

func TestDisposed(t *testing.T) {
	old := NewNode("root", "div")
	wrap := NewNode("wrap", "div")
	wrap.Append(NewNode("a", "span"))
	old.Append(wrap)
	newRoot := NewNode("root", "div")
	res := Reconcile(old, newRoot)
	if len(res.Disposed) != 2 {
		t.Fatalf("disposed=%d", len(res.Disposed))
	}
}
