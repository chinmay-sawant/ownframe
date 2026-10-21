package page

import (
	"html"
	"strings"
)

func pickOption(opts []Option) string {
	if len(opts) == 0 {
		return ""
	}

	chosen := 0
	for i := range opts {
		if opts[i].Selected {
			chosen = i
			break
		}
	}
	for i := range opts {
		opts[i].Selected = i == chosen
	}

	return opts[chosen].Value
}

func readOptions(body string) []Option {
	out := []Option{}
	i := 0
	for i < len(body) {
		next := strings.IndexByte(body[i:], '<')
		if next < 0 {
			break
		}
		i += next
		if strings.HasPrefix(body[i:], "<!--") {
			end := strings.Index(body[i+commentOpenLen:], "-->")
			if end < 0 {
				break
			}
			i += end + commentTotalLen
			continue
		}

		name, raw, end, ok := readOpen(body, i)
		if !ok || !strings.EqualFold(name, "option") {
			if ok {
				i = end
			} else {
				i++
			}
			continue
		}

		opt, at := oneOption(body, end, readAttrs(raw))
		out = append(out, opt)
		i = at
	}

	return out
}

func oneOption(body string, openEnd int, a fieldAttr) (Option, int) {
	labelEnd := openEnd
	for labelEnd < len(body) && body[labelEnd] != '<' {
		labelEnd++
	}

	label := html.UnescapeString(body[openEnd:labelEnd])
	end := labelEnd
	if optionClose(body, labelEnd) {
		_, end = closeSpan(body, openEnd, "option")
	}

	val := label
	if a.hasVal {
		val = a.val
	}

	return Option{Value: val, Label: label, Selected: a.selected}, end
}

func optionClose(s string, i int) bool {
	const prefix = "</option"
	n := len(prefix)
	if i+n > len(s) || !strings.EqualFold(s[i:i+n], prefix) {
		return false
	}
	if i+n == len(s) {
		return true
	}

	return s[i+n] == '>' || isSpaceByte(s[i+n])
}
