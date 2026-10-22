package form

// Order is the tab order: document order of control ids, with disabled
// controls and tabindex="-1" left out. Duplicate ids are skipped.
func Order(ids []string, lookup func(id string) (Control, bool)) []string {
	seen := map[string]bool{}
	out := []string{}
	for _, id := range ids {
		if id == "" || seen[id] {
			continue
		}

		c, ok := lookup(id)
		if !ok || c.Disabled || c.TabIndex < 0 {
			continue
		}

		seen[id] = true
		out = append(out, id)
	}

	return out
}
