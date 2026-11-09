package draw

import (
	"strings"
	"unicode"
)

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
	global, scopes := HoverPatchRiskyScopes(sources...)
	return global || len(scopes) > 0
}

// HoverPatchRiskyScopes returns any selector scopes that use :has() or
// sibling combinators after :hover. A true global indicates an unscoped
// rule (such as body:has() or :has() at root) that can escape from any
// element on the page. Scoped entries identify classes or IDs (e.g.
// "of-menubar-menu-row" or "of-tooltip--css-hover") whose subtrees alone
// can trigger the escaping behavior.
func HoverPatchRiskyScopes(sources ...string) (global bool, scopes []string) {
	seen := make(map[string]bool)
	for _, source := range sources {
		g, s := hoverSourceRiskyScopes(source)
		if g {
			global = true
		}
		for _, scope := range s {
			if !seen[scope] {
				seen[scope] = true
				scopes = append(scopes, scope)
			}
		}
	}

	return global, scopes
}

// hoverSourceRiskyScopes scans one source for the escaping shapes and their scopes.
func hoverSourceRiskyScopes(source string) (global bool, scopes []string) {
	const (
		normal = iota
		comment
		single
		double
	)

	state := normal
	i := 0
	selStart := 0

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
			case c == '{' || c == '}' || c == ';' || c == ',':
				selStart = i + 1
				i++
			case c == ':' && matchFoldPrefix(source, i+1, "has("):
				prefix := ""
				if i > selStart {
					prefix = source[selStart:i]
				}
				g, scope := classifyRiskyPrefix(prefix)
				if g {
					global = true
				} else if scope != "" {
					scopes = append(scopes, scope)
				}
				i += 1 + len("has(")
			case c == ':' && matchFoldPrefix(source, i+1, "hover"):
				if hoverSelectorEscapes(source, i+1+len("hover")) {
					prefix := ""
					if i > selStart {
						prefix = source[selStart:i]
					}
					g, scope := classifyRiskyPrefix(prefix)
					if g {
						global = true
					} else if scope != "" {
						scopes = append(scopes, scope)
					}
				}
				i += 1 + len("hover")
			default:
				i++
			}
		}
	}

	return global, scopes
}

// classifyRiskyPrefix analyzes the selector prefix before :has( or :hover.
// If the selector targets the whole document (empty, body, html, *), it returns global=true.
// Otherwise, it extracts the class or ID of the scoping element.
func classifyRiskyPrefix(prefix string) (global bool, scope string) {
	trimmed := strings.TrimSpace(prefix)
	if trimmed == "" {
		return true, ""
	}
	lower := strings.ToLower(trimmed)
	if lower == "body" || lower == "html" || lower == "*" || lower == ":root" ||
		strings.HasSuffix(lower, " body") || strings.HasSuffix(lower, " html") ||
		strings.HasSuffix(lower, ">body") || strings.HasSuffix(lower, ">html") ||
		strings.HasSuffix(lower, ">*") || strings.HasSuffix(lower, " *") {
		return true, ""
	}

	// Look for the last class (.name) or ID (#name) in the prefix.
	lastDot := strings.LastIndex(trimmed, ".")
	lastHash := strings.LastIndex(trimmed, "#")
	lastSep := max(lastDot, lastHash)

	if lastSep >= 0 && lastSep+1 < len(trimmed) {
		name := extractIdentifier(trimmed[lastSep+1:])
		if name != "" {
			return false, name
		}
	}

	return true, ""
}

// extractIdentifier extracts a CSS identifier from the start of s.
func extractIdentifier(s string) string {
	end := 0
	for end < len(s) {
		r := rune(s[end])
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-' || r == '_' {
			end++
		} else {
			break
		}
	}
	return s[:end]
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
