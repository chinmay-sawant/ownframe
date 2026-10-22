package page

import "html"

func controlFrom(tag string, a fieldAttr) Control {
	return Control{
		ID:        a.id,
		Tag:       tag,
		Name:      a.name,
		Bind:      a.bind,
		Disabled:  a.disabled,
		TabIndex:  a.tabindex,
		MaxLength: a.maxlen,
	}
}

func makeInput(start, end int, a fieldAttr) (controlSpan, bool) {
	kind := a.kind
	text := textLike(kind)
	check := kind == "checkbox" || kind == "radio"
	if a.id == "" || (!text && !check) {
		return controlSpan{}, false
	}

	c := controlFrom("input", a)
	c.Type = kind
	c.Value = a.val
	if check {
		c.Checked = a.checked
	}

	return controlSpan{Control: c, Start: start, End: end}, true
}

func makeArea(s string, start, openEnd int, a fieldAttr) (controlSpan, int) {
	from, end := closeSpan(s, openEnd, "textarea")
	if a.id == "" {
		return controlSpan{}, end
	}

	c := controlFrom("textarea", a)
	c.Value = html.UnescapeString(s[openEnd:from])

	return controlSpan{Control: c, Start: start, End: end}, end
}

func makeSelect(s string, start, openEnd int, a fieldAttr) (controlSpan, int) {
	from, end := closeSpan(s, openEnd, "select")
	if a.id == "" {
		return controlSpan{}, end
	}

	opts := readOptions(s[openEnd:from])
	c := controlFrom("select", a)
	c.Options = opts
	c.Value = pickOption(opts)

	return controlSpan{Control: c, Start: start, End: end}, end
}
