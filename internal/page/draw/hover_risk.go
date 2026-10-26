package draw

// HoverPatchRisky reports whether the stylesheet text can style an element
// outside the hovered elements' subtrees when the hover changes, which is
// the case the scoped hover patch cannot represent. Two selector shapes
// escape the subtree: :has(), whose argument can match any element relative
// to the styled one, and a sibling combinator after :hover, which styles a
// sibling the patch roots do not contain. Either shape anywhere in the page
// or theme source disables the patch and the caller lays the whole page out
// again, so a risky page draws exactly what it drew before.
//
// The scan is textual and conservative: comments and quoted strings never
// flag, but an escaped spelling of :hover or :has still does. A page that
// only shows those spellings as text (a CSS reference page, for example)
// keeps the full layout and stays correct.
func HoverPatchRisky(sources ...string) bool {
	for _, source := range sources {
		if hoverSourceRisky(source) {
			return true
		}
	}

	return false
}

// hoverSourceRisky scans one source for the escaping shapes.
func hoverSourceRisky(source string) bool {
	const (
		normal = iota
		comment
		single
		double
	)

	state := normal
	i := 0

	for i < len(source) {
		c := source[i]

		switch state {
		case comment:
			switch {
			case c == '*' && i+1 < len(source) && source[i+1] == '/':
				state = normal
				i += 2
			default:
				i++
			}
		case single:
			switch c {
			case '\\':
				i += 2
			case '\'':
				state = normal
				i++
			default:
				i++
			}
		case double:
			switch c {
			case '\\':
				i += 2
			case '"':
				state = normal
				i++
			default:
				i++
			}
		default:
			switch {
			case c == '/' && i+1 < len(source) && source[i+1] == '*':
				state = comment
				i += 2
			case c == '\'':
				state = single
				i++
			case c == '"':
				state = double
				i++
			case c == ':' && matchFoldPrefix(source, i+1, "has("):
				return true
			case c == ':' && matchFoldPrefix(source, i+1, "hover"):
				if hoverSelectorEscapes(source, i+1+len("hover")) {
					return true
				}

				i += 1 + len("hover")
			default:
				i++
			}
		}
	}

	return false
}

// hoverSelectorEscapes reports whether the selector after a :hover match
// reaches a sibling before the selector ends. The selector ends at an opening
// brace or a comma; a + or ~ first means the rule can style an element the
// hovered subtrees do not contain.
func hoverSelectorEscapes(source string, i int) bool {
	for i < len(source) {
		switch source[i] {
		case '{', ',', '}':
			return false
		case '+', '~':
			return true
		default:
			i++
		}
	}

	return false
}

// matchFoldPrefix reports whether s[i:] starts with prefix, folding ASCII
// upper case to lower case.
func matchFoldPrefix(s string, i int, prefix string) bool {
	if i < 0 || len(s)-i < len(prefix) {
		return false
	}

	for j := 0; j < len(prefix); j++ {
		c := s[i+j]
		if 'A' <= c && c <= 'Z' {
			c += 'a' - 'A'
		}

		if c != prefix[j] {
			return false
		}
	}

	return true
}
