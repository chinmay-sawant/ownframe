package a11y

import (
	"strings"

	"github.com/chinmay-sawant/blinkless/markup"
	"github.com/chinmay-sawant/ownframe/internal/host"
)

// Name is the computed accessible name of n for role.
func (b Names) Name(n *markup.Node, role host.AccessibilityRole) string {
	a := n.Attrs
	if v := b.ReferencedText(a["aria-labelledby"]); v != "" {
		return v
	}
	if v := strings.TrimSpace(a["aria-label"]); v != "" {
		return v
	}
	if v := b.LabelText(n); v != "" {
		return v
	}
	if role == host.AccessibilityImage {
		if v := strings.TrimSpace(a["alt"]); v != "" {
			return v
		}
	}
	switch role {
	case host.AccessibilityButton, host.AccessibilityLink, host.AccessibilityHeading, host.AccessibilityParagraph, host.AccessibilityListItem:
		if v := NodeText(n); v != "" {
			return v
		}
	}
	if role == host.AccessibilityTextbox {
		if v := strings.TrimSpace(a["placeholder"]); v != "" {
			return v
		}
	}
	return strings.TrimSpace(a["title"])
}

// Description is the aria-describedby text of n.
func (b Names) Description(n *markup.Node) string {
	return b.ReferencedText(n.Attrs["aria-describedby"])
}
