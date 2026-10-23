package a11y

import (
	"strings"

	"github.com/chinmay-sawant/blinkless/markup"
)

// Hidden reports that n is omitted from the accessibility tree.
func Hidden(n *markup.Node) bool {
	if HasAttr(n, "hidden") || strings.EqualFold(n.Attrs["aria-hidden"], "true") {
		return true
	}
	switch n.Name {
	case "head", "script", "style", "template":
		return true
	}
	return false
}

// HasAttr reports that n carries name.
func HasAttr(n *markup.Node, name string) bool {
	_, ok := n.Attrs[name]
	return ok
}
