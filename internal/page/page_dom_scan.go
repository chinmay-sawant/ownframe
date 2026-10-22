package page

import "strings"

// domTag parses one open or close tag at s[i] == '<'. It returns the
// lowercase name, whether it closes, whether it self-closes, and the
// index after '>'. A missing '>' consumes the rest.
// quotePair is the opening plus closing quote a skip jumps over.
const quotePair = 2

func domTag(s string, i int) (string, bool, bool, int) {
	j := i + 1
	close := false
	if j < len(s) && s[j] == '/' {
		close = true
		j++
	}
	from := j
	for j < len(s) && domNameByte(s[j]) {
		j++
	}
	name := strings.ToLower(s[from:j])
	for j < len(s) && s[j] != '>' {
		if s[j] == '"' || s[j] == '\'' {
			if end := strings.IndexByte(s[j+1:], s[j]); end >= 0 {
				j += end + quotePair
				continue
			}
			return name, close, true, len(s)
		}
		j++
	}
	self := j > i+1 && s[j-1] == '/'
	if j < len(s) {
		j++
	}

	return name, close, self, j
}

func domNameByte(c byte) bool {
	switch {
	case c == '_' || c == '-' || c == ':':
		return true
	case c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z':
		return true
	case c >= '0' && c <= '9':
		return true
	default:
		return false
	}
}
