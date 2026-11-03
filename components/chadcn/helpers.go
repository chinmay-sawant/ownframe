package chadcn

import (
	"strconv"
	"strings"

	"github.com/chinmay-sawant/ownframe/ui"
)

// classes joins non-empty class names for one class attribute.
func classes(names ...string) string {
	kept := names[:0]
	for _, name := range names {
		if name != "" {
			kept = append(kept, name)
		}
	}
	return strings.Join(kept, " ")
}

// partID derives a stable part ID from a host ID. It returns an empty string
// when the host did not supply an ID, so callers never invent one.
func partID(id, suffix string) string {
	if id == "" {
		return ""
	}
	return id + suffix
}

func ariaBool(value bool) string { return strconv.FormatBool(value) }

// setDisabled writes the disabled attribute and returns the element.
func setDisabled(element *ui.Element, disabled bool) *ui.Element {
	if disabled {
		return element.Attr("disabled", "disabled")
	}
	return element
}
