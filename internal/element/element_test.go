package element

import "testing"

func TestNodeTree(t *testing.T) {
	root := NewNode("root", "div")
	child := NewNode("btn", "button")
	root.Append(child)
	if len(root.Children) != 1 || child.Parent != root {
		t.Fatal("append broken")
	}
	if !root.Remove(child) {
		t.Fatal("remove failed")
	}
	if len(root.Children) != 0 || child.Parent != nil {
		t.Fatal("remove broken")
	}
}

func TestDispatchBubble(t *testing.T) {
	root := NewNode("root", "div")
	child := NewNode("btn", "button")
	root.Append(child)
	called := false
	root.Handlers["click"] = func(e Event) bool { called = true; return false }
	Dispatch(root, Event{Type: "click", Target: child})
	if !called {
		t.Fatal("bubble handler not called")
	}
}

func TestPatchList(t *testing.T) {
	l := PatchList{{Op: OpCreate}, {Op: OpDelete}}
	if !l.HasOp(OpDelete) || l.HasOp(OpMove) {
		t.Fatal("patch list broken")
	}
}

func TestMountUnmount(t *testing.T) {
	n := NewNode("x", "span")
	Mount(n)
	if !n.Mounted {
		t.Fatal("not mounted")
	}
	Unmount(n)
	if n.Mounted {
		t.Fatal("not unmounted")
	}
}
