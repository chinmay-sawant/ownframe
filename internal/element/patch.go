package element

// Op is the kind of change detected between two retained trees.
type Op int

const (
	OpCreate Op = iota
	OpUpdate
	OpMove
	OpDelete
)

// String returns the op name.
func (o Op) String() string {
	switch o {
	case OpCreate:
		return "create"
	case OpUpdate:
		return "update"
	case OpMove:
		return "move"
	case OpDelete:
		return "delete"
	default:
		return "unknown"
	}
}

// Patch is an immutable-ish record describing one change.
type Patch struct {
	Op     Op
	Node   *Node
	Before *Node
	Props  []string
}

// PatchList collects patches in deterministic order.
type PatchList []Patch

// HasOp reports whether any patch has the given op.
func (l PatchList) HasOp(op Op) bool {
	for _, p := range l {
		if p.Op == op {
			return true
		}
	}
	return false
}
