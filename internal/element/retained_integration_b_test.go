package element

import "testing"

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
