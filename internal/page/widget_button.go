package page

import (
	"html"
	"strings"
)

// rewriteButtons replaces a submit, button, or reset input with a button
// element so the engine gives it a face and a hit box.
func rewriteButtons(source string) string {
	var b strings.Builder
	i, prev := 0, 0
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
		if low == "script" || low == "style" {
			_, i = closeSpan(source, end, low)
			continue
		}
		kind := ""
		if low == "input" {
			kind = strings.ToLower(attrValue(raw, "type"))
		}
		if kind != "submit" && kind != "button" && kind != "reset" {
			i = end
			continue
		}
		b.WriteString(source[prev:i])
		b.WriteString(buttonFromInput(raw, kind))
		prev = end
		i = end
	}
	if prev == 0 {
		return source
	}
	b.WriteString(source[prev:])

	return b.String()
}

func buttonFromInput(raw, kind string) string {
	label := attrValue(raw, "value")
	switch {
	case label != "":
	case kind == "submit":
		label = "Submit"
	case kind == "reset":
		label = "Reset"
	}

	return openTagDrop("button", raw, nil, buttonDrop) + html.EscapeString(label) + "</button>"
}

func buttonDrop(name string) bool {
	return name == "type" || name == "value"
}
