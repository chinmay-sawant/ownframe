package page

import (
	"errors"
	"strings"
)

// MaxDOMDepth caps parse-time nesting (GEN-29): layout recursion past
// it is a fatal stack overflow that recover cannot catch.
const MaxDOMDepth = 256

// MaxDOMNodes caps the parse-time element count.
const MaxDOMNodes = 20000

// ErrDOMTooDeep means the template nests past MaxDOMDepth.
var ErrDOMTooDeep = errors.New("ownframe: template nests too deep")

// ErrTooManyNodes means the template holds past MaxDOMNodes elements.
var ErrTooManyNodes = errors.New("ownframe: template has too many elements")

// isVoidTag reports void elements, which never push depth.
func isVoidTag(name string) bool {
	switch name {
	case "area", "base", "br", "col", "embed", "hr", "img", "input", "link", "meta", "source", "track", "wbr":
		return true
	default:
		return false
	}
}

// checkDOMLimits rejects nesting and element counts past the caps.
func checkDOMLimits(source string) error {
	depth, nodes := 0, 0
	i := 0
	for i < len(source) {
		if source[i] != '<' || i+1 >= len(source) {
			i++
			continue
		}
		if end, ok := domComment(source, i); ok {
			i = end
			continue
		}
		name, close, self, end := domTag(source, i)
		i = end
		if name == "" {
			continue
		}
		if close {
			if depth > 0 {
				depth--
			}
			continue
		}
		nodes++
		if nodes > MaxDOMNodes {
			return ErrTooManyNodes
		}
		if self || isVoidTag(name) {
			continue
		}
		depth++
		if depth > MaxDOMDepth {
			return ErrDOMTooDeep
		}
	}

	return nil
}

// domComment skips comments and <! / <? declarations.
func domComment(s string, i int) (int, bool) {
	if s[i+1] == '!' || s[i+1] == '?' {
		if strings.HasPrefix(s[i:], "<!--") {
			if end := strings.Index(s[i:], "-->"); end >= 0 {
				return i + end + len("-->"), true
			}
			return len(s), true
		}
		if end := strings.IndexByte(s[i:], '>'); end >= 0 {
			return i + end + 1, true
		}
		return len(s), true
	}

	return 0, false
}
