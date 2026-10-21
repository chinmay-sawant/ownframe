package page

import "strings"

const (
	commentOpenLen  = 4 // len("<!--").
	commentCloseLen = 3 // len("-->").
	commentTotalLen = commentOpenLen + commentCloseLen
	closeTagLen     = 2 // len("</").
)

func findHead(source string) int {
	const end = "</head>"
	for i := 0; i+len(end) <= len(source); {
		if strings.HasPrefix(source[i:], "<!--") {
			i = skipComment(source, i)
			continue
		}
		if strings.EqualFold(source[i:i+len(end)], end) {
			return i
		}
		i++
	}

	return -1
}

func headOpen(source string) int {
	const open = "<head"
	for i := 0; i+len(open) <= len(source); {
		if strings.HasPrefix(source[i:], "<!--") {
			i = skipComment(source, i)
			continue
		}
		if !strings.EqualFold(source[i:i+len(open)], open) {
			i++
			continue
		}
		j := i + len(open)
		if j < len(source) && source[j] != '>' && !isSpace(source[j]) {
			i++
			continue
		}
		quote := byte(0)
		for j < len(source) {
			c := source[j]
			switch {
			case quote != 0:
				if c == quote {
					quote = 0
				}
			case c == '"' || c == '\'':
				quote = c
			case c == '>':
				return j + 1
			}
			j++
		}
		i++
	}

	return -1
}

func skipComment(source string, start int) int {
	end := strings.Index(source[start+commentOpenLen:], "-->")
	if end < 0 {
		return len(source)
	}

	return start + commentOpenLen + end + commentCloseLen
}

func spanCovers(spans []controlSpan, index int) bool {
	for _, sp := range spans {
		if index > sp.Start && index < sp.End {
			return true
		}
	}

	return false
}

func pickControl(sp controlSpan, live map[string]Control) Control {
	ctrl := sp.Control
	got, ok := live[ctrl.ID]
	if !ok {
		return ctrl
	}

	ctrl.Value = got.Value
	ctrl.Checked = got.Checked
	ctrl.Disabled = got.Disabled
	ctrl.Options = got.Options
	ctrl.Type = got.Type
	ctrl.Name = got.Name
	ctrl.Tag = got.Tag

	return ctrl
}
