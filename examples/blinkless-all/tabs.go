package main

// Prop is one example on a tab. Name is the property the example sets. Covers
// lists the other properties the same example also exercises, so a family
// such as the six offset-* longhands can be shown by one box and still count
// as six properties.
type Prop struct {
	Name   string
	Covers []string
	// Note says in one plain sentence what the host or engine does with it.
	Note string
	// Description says what the property does and which values it accepts.
	Description string
	// Validate says how to check the example by eye and by readout.
	Validate string
	// ValidateFn computes Validate lazily when set, so generated rows only lay
	// out their demos when their tab is rendered.
	ValidateFn func() string
	// Demo returns the HTML that shows this example working. The page draws it
	// directly under the example's row, so the status, the note and the
	// showcase read as one block. It holds what the example's boxes look like
	// and the host's readout for those boxes, and nothing about other rows.
	Demo func() string
}

func (p Prop) validateText() string {
	if p.ValidateFn != nil {
		return p.ValidateFn()
	}

	return p.Validate
}

// Names returns Name followed by Covers.
func (p Prop) Names() []string {
	return append([]string{p.Name}, p.Covers...)
}

// Tab is one group of properties with the page section that shows them.
type Tab struct {
	ID    string
	Title string
	// Props lists the examples on this tab. Every property of the group
	// appears once, either as a Name or inside a Covers list.
	Props []Prop
	// Head returns HTML shared by the rows of this tab, such as one <style>
	// element. It is drawn once, above the first row. It may be nil.
	Head func() string
	// Prepare runs lazily before the tab's rows are rendered so generated tabs
	// compute their readouts on demand instead of at startup. It may be nil.
	Prepare func()
}

// Count is the number of properties the tab covers.
func (t Tab) Count() int {
	n := 0
	for _, p := range t.Props {
		n += len(p.Names())
	}

	return n
}

// Names returns every covered property in order.
func (t Tab) Names() []string {
	var out []string
	for _, p := range t.Props {
		out = append(out, p.Names()...)
	}

	return out
}

// allTabs is the tab order. Each tab lives in its own tab_<id>.go file.
func allTabs() []Tab {
	tabs := []Tab{navTab(), shapesTab(), motionTab(), fragTab(), readTab(), svgTab()}

	taken := map[string]bool{}
	for _, t := range tabs {
		for _, n := range t.Names() {
			taken[n] = true
		}
	}

	tabs = append(tabs, genTabs(taken)...)
	fillFields(tabs)

	return tabs
}

// fillFields gives every example a description and a validation line. The
// hand-built tabs keep the text they carry; the others get the accepted
// values from the compatibility matrix and a generic check.
func fillFields(tabs []Tab) {
	accepted := map[string]string{}
	for _, g := range genProps() {
		accepted[g.Name] = g.Accepted
	}

	for ti := range tabs {
		for pi := range tabs[ti].Props {
			p := &tabs[ti].Props[pi]

			if p.Description == "" {
				if a, ok := accepted[p.Name]; ok {
					p.Description = p.Name + ". Accepted values: " + a + "."
				} else {
					p.Description = p.Name + "."
				}
			}

			if p.Validate == "" && p.ValidateFn == nil {
				p.Validate = "Read the picture and the readout under the row. They must agree with the note, and the status badge must read implemented."
			}
		}
	}
}
