package chadcn

import "github.com/chinmay-sawant/ownframe/ui"

// SelectValue is one choice in a NativeSelect. The zero value is an option with
// an empty value and an empty label.
type SelectValue struct {
	// Value is the submitted value.
	Value string
	// Label is the visible text. An empty label falls back to Value.
	Label string
	// Disabled keeps the option visible but unselectable.
	Disabled bool
}

// NativeSelect builds the platform select control with one option per value and
// marks the option that matches selected. Add an entry with an empty Value for
// a placeholder choice. The page keeps the live value on Page.Form, so the host
// reads it back and passes it as selected on the next rebuild.
func NativeSelect(id string, values []SelectValue, selected string) *ui.Element {
	selectElement := ui.Tag("select").ID(id).
		Class(classes("of-native-select", "of-focus")).
		Children(nativeOptions(values, selected)...)
	return ui.Div().Class(classes("of-scope", "of-native-select-wrap")).
		Children(selectElement, nativeSelectIcon())
}

func nativeOptions(values []SelectValue, selected string) []*ui.Element {
	options := make([]*ui.Element, 0, len(values))
	for _, value := range values {
		label := value.Label
		if label == "" {
			label = value.Value
		}
		option := ui.Tag("option").Attr("value", value.Value).Text(label)
		if value.Value == selected {
			option.Attr("selected", "selected")
		}
		if value.Disabled {
			option.Attr("disabled", "disabled")
		}
		options = append(options, option)
	}
	return options
}

// nativeSelectIcon draws the chevron in CSS so the control does not depend on
// a glyph in the font. It is decorative, so it stays out of the accessibility
// tree.
func nativeSelectIcon() *ui.Element {
	return ui.Span().Class("of-native-select-icon").Attr("aria-hidden", "true")
}
