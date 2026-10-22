package form

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

func makeInput(start, end int, a fieldAttr) (Span, bool) {
	kind := a.kind
	text := TextLike(kind)
	check := kind == "checkbox" || kind == "radio"
	if a.id == "" || (!text && !check) {
		return Span{}, false
	}

	c := controlFrom("input", a)
	c.Type = kind
	c.Value = a.val
	if check {
		c.Checked = a.checked
	}

	return Span{Control: c, Start: start, End: end}, true
}

func makeArea(s string, start, openEnd int, a fieldAttr) (Span, int) {
	from, end := CloseSpan(s, openEnd, "textarea")
	if a.id == "" {
		return Span{}, end
	}

	c := controlFrom("textarea", a)
	c.Value = html.UnescapeString(s[openEnd:from])

	return Span{Control: c, Start: start, End: end}, end
}

func makeSelect(s string, start, openEnd int, a fieldAttr) (Span, int) {
	from, end := CloseSpan(s, openEnd, "select")
	if a.id == "" {
		return Span{}, end
	}

	opts := readOptions(s[openEnd:from])
	c := controlFrom("select", a)
	c.Options = opts
	c.Value = PickOption(opts)

	return Span{Control: c, Start: start, End: end}, end
}
