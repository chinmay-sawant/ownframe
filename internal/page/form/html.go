package form

import (
	"html"
	"sort"
	"strings"
)

const formCSS = `<style>[data-ownframe-field]{display:inline-block;border:1px solid #c8c2b4;padding:4px 6px;min-width:10em;min-height:1.6em;white-space:pre-wrap}[data-ownframe-field][data-ownframe-focus="1"]{border:1px solid #1a56db;outline:1px solid #1a56db}[data-ownframe-selection]{background:#d6e2ff}[data-ownframe-composition]{text-decoration:underline}[data-ownframe-field][data-ownframe-placeholder="1"]{color:#6b7280}</style>`

// Rewrite copies source, swapping each kept control for paintable
// HTML. It paints a select-all when selected is true and puts the caret at
// the end of the value. The page uses RewriteState, which carries
// the caret and the range.
func Rewrite(source string, spans []Span, live map[string]Control, focusID string, selected bool) string {
	st := Caret{Caret: -1}
	if focusID != "" {
		st.Focus = focusID
		st.All = selected
	}

	return RewriteState(source, spans, live, st)
}

func openTag(tag, raw string, extra []string, input bool) string {
	return OpenTagDrop(tag, raw, extra, func(n string) bool { return skipAttr(n, input) })
}

func skipAttr(name string, input bool) bool {
	if !input {
		return false
	}

	return name == "type" || name == "checked" || name == "disabled"
}

func attrKey(attr string) string {
	cut := strings.IndexByte(attr, '=')
	if cut >= 0 {
		attr = attr[:cut]
	}

	return strings.ToLower(strings.TrimSpace(attr))
}

func attrRegion(raw string) string {
	end := len(raw)
	quote := byte(0)
	for i := 0; i < len(raw); i++ {
		c := raw[i]
		if quote != 0 {
			if c == quote {
				quote = 0
			}
			continue
		}
		if c == '"' || c == '\'' {
			quote = c
			continue
		}
		if c == '>' {
			end = i
			break
		}
	}

	i := 0
	if raw != "" && raw[0] == '<' {
		i = 1
	}
	for i < end && !isSpace(raw[i]) && raw[i] != '/' {
		i++
	}
	s := strings.TrimSpace(raw[i:end])

	return strings.TrimSpace(strings.TrimSuffix(s, "/"))
}

func splitAttrs(s string) []string {
	out := make([]string, 0)
	i := 0
	for i < len(s) {
		i = skipSpace(s, i)
		if i >= len(s) {
			break
		}
		start := i
		for i < len(s) && !isSpace(s[i]) && s[i] != '=' {
			i++
		}
		j := skipSpace(s, i)
		if j < len(s) && s[j] == '=' {
			i = skipValue(s, skipSpace(s, j+1))
		}
		out = append(out, strings.TrimSpace(s[start:i]))
	}

	return out
}

// fieldSpanState rewrites one text-like control into a span. It paints the
// range as a selection span and the whole-value select as the field
// background. The page paints the caret after layout.
func fieldSpanState(raw string, ctrl Control, st Caret, focused bool) string {
	kind := strings.ToLower(strings.TrimSpace(ctrl.Type))
	tag := strings.ToLower(strings.TrimSpace(ctrl.Tag))
	text := ShownText(kind, ctrl.Value)
	ph := ""
	if text == "" && kind != "file" {
		ph = AttrValue(raw, "placeholder")
	}
	if st.Preedit != "" {
		ph = ""
	}

	start, end := RangeOf(st.Start, st.End, RuneLen(ctrl.Value))
	whole := focused && (st.All || (start != end && start == 0 && end == RuneLen(ctrl.Value)))

	extras := []string{`data-ownframe-field="` + tag + `"`}
	if focused {
		extras = append(extras, focusAttr)
	}
	if whole {
		extras = append(extras, selectedAttr)
	}

	inner := ""
	switch {
	case ph != "":
		extras = append(extras, placeholderAttr)
		inner = html.EscapeString(ph)
	case st.Preedit != "":
		inner = markedPreedit(text, ShownText(kind, st.Preedit), st.Start,
			st.End, st.PreeditStart, st.PreeditEnd)
	case whole:
		inner = selectionSpan(text)
	default:
		inner = markedText(text, focused, start, end)
	}

	return OpenTagDrop("span", raw, extras, spanDrop) + inner + "</span>"
}

// OpenTagDrop writes an open tag, copying attributes from raw except names drop rejects.
func OpenTagDrop(tag, raw string, extra []string, drop func(string) bool) string {
	var b strings.Builder
	b.WriteByte('<')
	b.WriteString(tag)
	for _, attr := range splitAttrs(attrRegion(raw)) {
		name := attrKey(attr)
		if strings.HasPrefix(name, "data-ownframe-") || drop(name) {
			continue
		}
		b.WriteByte(' ')
		b.WriteString(attr)
	}
	for _, attr := range extra {
		b.WriteByte(' ')
		b.WriteString(attr)
	}
	b.WriteByte('>')

	return b.String()
}

// AttrValue returns the unescaped value of name in an open tag, or "".
func AttrValue(raw, name string) string {
	for _, attr := range splitAttrs(attrRegion(raw)) {
		if attrKey(attr) != name {
			continue
		}
		cut := strings.IndexByte(attr, '=')
		if cut < 0 {
			return ""
		}
		val := strings.TrimSpace(attr[cut+1:])
		if len(val) >= 2 && (val[0] == '"' || val[0] == '\'') && val[len(val)-1] == val[0] {
			val = val[1 : len(val)-1]
		}

		return html.UnescapeString(val)
	}

	return ""
}

func spanDrop(name string) bool {
	return name == "value" || name == "type"
}

func keepSpans(source string, spans []Span) []Span {
	ordered := make([]Span, len(spans))
	copy(ordered, spans)
	sort.Slice(ordered, func(i, j int) bool {
		return ordered[i].Start < ordered[j].Start
	})

	kept := make([]Span, 0, len(ordered))
	prev := 0
	n := len(source)
	for _, sp := range ordered {
		if sp.Start < prev || sp.End > n || sp.End <= sp.Start {
			continue
		}
		kept = append(kept, sp)
		prev = sp.End
	}

	return kept
}

// selectionSpan wraps selected runes in the library selection element, so
// the highlight paints inside the field and author field backgrounds cannot
// cover it.
func selectionSpan(text string) string {
	return `<span data-ownframe-selection="1">` + html.EscapeString(text) + `</span>`
}

// markedText leaves caret placement to the display-list overlay so adding a
// caret never splits or reflows the field's text run.
func markedText(text string, focused bool, start, end int) string {
	if !focused {
		return html.EscapeString(text)
	}

	runes := []rune(text)
	if end > start {
		return html.EscapeString(string(runes[:start])) +
			selectionSpan(string(runes[start:end])) +
			html.EscapeString(string(runes[end:]))
	}

	return html.EscapeString(text)
}

const (
	CommentOpenLen  = len("<!--")
	CommentCloseLen = len("-->")
	CommentTotalLen = CommentOpenLen + CommentCloseLen
	CloseTagLen     = len("</")
)

// FindHead is the index of the first </head> outside a comment, or -1.
func FindHead(source string) int {
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

// HeadOpen is the index just after the first <head> outside a comment, or -1.
func HeadOpen(source string) int {
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
	end := strings.Index(source[start+CommentOpenLen:], "-->")
	if end < 0 {
		return len(source)
	}

	return start + CommentOpenLen + end + CommentCloseLen
}

func spanCovers(spans []Span, index int) bool {
	for _, sp := range spans {
		if index > sp.Start && index < sp.End {
			return true
		}
	}

	return false
}

func pickControl(sp Span, live map[string]Control) Control {
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

func markedPreedit(value, preedit string, start, end, selStart, selEnd int) string {
	base, composing := []rune(value), []rune(preedit)
	start, end = RangeOf(start, end, len(base))
	if end == start {
		start = ClampPos(start, len(base))
	}
	a, b := RuneAtByte(preedit, selStart), RuneAtByte(preedit, selEnd)
	a, b = RangeOf(a, b, len(composing))
	var out strings.Builder
	out.WriteString(html.EscapeString(string(base[:start])))
	out.WriteString(`<span data-ownframe-composition="1">`)
	out.WriteString(html.EscapeString(string(composing[:a])))
	if a != b {
		out.WriteString(selectionSpan(string(composing[a:b])))
	}
	out.WriteString(html.EscapeString(string(composing[b:])))
	out.WriteString(`</span>`)
	out.WriteString(html.EscapeString(string(base[end:])))
	return out.String()
}

// RewriteState copies source, swapping each kept control for
// paintable HTML and adding the default form styles.
func RewriteState(source string, spans []Span, live map[string]Control, st Caret) string {
	if !hasFieldSpans(spans) {
		return source
	}

	css := formCSS
	kept := keepSpans(source, spans)
	cssAt := HeadOpen(source)
	if cssAt < 0 {
		cssAt = FindHead(source)
	}
	if cssAt >= 0 && spanCovers(kept, cssAt) {
		cssAt = -1
	}

	var b strings.Builder
	prev := 0
	placed := false
	write := func(end int) {
		if !placed && cssAt >= prev && cssAt <= end {
			b.WriteString(source[prev:cssAt])
			b.WriteString(css)
			b.WriteString(source[cssAt:end])
			placed = true

			return
		}

		b.WriteString(source[prev:end])
	}
	for _, sp := range kept {
		write(sp.Start)
		ctrl := pickControl(sp, live)
		b.WriteString(replaceControlState(source[sp.Start:sp.End], ctrl, st))
		prev = sp.End
	}
	write(len(source))
	if !placed {
		return css + b.String()
	}

	return b.String()
}

func replaceControlState(raw string, ctrl Control, st Caret) string {
	focused := st.Focus != "" && ctrl.ID == st.Focus
	kind := strings.ToLower(strings.TrimSpace(ctrl.Type))
	tag := strings.ToLower(strings.TrimSpace(ctrl.Tag))
	switch {
	case tag == "textarea":
		return fieldSpanState(raw, ctrl, st, focused)
	case tag == "select":
		return boxElement(raw, "select", selectLabel(ctrl), ctrl.ID, st.Focus, focused && st.All)
	case kind == "checkbox" || kind == "radio":
		return openTag("input", raw, inputExtras(ctrl, kind, st.Focus), true)
	case TextLike(kind):
		return fieldSpanState(raw, ctrl, st, focused)
	default:
		return raw
	}
}

func boxElement(raw, tag, body, id, focusID string, selected bool) string {
	extra := []string{`data-ownframe-field="` + tag + `"`}
	if focusID != "" && id == focusID {
		extra = append(extra, focusAttr)
	}
	if selected {
		extra = append(extra, selectedAttr)
	}

	return OpenTagDrop(tag, raw, extra, func(string) bool { return false }) + html.EscapeString(body) + "</" + tag + ">"
}

func inputExtras(ctrl Control, kind, focusID string) []string {
	extra := []string{`type="` + kind + `"`}
	if ctrl.Checked {
		extra = append(extra, "checked")
	}
	if ctrl.Disabled {
		extra = append(extra, "disabled")
	}
	if focusID != "" && ctrl.ID == focusID {
		extra = append(extra, focusAttr)
	}

	return extra
}

func selectLabel(ctrl Control) string {
	for _, opt := range ctrl.Options {
		if opt.Value == ctrl.Value {
			return opt.Label
		}
	}
	if len(ctrl.Options) > 0 {
		return ctrl.Options[0].Label
	}

	return ctrl.Value
}

func skipSpace(s string, i int) int {
	for i < len(s) && isSpace(s[i]) {
		i++
	}

	return i
}

func skipValue(s string, i int) int {
	if i >= len(s) {
		return i
	}
	if s[i] == '"' || s[i] == '\'' {
		q := s[i]
		i++
		for i < len(s) && s[i] != q {
			i++
		}
		if i < len(s) {
			i++
		}

		return i
	}
	for i < len(s) && !isSpace(s[i]) {
		i++
	}

	return i
}

func isSpace(c byte) bool {
	return c == ' ' || c == '\t' || c == '\n' || c == '\r' || c == '\f'
}

const (
	focusAttr       = `data-ownframe-focus="1"`
	selectedAttr    = `data-ownframe-selected="1"`
	placeholderAttr = `data-ownframe-placeholder="1"`
)
