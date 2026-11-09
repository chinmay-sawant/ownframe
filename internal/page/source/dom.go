package source

const (
	// commentOpenLen is the length of the "<!--" comment opener.
	commentOpenLen = 4
	// commentCloseLen is the length of the "-->" comment closer.
	commentCloseLen = 3
)

// CheckDOM rejects nesting past maxDepth and element counts past maxNodes.
// It never allocates: tag names compare case-insensitively without
// lowercasing, so a guard run on adversarial input costs no heap.
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
		from, to, close, self, end := scanTag(s, i)
		i = end
		if from == to {
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
		if self || isVoidFold(s, from, to) {
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
// It scans by hand so the guard path allocates nothing.
func SkipMeta(s string, i int) (int, bool) {
	if i+1 >= len(s) {
		return 0, false
	}
	c := s[i+1]
	if c != '!' && c != '?' {
		return 0, false
	}
	if c == '!' && i+commentOpenLen <= len(s) && s[i+2] == '-' && s[i+3] == '-' {
		for k := i; k+2 < len(s); k++ {
			if s[k] == '-' && s[k+1] == '-' && s[k+2] == '>' {
				return k + commentCloseLen, true
			}
		}
		return len(s), true
	}
	for k := i; k < len(s); k++ {
		if s[k] == '>' {
			return k + 1, true
		}
	}
	return len(s), true
}

// scanTag parses one open or close tag at s[i] == '<'. It returns the
// name slice bounds, whether it closes, whether it self-closes, and the
// index after '>'. A missing '>' or an unterminated quote consumes the
// rest as self-closing, matching the previous tag parser verdicts.
func scanTag(s string, i int) (from, to int, close, self bool, end int) {
	j := i + 1
	if j < len(s) && s[j] == '/' {
		close = true
		j++
	}
	from = j
	for j < len(s) && NameByte(s[j]) {
		j++
	}
	to = j
	for j < len(s) && s[j] != '>' {
		q := s[j]
		if q == '"' || q == '\'' {
			k := j + 1
			for k < len(s) && s[k] != q {
				k++
			}
			if k >= len(s) {
				return from, to, close, true, len(s)
			}
			j = k + 1
			continue
		}
		j++
	}
	if j > i+1 && j <= len(s) && j > 0 && s[j-1] == '/' {
		self = true
	}
	if j < len(s) {
		j++
	}
	return from, to, close, self, j
}

// foldEqual reports s[from:to] equal to want ignoring ASCII case.
// Tag names are NameByte bytes, so ASCII folding matches ToLower.
func foldEqual(s string, from, to int, want string) bool {
	if to-from != len(want) {
		return false
	}
	for k := 0; k < len(want); k++ {
		c := s[from+k]
		w := want[k]
		if c == w {
			continue
		}
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		if c != w {
			return false
		}
	}
	return true
}

// isVoidFold reports void elements without allocating a lowercase copy.
func isVoidFold(s string, from, to int) bool {
	return foldEqual(s, from, to, "area") ||
		foldEqual(s, from, to, "base") ||
		foldEqual(s, from, to, "br") ||
		foldEqual(s, from, to, "col") ||
		foldEqual(s, from, to, "embed") ||
		foldEqual(s, from, to, "hr") ||
		foldEqual(s, from, to, "img") ||
		foldEqual(s, from, to, "input") ||
		foldEqual(s, from, to, "link") ||
		foldEqual(s, from, to, "meta") ||
		foldEqual(s, from, to, "source") ||
		foldEqual(s, from, to, "track") ||
		foldEqual(s, from, to, "wbr")
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
