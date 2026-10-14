package element

// DisposalQueue batches stale removals until before paint.
type DisposalQueue struct {
	items []*Node
}

// NewDisposalQueue creates an empty queue.
func NewDisposalQueue() *DisposalQueue {
	return &DisposalQueue{}
}

// Enqueue adds a node.
func (q *DisposalQueue) Enqueue(n *Node) {
	if q == nil || n == nil {
		return
	}
	q.items = append(q.items, n)
}

// Flush disposes queued nodes and returns them.
func (q *DisposalQueue) Flush() []*Node {
	if q == nil {
		return nil
	}
	out := append([]*Node(nil), q.items...)
	for _, n := range out {
		if n != nil {
			n.Dispose()
		}
	}
	q.items = q.items[:0]
	return out
}

// Pending reports queued count.
func (q *DisposalQueue) Pending() int {
	if q == nil {
		return 0
	}
	return len(q.items)
}

// DisposeStale removes and disposes nodes missing from alive.
func DisposeStale(root *Node, alive map[string]bool, q *DisposalQueue) []*Node {
	if root == nil {
		return nil
	}
	var removed []*Node
	var walk func(n *Node)
	walk = func(n *Node) {
		if n == nil {
			return
		}
		for _, c := range append([]*Node(nil), n.Children...) {
			walk(c)
		}
		key := KeyOf(n)
		if alive != nil && !alive[key] && key != "" {
			removed = append(removed, n)
			if n.Parent != nil {
				n.Parent.Remove(n)
			}
			if q != nil {
				q.Enqueue(n)
			} else {
				n.Dispose()
			}
		}
	}
	walk(root)
	return removed
}
