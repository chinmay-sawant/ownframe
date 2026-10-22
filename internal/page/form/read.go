package form

import (
	"html"
	"strconv"
	"strings"
)

func readAttrs(raw string) fieldAttr {
	var a fieldAttr
	i := 0
	for i < len(raw) {
		for i < len(raw) && isSpaceByte(raw[i]) {
			i++
		}
		if i >= len(raw) || raw[i] == '/' {
			break
		}

		from := i
		for i < len(raw) && isNameByte(raw[i]) {
			i++
		}
		if i == from {
			i++
			continue
		}

		key := strings.ToLower(raw[from:i])
		for i < len(raw) && isSpaceByte(raw[i]) {
			i++
		}

		val := ""
		has := false
		if i < len(raw) && raw[i] == '=' {
			has = true
			i++
			for i < len(raw) && isSpaceByte(raw[i]) {
				i++
			}
			val, i = readValue(raw, i)
		}
		applyAttr(&a, key, val, has)
	}

	return a
}

func readValue(raw string, i int) (string, int) {
	if i >= len(raw) {
		return "", i
	}
	if raw[i] == '"' || raw[i] == '\'' {
		q := raw[i]
		i++
		from := i
		for i < len(raw) && raw[i] != q {
			i++
		}
		val := raw[from:i]
		if i < len(raw) {
			i++
		}

		return html.UnescapeString(val), i
	}

	from := i
	for i < len(raw) && !isSpaceByte(raw[i]) {
		i++
	}

	val := strings.TrimRight(raw[from:i], "/")

	return html.UnescapeString(val), i
}

func applyAttr(a *fieldAttr, key, val string, has bool) {
	switch key {
	case "id":
		a.id = val
	case "type":
		a.kind = strings.ToLower(val)
	case "name":
		a.name = val
	case "data-bind":
		a.bind = val
	case "value":
		a.val = val
		a.hasVal = has
	case "checked":
		a.checked = true
	case "disabled":
		a.disabled = true
	case "tabindex":
		if n, err := strconv.Atoi(strings.TrimSpace(val)); err == nil {
			a.tabindex = n
		}
	case "maxlength":
		if n, err := strconv.Atoi(strings.TrimSpace(val)); err == nil && n > 0 {
			a.maxlen = n
			a.hasMaxlen = true
		}
	case "selected":
		a.selected = true
	}
}
