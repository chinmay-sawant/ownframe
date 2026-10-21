package page

import "strings"

func scanControls(source string) []controlSpan {
	out := []controlSpan{}
	i := 0
	for i < len(source) {
		next := strings.IndexByte(source[i:], '<')
		if next < 0 {
			break
		}
		i += next
		if strings.HasPrefix(source[i:], "<!--") {
			end := strings.Index(source[i+commentOpenLen:], "-->")
			if end < 0 {
				break
			}
			i += end + commentTotalLen
			continue
		}

		name, raw, end, ok := readOpen(source, i)
		if !ok {
			i++
			continue
		}

		low := strings.ToLower(name)
		switch low {
		case "script", "style":
			_, i = closeSpan(source, end, low)
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
			_, at := closeSpan(source, end, low)
			a := readAttrs(raw)
			if a.id != "" {
				c := controlFrom(low, a)
				c.Type = "button"
				out = append(out, controlSpan{Control: c, Start: i, End: at})
			}
			i = at
		default:
			i = end
		}
	}

	return out
}

func closeSpan(s string, i int, tag string) (int, int) {
	n := len(tag)
	for i < len(s) {
		j := strings.Index(s[i:], "</")
		if j < 0 {
			return len(s), len(s)
		}
		i += j
		k := i + closeTagLen
		for k < len(s) && isSpaceByte(s[k]) {
			k++
		}
		if k+n > len(s) || !strings.EqualFold(s[k:k+n], tag) {
			i += closeTagLen
			continue
		}
		k += n
		if k < len(s) && !isSpaceByte(s[k]) && s[k] != '>' {
			i += closeTagLen
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
