package form

import "strings"

// Scan reads input, textarea, select, and button controls from source.
// A control with an empty id is skipped. Script, style, and comments are not.
func Scan(source string) []Span {
	if !strings.Contains(source, "<input") && !strings.Contains(source, "<INPUT") &&
		!strings.Contains(source, "<textarea") && !strings.Contains(source, "<TEXTAREA") &&
		!strings.Contains(source, "<select") && !strings.Contains(source, "<SELECT") &&
		!strings.Contains(source, "<button") && !strings.Contains(source, "<BUTTON") {
		return nil
	}
	var out []Span
	i := 0
	for i < len(source) {
		next := strings.IndexByte(source[i:], '<')
		if next < 0 {
			break
		}
		i += next
		if strings.HasPrefix(source[i:], "<!--") {
			end := strings.Index(source[i+CommentOpenLen:], "-->")
			if end < 0 {
				break
			}
			i += end + CommentTotalLen
			continue
		}

		name, raw, end, ok := ReadOpen(source, i)
		if !ok {
			i++
			continue
		}

		low := strings.ToLower(name)
		switch low {
		case "script", "style":
			_, i = CloseSpan(source, end, low)
		case "input":
			if sp, add := makeInput(i, end, readAttrs(raw)); add {
				out = append(out, sp)
			}
			i = end
		case "textarea":
			sp, at := makeArea(source, i, end, readAttrs(raw))
			if sp.Control.ID != "" {
				out = append(out, sp)
			}
			i = at
		case "select":
			sp, at := makeSelect(source, i, end, readAttrs(raw))
			if sp.Control.ID != "" {
				out = append(out, sp)
			}
			i = at
		case "button":
			_, at := CloseSpan(source, end, low)
			a := readAttrs(raw)
			if a.id != "" {
				c := controlFrom(low, a)
				c.Type = "button"
				out = append(out, Span{Control: c, Start: i, End: at})
			}
			i = at
		default:
			i = end
		}
	}

	return out
}

// CloseSpan returns the indexes of the matching close tag, searching from i.
// The first result is the '<' of the close tag. The second is the index after it.
// Both are len(s) when the close tag is missing.
func CloseSpan(s string, i int, tag string) (int, int) {
	n := len(tag)
	for i < len(s) {
		j := strings.Index(s[i:], "</")
		if j < 0 {
			return len(s), len(s)
		}
		i += j
		k := i + CloseTagLen
		for k < len(s) && isSpaceByte(s[k]) {
			k++
		}
		if k+n > len(s) || !strings.EqualFold(s[k:k+n], tag) {
			i += CloseTagLen
			continue
		}
		k += n
		if k < len(s) && !isSpaceByte(s[k]) && s[k] != '>' {
			i += CloseTagLen
			continue
		}
		for k < len(s) && s[k] != '>' {
			k++
		}
		if k < len(s) {
			k++
		}

		return i, k
	}

	return len(s), len(s)
}
