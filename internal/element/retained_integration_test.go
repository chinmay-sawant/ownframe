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

func TestDisposalQueueOrdering(t *testing.T) {
	n1 := NewNode("a", "div")
	n2 := NewNode("b", "div")
	n3 := NewNode("c", "div")

	q := NewDisposalQueue()
	q.Enqueue(n1)
	q.Enqueue(n2)
	q.Enqueue(n3)

	if q.Pending() != 3 {
		t.Fatalf("pending=%d", q.Pending())
	}
	out := q.Flush()
	if len(out) != 3 {
		t.Fatalf("flushed=%d", len(out))
	}
	for i, n := range []*Node{n1, n2, n3} {
		if out[i] != n {
			t.Fatalf("order mismatch at %d", i)
		}
		if !n.Disposed {
			t.Fatalf("node %d not disposed", i)
		}
	}
	if q.Pending() != 0 {
		t.Fatalf("pending after flush=%d", q.Pending())
	}

	// DisposeStale keeps queue order stable while removing nodes.
	root := NewNode("root", "div")
	child := NewNode("child", "span")
	root.Append(child)
	alive := map[string]bool{}
	removed := DisposeStale(root, alive, q)
	if len(removed) != 2 {
		t.Fatalf("removed=%d", len(removed))
	}
	if removed[0] != child || removed[1] != root {
		t.Fatalf("disposal order mismatch: %v", removed)
	}
	if q.Pending() != 2 {
		t.Fatalf("pending=%d", q.Pending())
	}
	q.Flush()
	if !child.Disposed || !root.Disposed {
		t.Fatal("queued nodes not disposed")
	}
}

func TestLifecycleEvents(t *testing.T) {
	n := NewNode("x", "span")
	var events []string
	var phase string
	n.Handlers["mount"] = func(e Event) bool {
		events = append(events, "mount")
		if !n.Mounted || !n.Lifecycle.Mounted {
			t.Fatal("not mounted during mount event")
		}
		return false
	}
	n.Handlers["update"] = func(e Event) bool {
		events = append(events, "update")
		phase = e.Phase
		if !n.Lifecycle.Updated {
			t.Fatal("not updated during update event")
		}
		return false
	}
	n.Handlers["unmount"] = func(e Event) bool {
		events = append(events, "unmount")
		return false
	}

	Mount(n)
	Update(n, []string{"attr:color", "metadata:count"})
	Unmount(n)

	if len(events) != 3 || events[0] != "mount" || events[1] != "update" || events[2] != "unmount" {
		t.Fatalf("events=%v", events)
	}
	if phase != "attr:color,metadata:count" {
		t.Fatalf("phase=%q", phase)
	}
	if n.Mounted {
		t.Fatal("still mounted after unmount")
	}
	if !n.Lifecycle.Unmounted {
		t.Fatal("lifecycle unmounted not recorded")
	}
}

func containsNode(nodes []*Node, target *Node) bool {
	for _, n := range nodes {
		if n == target {
			return true
		}
	}
	return false
}
