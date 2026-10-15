package element

import "testing"

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
