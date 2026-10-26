package main

import (
	"strconv"
	"strings"
	"unicode/utf8"
)

const (
	previewEntityBytes = 12
	previewEntityBase  = 10
	previewEntityHex   = 16
	previewEntityBits  = 32
)

// stripHTML removes comments and tags, turns <br> into a newline, and keeps
// the alt text of an image. Entities are decoded later by previewInline.
func stripHTML(text string) string {
	if !strings.ContainsRune(text, '<') {
		return text
	}

	var b strings.Builder

	b.Grow(len(text))

	for i := 0; i < len(text); {
		if strings.HasPrefix(text[i:], "<!--") {
			end := strings.Index(text[i+len("<!--"):], "-->")
			if end < 0 {
				break
			}

			i += len("<!--") + end + len("-->")

			continue
		}

		if text[i] != '<' {
			b.WriteByte(text[i])
			i++

			continue
		}

		end, name, rest, closing, ok := htmlTagAt(text, i)
		if !ok {
			b.WriteByte('<')
			i++

			continue
		}

		switch {
		case name == "br":
			b.WriteByte('\n')
		case name == "hr":
			b.WriteByte('\n')
		case name == "img" && !closing:
			b.WriteString(htmlAttr(rest, "alt"))
		case closing && isHTMLBlockTag(name):
			b.WriteByte('\n')
		}

		i = end
	}

	return b.String()
}

// htmlTagAt parses the tag beginning at text[i].
func htmlTagAt(text string, i int) (end int, name, rest string, closing, ok bool) {
	closeAt := strings.IndexByte(text[i:], '>')
	if closeAt < 0 {
		return 0, "", "", false, false
	}

	end = i + closeAt + 1
	content := text[i+1 : end-1]

	name, rest, closing, ok = htmlTagHead(content)

	return end, name, rest, closing, ok
}

// htmlTagHead splits a tag body into a lower-case name and its attributes.
func htmlTagHead(content string) (name, rest string, closing, ok bool) {
	if strings.HasPrefix(content, "/") {
		closing = true
		content = content[1:]
	}

	if content == "" || !isHTMLNameStart(content[0]) {
		return "", "", false, false
	}

	j := 1
	for j < len(content) && isHTMLNameChar(content[j]) {
		j++
	}

	rest = content[j:]
	if rest != "" && !isHTMLSpace(rest[0]) && rest[0] != '/' {
		return "", "", false, false
	}

	return strings.ToLower(content[:j]), rest, closing, true
}

// isHTMLBlockStart reports whether a line starts a block-level HTML tag.
func isHTMLBlockStart(line string) bool {
	trimmed := strings.TrimSpace(line)
	if !strings.HasPrefix(trimmed, "<") {
		return false
	}

	closeAt := strings.IndexByte(trimmed, '>')
	if closeAt < 0 {
		return false
	}

	name, _, _, ok := htmlTagHead(trimmed[1:closeAt])
	if !ok {
		return false
	}

	return isHTMLBlockTag(name)
}

// isHTMLBlockTag reports whether a tag is rendered as a paragraph block.
func isHTMLBlockTag(name string) bool {
	switch name {
	case "p", "div", "details", "summary", "section", "article", "aside",
		"header", "footer", "main", "nav", "ul", "ol", "li", "blockquote",
		"pre", "table", "thead", "tbody", "tr", "td", "th", "figure",
		"figcaption", "br", "hr":
		return true
	}

	return false
}

// htmlTextLines turns a block of raw HTML into readable paragraphs.
func htmlTextLines(raw string) []string {
	plain := stripHTML(raw)
	lines := strings.Split(plain, "\n")
	out := make([]string, 0, len(lines))

	for _, line := range lines {
		line = strings.Join(strings.Fields(line), " ")
		if line != "" {
			out = append(out, line)
		}
	}

	return out
}

// htmlAttr returns one attribute value from a tag body.
func htmlAttr(rest, want string) string {
	lower := strings.ToLower(rest)

	for pos := 0; pos < len(lower); {
		idx := strings.Index(lower[pos:], want)
		if idx < 0 {
			return ""
		}

		idx += pos
		after := idx + len(want)

		if (idx > 0 && isHTMLNameChar(lower[idx-1])) || (after < len(lower) && isHTMLNameChar(lower[after])) {
			pos = after

			continue
		}

		for after < len(lower) && isHTMLSpace(lower[after]) {
			after++
		}

		if after >= len(lower) || lower[after] != '=' {
			pos = after

			continue
		}

		after++
		for after < len(lower) && isHTMLSpace(lower[after]) {
			after++
		}

		if after >= len(lower) {
			return ""
		}

		if quote := lower[after]; quote == '"' || quote == '\'' {
			end := strings.IndexByte(rest[after+1:], quote)
			if end < 0 {
				return ""
			}

			return rest[after+1 : after+1+end]
		}

		end := after
		for end < len(rest) && !isHTMLSpace(rest[end]) && rest[end] != '>' {
			end++
		}

		return rest[after:end]
	}

	return ""
}

// isHTMLNameStart reports whether c can start an HTML tag name.
func isHTMLNameStart(c byte) bool {
	return c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z'
}

// isHTMLNameChar reports whether c can continue an HTML tag or attribute name.
func isHTMLNameChar(c byte) bool {
	return isHTMLNameStart(c) || c >= '0' && c <= '9' || c == '-' || c == '_'
}

// isHTMLSpace reports whether c is HTML whitespace.
func isHTMLSpace(c byte) bool {
	return c == ' ' || c == '\t' || c == '\n' || c == '\r'
}

// decodeHTMLEntities turns the basic named and numeric entities into text.
func decodeHTMLEntities(text string) string {
	if !strings.ContainsRune(text, '&') {
		return text
	}

	var b strings.Builder

	b.Grow(len(text))

	for i := 0; i < len(text); {
		if text[i] != '&' {
			b.WriteByte(text[i])
			i++

			continue
		}

		end := strings.IndexByte(text[i:], ';')
		if end < 0 || end > previewEntityBytes {
			b.WriteByte('&')
			i++

			continue
		}

		value, ok := htmlEntityValue(text[i+1 : i+end])
		if !ok {
			b.WriteByte('&')
			i++

			continue
		}

		b.WriteString(value)
		i += end + 1
	}

	return b.String()
}

// htmlEntityValue returns the replacement for one entity name.
func htmlEntityValue(name string) (string, bool) {
	switch name {
	case "amp":
		return "&", true
	case "lt":
		return "<", true
	case "gt":
		return ">", true
	case "quot":
		return `"`, true
	case "apos":
		return "'", true
	case "nbsp":
		return " ", true
	case "ndash":
		return "–", true
	case "mdash":
		return "—", true
	case "hellip":
		return "…", true
	}

	if !strings.HasPrefix(name, "#") {
		return "", false
	}

	digits := name[1:]
	base := previewEntityBase

	if strings.HasPrefix(digits, "x") || strings.HasPrefix(digits, "X") {
		digits = digits[1:]
		base = previewEntityHex
	}

	if digits == "" {
		return "", false
	}

	value, err := strconv.ParseInt(digits, base, previewEntityBits)
	if err != nil || value <= 0 || !utf8.ValidRune(rune(value)) {
		return "", false
	}

	return string(rune(value)), true
}
