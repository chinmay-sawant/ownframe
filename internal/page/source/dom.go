package source

import "strings"

// quotePair is the opening plus closing quote a skip jumps over.
const quotePair = 2

// CheckDOM rejects nesting past maxDepth and element counts past maxNodes.
func CheckDOM(s string, maxDepth, maxNodes int) error {
	depth, nodes := 0, 0
	i := 0
	for i < len(s) {
		if s[i] != '<' || i+1 >= len(s) {
			i++
			continue
		}
		if end, ok := SkipMeta(s, i); ok {
			i = end
			continue
		}
		name, close, self, end := tag(s, i)
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
		if nodes > maxNodes {
			return ErrTooManyNodes
		}
		if self || Void(name) {
			continue
		}
		depth++
		if depth > maxDepth {
			return ErrDOMTooDeep
		}
	}
	return nil
}

// SkipMeta skips a comment or a <! / <? declaration at s[i] == '<'.
func SkipMeta(s string, i int) (int, bool) {
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

// tag parses one open or close tag at s[i] == '<'. It returns the
// lowercase name, whether it closes, whether it self-closes, and the
// index after '>'. A missing '>' consumes the rest.
func tag(s string, i int) (string, bool, bool, int) {
	j := i + 1
	close := false
	if j < len(s) && s[j] == '/' {
		close = true
		j++
	}
	from := j
	for j < len(s) && NameByte(s[j]) {
		j++
	}
	name := strings.ToLower(s[from:j])
	for j < len(s) && s[j] != '>' {
		if s[j] == '"' || s[j] == '\'' {
			if end := strings.IndexByte(s[j+1:], s[j]); end >= 0 {
				j += end + quotePair
				continue
			}
			return name, close, true, len(s)
		}
		j++
	}
	self := j > i+1 && s[j-1] == '/'
	if j < len(s) {
		j++
	}
	return name, close, self, j
}

// NameByte reports a byte that can appear in a tag name.
func NameByte(c byte) bool {
	switch {
	case c == '_' || c == '-' || c == ':':
		return true
	case c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z':
		return true
	case c >= '0' && c <= '9':
		return true
	default:
		return false
	}
}

// Void reports elements that never push depth.
func Void(name string) bool {
	switch name {
	case "area", "base", "br", "col", "embed", "hr", "img", "input", "link", "meta", "source", "track", "wbr":
		return true
	default:
		return false
	}
}
