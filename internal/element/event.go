package element

// Event is a synthetic event covering keyboard, pointer, focus and IME.
type Event struct {
	Type   string
	Target *Node
	Phase  string // capture, target, bubble
	Key    string
	X, Y   float64
	Text   string
}

// Handler reacts to an event; returning true stops propagation.
type Handler func(Event) bool

// Dispatch walks capture then bubble, calling matching handlers.
func Dispatch(root *Node, e Event) bool {
	if root == nil || e.Target == nil {
		return false
	}
	// Capture phase: root -> target.
	path := pathTo(root, e.Target)
	for _, n := range path {
		if h, ok := n.Handlers["capture:"+e.Type]; ok && h(e) {
			return true
		}
	}
	// Target handlers.
	if h, ok := e.Target.Handlers[e.Type]; ok && h(e) {
		return true
	}
	// Bubble phase: target -> root.
	for i := len(path) - 1; i >= 0; i-- {
		n := path[i]
		if n == e.Target {
			continue
		}
		if h, ok := n.Handlers[e.Type]; ok && h(e) {
			return true
		}
	}
	return false
}

func pathTo(root, target *Node) []*Node {
	// Build parent chain from target up to root.
	var chain []*Node
	cur := target
	for cur != nil {
		chain = append(chain, cur)
		if cur == root {
			break
		}
		cur = cur.Parent
	}
	// reverse
	for i, j := 0, len(chain)-1; i < j; i, j = i+1, j-1 {
		chain[i], chain[j] = chain[j], chain[i]
	}
	if len(chain) == 0 || chain[0] != root {
		return nil
	}
	return chain
}
