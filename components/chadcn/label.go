package chadcn

import "github.com/chinmay-sawant/ownframe/ui"

// LabelOptions describes the control a label is bound to.
type LabelOptions struct {
	For string
}

// Label names a form control. For is written as the for attribute and should
// match the control ID.
func Label(text string, options ...LabelOptions) *ui.Element {
	label := ui.Tag("label").Class("of-label").Text(text)
	if len(options) > 0 && options[0].For != "" {
		label.Attr("for", options[0].For)
	}
	return label
}
