package form

var emptyOrder = []string{}

// Merge keeps the user's values across a redraw. The caller clamps
// the caret and the anchor after the merge.
func Merge(prev map[string]Control, spans []Span) (map[string]Control, []string) {
	if len(spans) == 0 {
		if prev != nil && len(prev) == 0 {
			return prev, emptyOrder
		}
		return map[string]Control{}, emptyOrder
	}
	out := make(map[string]Control, len(spans))
	order := make([]string, 0, len(spans))
	for _, sp := range spans {
		id := sp.Control.ID
		if id == "" {
			continue
		}
		if _, seen := out[id]; seen {
			continue
		}

		order = append(order, id)
		cur := sp.Control
		if old, ok := prev[id]; ok {
			cur = keepUser(cur, old)
		}
		out[id] = cur
	}

	return out, order
}

func keepUser(cur, old Control) Control {
	if old.Tag != cur.Tag || old.Type != cur.Type {
		return cur
	}
	if cur.Tag == "select" {
		return keepSelect(cur, old.Value)
	}
	if cur.Type == "checkbox" || cur.Type == "radio" {
		cur.Checked = old.Checked
		return cur
	}
	if TextLike(cur.Type) {
		cur.Value = old.Value
	}

	return cur
}

func keepSelect(cur Control, value string) Control {
	hit := -1
	for i := range cur.Options {
		if cur.Options[i].Value == value {
			hit = i
			break
		}
	}
	if hit < 0 {
		return cur
	}

	opts := make([]Option, len(cur.Options))
	copy(opts, cur.Options)
	for i := range opts {
		opts[i].Selected = i == hit
	}
	cur.Options = opts
	cur.Value = value

	return cur
}

// ApplyBinds copies each bound control through read.
// A control with an empty bind, or a read that reports false, stays as it is.
// A nil map does nothing.
func ApplyBinds(byID map[string]Control, read func(Control) (Control, bool)) {
	for id, ctrl := range byID {
		if ctrl.Bind == "" {
			continue
		}

		if c, ok := read(ctrl); ok {
			byID[id] = c
		}
	}
}
