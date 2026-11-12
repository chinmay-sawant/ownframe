package web

import (
	"html"
	"strings"

	"github.com/chinmay-sawant/ownframe/internal/page"
)

// browserDocument turns the executed paint source into the HTML document
// the browser shows. controls carry the real values and select options.
// theme is inserted as a stylesheet. boot, when non-empty, is the live
// script and is inserted once.
func browserDocument(html, theme string, controls []page.WebControl, boot string) string {
	doc := restoreControls(html, controls)

	return insertPreviewHead(doc, theme, boot)
}

// attr is one attribute copied from an open tag. name is lower-case.
// raw is the author text, including its quotes. value is unescaped.
type attr struct {
	name  string
	raw   string
	value string
}

// htmlTag is one tag in the paint source. end is the index after '>'.
type htmlTag struct {
	name      string
	attrs     []attr
	start     int
	end       int
	closing   bool
	selfClose bool
	comment   bool
	other     bool
}

func restoreControls(src string, controls []page.WebControl) string {
	byID := indexControls(controls)
	if len(byID) == 0 {
		return src
	}

	var b strings.Builder
	b.Grow(len(src))
	i := 0
	for i < len(src) {
		if src[i] != '<' {
			next := strings.IndexByte(src[i:], '<')
			if next < 0 {
				b.WriteString(src[i:])

				return b.String()
			}
			b.WriteString(src[i : i+next])
			i += next

			continue
		}

		tag, ok := parseTag(src, i)
		if !ok || tag.end <= i {
			b.WriteByte(src[i])
			i++

			continue
		}

		kind := fieldKind(tag)
		ctrl, matched := matchControl(tag, kind, byID)
		if !matched {
			b.WriteString(src[tag.start:tag.end])
			i = tag.end

			continue
		}

		end := matchingEnd(src, tag)
		if end < 0 {
			b.WriteString(src[tag.start:tag.end])
			i = tag.end

			continue
		}

		writeControl(&b, kind, tag.attrs, ctrl)
		i = end
	}

	return b.String()
}

func indexControls(controls []page.WebControl) map[string]*page.WebControl {
	if len(controls) == 0 {
		return nil
	}

	byID := make(map[string]*page.WebControl, len(controls))
	for i := range controls {
		id := controls[i].ID
		if id == "" {
			continue
		}
		byID[id] = &controls[i]
	}

	return byID
}

// matchControl pairs a field element with the live control of the same id.
// Checkbox and radio inputs stay as the engine wrote them.
func matchControl(tag htmlTag, kind string, byID map[string]*page.WebControl) (*page.WebControl, bool) {
	if kind == "" || realCheck(tag) {
		return nil, false
	}

	id := attrValue(tag.attrs, "id")
	if id == "" {
		return nil, false
	}

	ctrl, ok := byID[id]
	if !ok {
		return nil, false
	}

	return ctrl, true
}

// realCheck reports a checkbox or radio the engine left as a real input.
func realCheck(tag htmlTag) bool {
	if tag.name != "input" {
		return false
	}

	switch strings.ToLower(strings.TrimSpace(attrValue(tag.attrs, "type"))) {
	case "checkbox", "radio":
		return true
	default:
		return false
	}
}

func fieldKind(tag htmlTag) string {
	if tag.closing || tag.comment || tag.other {
		return ""
	}

	kind := strings.ToLower(strings.TrimSpace(attrValue(tag.attrs, "data-ownframe-field")))
	switch kind {
	case "input", "textarea", "select":
		return kind
	default:
		return ""
	}
}

func attrValue(attrs []attr, name string) string {
	for _, a := range attrs {
		if a.name == name {
			return a.value
		}
	}

	return ""
}

// matchingEnd is the index after the close tag that pairs with open.
// Depth counts only that tag name, so a nested selection span stays inside.
func matchingEnd(src string, open htmlTag) int {
	if open.selfClose {
		return open.end
	}

	depth := 1
	i := open.end
	for i < len(src) {
		if src[i] != '<' {
			i++

			continue
		}

		tag, ok := parseTag(src, i)
		if !ok || tag.end <= i {
			i++

			continue
		}
		if tag.name == open.name {
			switch {
			case tag.closing:
				depth--
				if depth == 0 {
					return tag.end
				}
			case !tag.selfClose:
				depth++
			}
		}
		i = tag.end
	}

	return -1
}

func writeControl(b *strings.Builder, kind string, attrs []attr, ctrl *page.WebControl) {
	switch kind {
	case "textarea":
		writeOpen(b, "textarea", attrs, ctrl, false)
		b.WriteString(html.EscapeString(ctrl.Value))
		b.WriteString("</textarea>")
	case "select":
		writeOpen(b, "select", attrs, ctrl, false)
		for _, opt := range ctrl.Options {
			writeOption(b, opt)
		}
		b.WriteString("</select>")
	default:
		writeOpen(b, "input", attrs, ctrl, true)
	}
}

func writeOpen(b *strings.Builder, tag string, attrs []attr, ctrl *page.WebControl, asInput bool) {
	b.WriteByte('<')
	b.WriteString(tag)

	disabled := false
	for _, a := range attrs {
		if skipCopiedAttr(a.name) {
			continue
		}
		if a.name == "disabled" {
			disabled = true
		}
		b.WriteByte(' ')
		b.WriteString(a.raw)
	}
	if asInput {
		b.WriteString(` type="`)
		b.WriteString(html.EscapeString(inputType(ctrl.Type)))
		b.WriteString(`" value="`)
		b.WriteString(html.EscapeString(ctrl.Value))
		b.WriteByte('"')
	}
	if ctrl.Disabled && !disabled {
		b.WriteString(" disabled")
	}
	b.WriteByte('>')
}

func writeOption(b *strings.Builder, opt page.SelectOption) {
	b.WriteString(`<option value="`)
	b.WriteString(html.EscapeString(opt.Value))
	b.WriteByte('"')
	if opt.Selected {
		b.WriteString(" selected")
	}
	if opt.Disabled {
		b.WriteString(" disabled")
	}
	b.WriteByte('>')
	b.WriteString(html.EscapeString(opt.Label))
	b.WriteString("</option>")
}

func skipCopiedAttr(name string) bool {
	return strings.HasPrefix(name, "data-ownframe-") || name == "value" || name == "type"
}

// inputType is the control type. An empty input type is text.
// A password type is left as password.
func inputType(kind string) string {
	kind = strings.TrimSpace(kind)
	if kind == "" {
		return "text"
	}

	return kind
}

func insertPreviewHead(doc, theme, boot string) string {
	extra := previewHeadExtra(doc, theme, boot)
	if extra == "" {
		return doc
	}
	if at := indexHeadClose(doc); at >= 0 {
		return doc[:at] + extra + doc[at:]
	}

	// No </head>: wrap the preview tags in a head of their own.
	head := "<head>" + extra + "</head>"
	if at := tagEdge(doc, "html", true); at >= 0 {
		return doc[:at] + head + doc[at:]
	}
	if at := tagEdge(doc, "body", false); at >= 0 {
		return doc[:at] + head + doc[at:]
	}

	return head + doc
}

// previewHeadExtra is the theme stylesheet and the boot script.
// The boot source is JavaScript and is not escaped. A document that
// already has id="ownframe-boot" does not receive a second copy.
func previewHeadExtra(doc, theme, boot string) string {
	var b strings.Builder
	if theme != "" {
		b.WriteString(`<style id="of-theme">`)
		b.WriteString(theme)
		b.WriteString("</style>")
	}
	if boot != "" && !hasBoot(doc) {
		b.WriteString(`<script id="ownframe-boot">`)
		b.WriteString(boot)
		b.WriteString("</script>")
	}

	return b.String()
}

func hasBoot(doc string) bool {
	return strings.Contains(doc, `id="ownframe-boot"`) || strings.Contains(doc, `id='ownframe-boot'`)
}

func indexHeadClose(doc string) int {
	const end = "</head>"
	n := len(end)
	for i := 0; i+n <= len(doc); {
		if strings.HasPrefix(doc[i:], "<!--") {
			i = skipComment(doc, i)

			continue
		}
		if strings.EqualFold(doc[i:i+n], end) {
			return i
		}
		i++
	}

	return -1
}

// tagEdge returns the start of the first open tag named name, or the index
// just after it when after is set. -1 when that tag is absent.
func tagEdge(doc, name string, after bool) int {
	i := 0
	for i < len(doc) {
		if doc[i] != '<' {
			rel := strings.IndexByte(doc[i:], '<')
			if rel < 0 {
				return -1
			}
			i += rel

			continue
		}

		tag, ok := parseTag(doc, i)
		if !ok || tag.end <= i {
			i++

			continue
		}
		if !tag.comment && !tag.other && !tag.closing && tag.name == name {
			if after {
				return tag.end
			}

			return tag.start
		}
		i = tag.end
	}

	return -1
}

func parseTag(src string, at int) (htmlTag, bool) {
	if at >= len(src) || src[at] != '<' {
		return htmlTag{}, false
	}
	if strings.HasPrefix(src[at:], "<!--") {
		end := skipComment(src, at)
		if end <= at {
			return htmlTag{}, false
		}

		return htmlTag{start: at, end: end, comment: true}, true
	}

	gt := scanTagEnd(src, at+1)
	if gt <= at {
		return htmlTag{}, false
	}

	tag := htmlTag{start: at, end: gt}
	innerLo := at + 1
	innerHi := gt - len(">")
	if innerHi < innerLo {
		return htmlTag{}, false
	}
	if self, slash := selfCloseSlash(src, at, gt); self {
		tag.selfClose = true
		innerHi = slash
	}

	inner := strings.TrimSpace(src[innerLo:innerHi])
	if inner == "" {
		return htmlTag{}, false
	}
	switch inner[0] {
	case '!', '?':
		tag.other = true

		return tag, true
	case '/':
		tag.closing = true
		inner = strings.TrimSpace(inner[len("/"):])
	}

	name, rest, ok := cutName(inner)
	if !ok {
		return htmlTag{}, false
	}
	tag.name = strings.ToLower(name)
	if !tag.closing {
		tag.attrs = parseAttrs(rest)
	}

	return tag, true
}

func selfCloseSlash(src string, at, gt int) (bool, int) {
	i := gt - len(">")
	if i <= at {
		return false, 0
	}
	i--
	for i > at && isSpace(src[i]) {
		i--
	}
	if i > at && src[i] == '/' {
		return true, i
	}

	return false, 0
}

func scanTagEnd(src string, i int) int {
	quote := byte(0)
	for i < len(src) {
		c := src[i]
		switch {
		case quote != 0:
			if c == quote {
				quote = 0
			}
		case c == '"' || c == '\'':
			quote = c
		case c == '>':
			return i + 1
		}
		i++
	}

	return -1
}

func skipComment(src string, i int) int {
	const open = "<!--"
	const close = "-->"
	rest := strings.Index(src[i+len(open):], close)
	if rest < 0 {
		return len(src)
	}

	return i + len(open) + rest + len(close)
}

func cutName(s string) (string, string, bool) {
	if s == "" || !isNameStart(s[0]) {
		return "", "", false
	}

	i := 1
	for i < len(s) && isNameChar(s[i]) {
		i++
	}

	return s[:i], s[i:], true
}

func parseAttrs(s string) []attr {
	out := make([]attr, 0)
	i := 0
	for i < len(s) {
		i = skipSpace(s, i)
		if i >= len(s) {
			break
		}
		if !isNameStart(s[i]) {
			for i < len(s) && !isSpace(s[i]) {
				i++
			}

			continue
		}

		a, next := readAttr(s, i)
		if next <= i {
			i++

			continue
		}
		out = append(out, a)
		i = next
	}

	return out
}

func readAttr(s string, i int) (attr, int) {
	start := i
	i++
	for i < len(s) && isNameChar(s[i]) {
		i++
	}

	name := strings.ToLower(s[start:i])
	j := skipSpace(s, i)
	if j >= len(s) || s[j] != '=' {
		return attr{name: name, raw: s[start:i]}, i
	}

	k := skipSpace(s, j+1)
	val, next := readValue(s, k)

	return attr{name: name, raw: strings.TrimSpace(s[start:next]), value: val}, next
}

func readValue(s string, i int) (string, int) {
	if i >= len(s) {
		return "", i
	}
	if s[i] == '"' || s[i] == '\'' {
		q := s[i]
		i++
		start := i
		for i < len(s) && s[i] != q {
			i++
		}
		inner := s[start:i]
		if i < len(s) {
			i++
		}

		return html.UnescapeString(inner), i
	}

	start := i
	for i < len(s) && !isSpace(s[i]) {
		i++
	}

	return html.UnescapeString(s[start:i]), i
}

func skipSpace(s string, i int) int {
	for i < len(s) && isSpace(s[i]) {
		i++
	}

	return i
}

func isSpace(c byte) bool {
	switch c {
	case ' ', '\t', '\n', '\r', '\f':
		return true
	default:
		return false
	}
}

func isNameStart(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

func isNameChar(c byte) bool {
	return isNameStart(c) || (c >= '0' && c <= '9') || c == '-' || c == ':' || c == '_'
}
