package ui

import "strings"

// canonicalTag returns the interned constant for an already-normalized tag
// name, or "" when the name is not in the fixed set the builder supports.
// The set matches supportedTag in builder.go, so every tag Render accepts
// resolves here. The switch holds only string constants: no map, no package
// state, nothing to mutate or guard.
func canonicalTag(normalized string) string {
	switch normalized {
	case "a":
		return "a"
	case "article":
		return "article"
	case "aside":
		return "aside"
	case "b":
		return "b"
	case "blockquote":
		return "blockquote"
	case "br":
		return "br"
	case "button":
		return "button"
	case "caption":
		return "caption"
	case "code":
		return "code"
	case "div":
		return "div"
	case "em":
		return "em"
	case "footer":
		return "footer"
	case "form":
		return "form"
	case "h1":
		return "h1"
	case "h2":
		return "h2"
	case "h3":
		return "h3"
	case "h4":
		return "h4"
	case "h5":
		return "h5"
	case "h6":
		return "h6"
	case "header":
		return "header"
	case "hr":
		return "hr"
	case "i":
		return "i"
	case "img":
		return "img"
	case "input":
		return "input"
	case "label":
		return "label"
	case "li":
		return "li"
	case "main":
		return "main"
	case "nav":
		return "nav"
	case "ol":
		return "ol"
	case "optgroup":
		return "optgroup"
	case "option":
		return "option"
	case "p":
		return "p"
	case "pre":
		return "pre"
	case "section":
		return "section"
	case "select":
		return "select"
	case "small":
		return "small"
	case "span":
		return "span"
	case "strong":
		return "strong"
	case "table":
		return "table"
	case "tbody":
		return "tbody"
	case "td":
		return "td"
	case "textarea":
		return "textarea"
	case "tfoot":
		return "tfoot"
	case "th":
		return "th"
	case "thead":
		return "thead"
	case "tr":
		return "tr"
	case "u":
		return "u"
	case "ul":
		return "ul"
	default:
		return ""
	}
}

// internTag normalizes a tag name exactly like
// strings.ToLower(strings.TrimSpace(name)) while returning the interned
// constant for known tags without allocating. Exact lowercase hits resolve
// on the first switch; whitespace-padded hits resolve after TrimSpace,
// which returns a substring without allocating; only inputs with uppercase
// letters reach ToLower, and known tags among those still return the shared
// constant so the one transient copy is discarded. Unknown, empty, and
// Unicode inputs fall through to the same ToLower(TrimSpace(name)) result
// as before, so the output is byte-identical for every input.
func internTag(name string) string {
	if canonical := canonicalTag(name); canonical != "" {
		return canonical
	}
	trimmed := strings.TrimSpace(name)
	if trimmed != name {
		if canonical := canonicalTag(trimmed); canonical != "" {
			return canonical
		}
	}
	lowered := strings.ToLower(trimmed)
	if canonical := canonicalTag(lowered); canonical != "" {
		return canonical
	}
	return lowered
}
