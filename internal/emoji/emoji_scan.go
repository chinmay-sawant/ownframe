package emoji

import "strings"

// tagStart reports whether html[i] opens a tag, a closing tag, or a
// declaration. A bare '<' followed by anything else is text.
func tagStart(html string, i int) bool {
	j := i + 1
	if j < len(html) && (html[j] == '/' || html[j] == '!' || html[j] == '?') {
		return true
	}

	return j < len(html) && (html[j] == '_' || html[j] == ':' ||
		'a' <= html[j] && html[j] <= 'z' || 'A' <= html[j] && html[j] <= 'Z')
}

// tagName returns the lowercase element name and the index after the tag's
// closing '>', skipping quoted attribute values. A missing '>' consumes
// the rest of the input.
func tagName(html string, i int) (string, int) {
	j := i + 1
	if j < len(html) && html[j] == '/' {
		j++
	}

	start := j
	for j < len(html) && isNameByte(html[j]) {
		j++
	}

	name := strings.ToLower(html[start:j])

	for j < len(html) && html[j] != '>' {
		if html[j] == '"' || html[j] == '\'' {
			quote := html[j]
			j++
			for j < len(html) && html[j] != quote {
				j++
			}
		}

		if j < len(html) {
			j++
		}
	}

	if j < len(html) {
		j++
	}

	return name, j
}

func isNameByte(c byte) bool {
	return c == '_' || c == ':' || c == '-' ||
		'a' <= c && c <= 'z' || 'A' <= c && c <= 'Z' || '0' <= c && c <= '9'
}

// rawTextTag reports whether the element's content is not markup: script and
// style hold raw text, textarea and title hold RCDATA. Rewriting any of them
// would put literal <img> source into what the page shows or runs.
func rawTextTag(name string) bool {
	switch name {
	case "script", "style", "textarea", "title":
		return true
	}

	return false
}

// rawText returns the script or style body through its closing tag. ok is
// false when the body never closes.
func rawText(html string, i int, name string) (string, bool) {
	close := "</" + name
	for k := i; k < len(html); k++ {
		if html[k] != '<' || len(html)-k < len(close) || !strings.EqualFold(html[k:k+len(close)], close) {
			continue
		}

		_, end := tagName(html, k)

		return html[i:end], true
	}

	return html[i:], false
}
