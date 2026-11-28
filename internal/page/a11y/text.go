package a11y

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/chinmay-sawant/blinkless/markup"
)

// collapseSpaces trims leading and trailing Unicode whitespace and replaces
// internal whitespace runs with a single ASCII space, matching
// strings.Join(strings.Fields(s), " ") without allocating a token slice.
func collapseSpaces(s string) string {
	// Fast path: check if already normalized (no leading/trailing space, no
	// non-ASCII-space whitespace, no consecutive spaces).
	n := len(s)
	if n == 0 {
		return ""
	}
	fast := true
	prevSpace := true
	for i := 0; i < n; i++ {
		b := s[i]
		if b >= utf8.RuneSelf {
			fast = false
			break
		}
		switch b {
		case ' ':
			if prevSpace {
				fast = false
			} else {
				prevSpace = true
			}
		case '\t', '\n', '\r', '\v', '\f':
			fast = false
		default:
			prevSpace = false
		}
		if !fast {
			break
		}
	}
	if fast && !prevSpace {
		return s
	}

	var b strings.Builder
	inSpace := false
	wrote := false
	for _, r := range s {
		if unicode.IsSpace(r) {
			if wrote {
				inSpace = true
			}
			continue
		}
		if inSpace {
			b.WriteByte(' ')
			inSpace = false
		}
		b.WriteRune(r)
		wrote = true
	}
	return b.String()
}

// NodeText is the flattened descendant text of n.
func NodeText(n *markup.Node) string {
	var out strings.Builder
	var walk func(*markup.Node)
	walk = func(node *markup.Node) {
		if node.Type == markup.TypeText {
			out.WriteString(node.Text)
			return
		}
		if node.Name == "script" || node.Name == "style" {
			return
		}
		for _, child := range node.Children {
			walk(child)
		}
	}
	walk(n)
	return collapseSpaces(out.String())
}
