package form

// NextOption selects the option after the current one, wrapping.
func NextOption(c Control) Control {
	n := len(c.Options)
	if n == 0 {
		return c
	}

	next := 0
	for i, opt := range c.Options {
		if opt.Selected {
			next = i + 1
			break
		}
	}

	if next == n {
		next = 0
	}

	for i := range c.Options {
		c.Options[i].Selected = i == next
	}

	c.Value = c.Options[next].Value

	return c
}

// SelectedIndex is the selected option index, or -1.
func SelectedIndex(opts []Option) int {
	for i := range opts {
		if opts[i].Selected {
			return i
		}
	}

	return -1
}

// PickOption marks one selected option and returns its value.
func PickOption(opts []Option) string {
	if len(opts) == 0 {
		return ""
	}

	chosen := 0
	for i := range opts {
		if opts[i].Selected {
			chosen = i
			break
		}
	}
	for i := range opts {
		opts[i].Selected = i == chosen
	}

	return opts[chosen].Value
}
