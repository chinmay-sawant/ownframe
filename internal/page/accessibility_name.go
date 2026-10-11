package page

import (
	"strings"

	"github.com/chinmay-sawant/blinkless/markup"
	"github.com/chinmay-sawant/ownframe/internal/host"
)

func (b *accessibilityBuilder) accessibilityName(n *markup.Node, role host.AccessibilityRole) string {
	a := n.Attrs
	if v := b.referencedText(a["aria-labelledby"]); v != "" {
		return v
	}
	if v := strings.TrimSpace(a["aria-label"]); v != "" {
		return v
	}
	if v := b.labelText(n); v != "" {
		return v
	}
	if role == host.AccessibilityImage {
		if v := strings.TrimSpace(a["alt"]); v != "" {
			return v
		}
	}
	switch role {
	case host.AccessibilityButton, host.AccessibilityLink, host.AccessibilityHeading, host.AccessibilityParagraph, host.AccessibilityListItem:
		if v := nodeText(n); v != "" {
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

func (b *accessibilityBuilder) accessibilityDescription(n *markup.Node) string {
	return b.referencedText(n.Attrs["aria-describedby"])
}

func (b *accessibilityBuilder) labelText(n *markup.Node) string {
	id := n.Attrs["id"]
	for _, label := range b.labels[id] {
		if v := nodeText(label); v != "" {
			return v
		}
	}
	for parent := b.parents[n]; parent != nil; parent = b.parents[parent] {
		if parent.Name == "label" {
			return nodeText(parent)
		}
	}
	return ""
}

func (b *accessibilityBuilder) referencedText(raw string) string {
	var parts []string
	for _, id := range strings.Fields(raw) {
		if nodes := b.ids[id]; len(nodes) == 1 {
			if v := nodeText(nodes[0]); v != "" {
				parts = append(parts, v)
			}
		}
	}
	return strings.Join(parts, " ")
}
