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
	// Demo returns the HTML that shows this example working. The page draws it
	// directly under the example's row, so the status, the note and the
	// showcase read as one block. It holds what the example's boxes look like
	// and the host's readout for those boxes, and nothing about other rows.
	Demo func() string
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
	return []Tab{navTab(), shapesTab(), motionTab(), fragTab(), readTab(), svgTab()}
}
