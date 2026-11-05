package form

// NextOption selects the next enabled option, wrapping.
// A disabled option is never the landed value. The control is unchanged
// when every option is disabled.
func NextOption(c Control) Control {
	n := len(c.Options)
	if n == 0 {
		return c
	}

	start := -1
	for i := range c.Options {
		if c.Options[i].Selected {
			start = i
			break
		}
	}

	next := enabledOptionAfter(c.Options, start)
	if next < 0 {
		return c
	}

	for i := range c.Options {
		c.Options[i].Selected = i == next
	}

	c.Value = c.Options[next].Value

	return c
}

// enabledOptionAfter returns the next option that is not disabled.
// A negative start searches from the first option. It returns -1 when
// every option is disabled.
func enabledOptionAfter(opts []Option, start int) int {
	n := len(opts)
	if n == 0 {
		return -1
	}

	begin := 0
	if start >= 0 {
		begin = start + 1
	}
	for step := 0; step < n; step++ {
		i := (begin + step) % n
		if opts[i].Disabled {
			continue
		}

		return i
	}

	return -1
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

// PickOption marks one enabled option and returns its value.
// A selected enabled option wins. A selected disabled option lands on the
// next enabled option. When every option is disabled, the selected option,
// or the first option, stays.
func PickOption(opts []Option) string {
	if len(opts) == 0 {
		return ""
	}

	start := -1
	chosen := -1
	for i := range opts {
		if !opts[i].Selected {
			continue
		}
		start = i
		if !opts[i].Disabled {
			chosen = i
		}
		break
	}
	if chosen < 0 {
		chosen = enabledOptionAfter(opts, start)
	}
	if chosen < 0 {
		chosen = 0
		if start >= 0 {
			chosen = start
		}
	}
	for i := range opts {
		opts[i].Selected = i == chosen
	}

	return opts[chosen].Value
}
