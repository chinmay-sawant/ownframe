package element

// Lifecycle tracks mount/update/unmount for a retained instance.
type Lifecycle struct {
	Mounted, Updated, Unmounted bool
}

// Mount marks a node mounted and runs its mount hook.
func Mount(n *Node) {
	if n == nil {
		return
	}
	n.Mounted = true
	if h, ok := n.Handlers["mount"]; ok {
		h(Event{Type: "mount", Target: n})
	}
}

// Update marks a node updated.
func Update(n *Node, props []string) {
	if n == nil {
		return
	}
	n.Updated = true
	if h, ok := n.Handlers["update"]; ok {
		h(Event{Type: "update", Target: n, Phase: join(props)})
	}
}

// Unmount detaches state and runs unmount hook.
func Unmount(n *Node) {
	if n == nil {
		return
	}
	if h, ok := n.Handlers["unmount"]; ok {
		h(Event{Type: "unmount", Target: n})
	}
	n.Mounted = false
	n.Parent = nil
	n.Children = nil
}

func join(props []string) string {
	out := ""
	for i, p := range props {
		if i > 0 {
			out += ","
		}
		out += p
	}
	return out
}
