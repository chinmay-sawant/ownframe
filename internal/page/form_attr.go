package page

// fieldAttr is the attribute set read from one open tag.
type fieldAttr struct {
	id, kind, name, val, bind   string
	tabindex                    int
	maxlen                      int
	hasVal                      bool
	hasMaxlen                   bool
	checked, disabled, selected bool
}

func isNameByte(c byte) bool {
	return c == '_' || c == '-' || c == ':' ||
		(c >= 'A' && c <= 'Z') ||
		(c >= 'a' && c <= 'z') ||
		(c >= '0' && c <= '9')
}

func isSpaceByte(c byte) bool {
	return c == ' ' || c == '\n' || c == '\r' || c == '\t' || c == '\f'
}

func readOpen(s string, i int) (name, raw string, end int, ok bool) {
	if i >= len(s) || s[i] != '<' {
		return "", "", i, false
	}

	j := i + 1
	if j < len(s) && (s[j] == '/' || s[j] == '!') {
		return "", "", i, false
	}

	from := j
	for j < len(s) && isNameByte(s[j]) {
		j++
	}
	if j == from {
		return "", "", i, false
	}

	name = s[from:j]
	rawFrom := j
	var quote byte
	for j < len(s) {
		c := s[j]
		switch {
		case quote != 0:
			if c == quote {
				quote = 0
			}
		case c == '"' || c == '\'':
			quote = c
		case c == '>':
			return name, s[rawFrom:j], j + 1, true
		}
		j++
	}

	return "", "", i, false
}
