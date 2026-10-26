package main

import (
	"strconv"
	"strings"
	"unicode/utf8"
)

const (
	previewEntityBytes = 12
	previewEntityBase  = 10
	previewEntityHex   = 16
	previewEntityBits  = 32
)

// stripHTML removes comments, script and style bodies, and tags, turns <br>
// into a newline, and keeps the alt text of an image. Entities are decoded
// later by previewInline.
func stripHTML(text string) string {
	text = stripHTMLHidden(text)
	if !strings.ContainsRune(text, '<') {
		return text
	}

	var b strings.Builder

	b.Grow(len(text))

	for i := 0; i < len(text); {
		if text[i] != '<' {
			b.WriteByte(text[i])
			i++

			continue
		}

		end, name, rest, closing, ok := htmlTagAt(text, i)
		if !ok {
			b.WriteByte('<')
			i++

			continue
		}

		switch {
		case name == "br":
			b.WriteByte('\n')
		case name == "hr":
			b.WriteByte('\n')
		case name == "img" && !closing:
			b.WriteString(htmlAttr(rest, "alt"))
		case closing && isHTMLBlockTag(name):
			b.WriteByte('\n')
		}

		i = end
	}

	return b.String()
}

// stripHTMLHidden removes comments and the whole body of a script or style
// element, so the preview never prints CSS or JavaScript as prose.
func stripHTMLHidden(text string) string {
	if !strings.ContainsRune(text, '<') {
		return text
	}

	var b strings.Builder

	b.Grow(len(text))

	for i := 0; i < len(text); {
		if strings.HasPrefix(text[i:], "<!--") {
			end := strings.Index(text[i+len("<!--"):], "-->")
			if end < 0 {
				break
			}

			i += len("<!--") + end + len("-->")

			continue
		}

		if text[i] != '<' {
			b.WriteByte(text[i])
			i++

			continue
		}

		end, name, _, closing, ok := htmlTagAt(text, i)
		if !ok {
			b.WriteByte('<')
			i++

			continue
		}

		if isHTMLHiddenTag(name) && !closing {
			i = htmlSkipRawElement(text, i, name)

			continue
		}

		b.WriteString(text[i:end])

		i = end
	}

	return b.String()
}

// htmlSkipRawElement returns the offset just past the closing tag of a script
// or style element, or the end of the text when the element never closes.
func htmlSkipRawElement(text string, i int, name string) int {
	lower := strings.ToLower(text[i:])
	needle := "</" + name

	end := strings.Index(lower, needle)
	if end < 0 {
		return len(text)
	}

	closeAt := strings.IndexByte(text[i+end:], '>')
	if closeAt < 0 {
		return len(text)
	}

	return i + end + closeAt + 1
}

// isHTMLHiddenTag reports whether the text inside a tag must be dropped.
func isHTMLHiddenTag(name string) bool {
	return name == "script" || name == "style"
}

// htmlTagAt parses the tag beginning at text[i].
func htmlTagAt(text string, i int) (end int, name, rest string, closing, ok bool) {
	closeAt := strings.IndexByte(text[i:], '>')
	if closeAt < 0 {
		return 0, "", "", false, false
	}

	end = i + closeAt + 1
	content := text[i+1 : end-1]

	name, rest, closing, ok = htmlTagHead(content)

	return end, name, rest, closing, ok
}

// htmlTagHead splits a tag body into a lower-case name and its attributes.
func htmlTagHead(content string) (name, rest string, closing, ok bool) {
	if strings.HasPrefix(content, "/") {
		closing = true
		content = content[1:]
	}

	if content == "" || !isHTMLNameStart(content[0]) {
		return "", "", false, false
	}

	j := 1
	for j < len(content) && isHTMLNameChar(content[j]) {
		j++
	}

	rest = content[j:]
	if rest != "" && !isHTMLSpace(rest[0]) && rest[0] != '/' {
		return "", "", false, false
	}

	return strings.ToLower(content[:j]), rest, closing, true
}

// isHTMLBlockStart reports whether a line starts a block-level HTML tag.
func isHTMLBlockStart(line string) bool {
	trimmed := strings.TrimSpace(line)
	if !strings.HasPrefix(trimmed, "<") {
		return false
	}

	closeAt := strings.IndexByte(trimmed, '>')
	if closeAt < 0 {
		return false
	}

	name, _, _, ok := htmlTagHead(trimmed[1:closeAt])
	if !ok {
		return false
	}

	return isHTMLBlockTag(name)
}

// isHTMLBlockTag reports whether a tag is rendered as a paragraph block.
func isHTMLBlockTag(name string) bool {
	switch name {
	case "p", "div", "details", "summary", "section", "article", "aside",
		"header", "footer", "main", "nav", "ul", "ol", "li", "blockquote",
		"pre", "table", "thead", "tbody", "tr", "td", "th", "figure",
		"figcaption", "br", "hr", "dl", "dt", "dd",
		"h1", "h2", "h3", "h4", "h5", "h6":
		return true
	}

	return false
}

// htmlTextLines turns a block of raw HTML into readable paragraphs.
func htmlTextLines(raw string) []string {
	plain := stripHTML(raw)
	lines := strings.Split(plain, "\n")
	out := make([]string, 0, len(lines))

	for _, line := range lines {
		line = strings.Join(strings.Fields(line), " ")
		if line != "" {
			out = append(out, line)
		}
	}

	return out
}

// previewInlineRuns parses inline HTML and Markdown into styled runs. It is the
// one entry point the preview uses for prose, so <b>, <code>, <a>, and <br>
// survive wrapping as styling instead of literal tags.
func previewInlineRuns(text string) []segment {
	return htmlInlineRuns(text, "")
}

// htmlInlineRuns parses inline HTML in text. base is the class an enclosing
// block has already applied, such as bold from <strong>.
func htmlInlineRuns(text, base string) []segment {
	var (
		out    []segment
		styles []string
	)

	write := func(chunk string) {
		if chunk == "" {
			return
		}

		class := base
		if len(styles) > 0 {
			class = styles[len(styles)-1]
		}

		for _, seg := range previewInline(chunk) {
			if seg.Class == "" {
				seg.Class = class
			}

			out = append(out, seg)
		}
	}

	for i := 0; i < len(text); {
		if strings.HasPrefix(text[i:], "<!--") {
			end := strings.Index(text[i+len("<!--"):], "-->")
			if end < 0 {
				break
			}

			i += len("<!--") + end + len("-->")

			continue
		}

		if text[i] != '<' {
			next := strings.IndexByte(text[i:], '<')
			if next < 0 {
				next = len(text) - i
			}

			write(text[i : i+next])
			i += next

			continue
		}

		end, name, rest, closing, ok := htmlTagAt(text, i)
		if !ok {
			write("<")
			i++

			continue
		}

		if isHTMLHiddenTag(name) {
			if !closing {
				i = htmlSkipRawElement(text, i, name)

				continue
			}

			i = end

			continue
		}

		switch {
		case name == "br":
			out = append(out, segment{Text: "\n"})
		case name == "img" && !closing:
			if alt := strings.TrimSpace(htmlAttr(rest, "alt")); alt != "" {
				write(alt)
			}
		default:
			if class := htmlTagClass(name); class != "" {
				switch {
				case closing:
					styles = popHTMLStyle(styles, class)
				case !strings.HasSuffix(strings.TrimSpace(rest), "/"):
					styles = append(styles, class)
				}
			}
		}

		i = end
	}

	return out
}

// htmlTagClass maps an inline tag to the token class it applies.
func htmlTagClass(name string) string {
	switch name {
	case "b", "strong", "mark":
		return classBold
	case "i", "em", "cite", "var":
		return classItal
	case "code", "kbd", "samp", "tt":
		return classCode
	case "a":
		return classLink
	case "del", "s":
		return classCom
	}

	return ""
}

// popHTMLStyle removes one class from the open inline styles.
func popHTMLStyle(styles []string, class string) []string {
	for i := len(styles) - 1; i >= 0; i-- {
		if styles[i] == class {
			return append(styles[:i], styles[i+1:]...)
		}
	}

	return styles
}

// htmlPreviewBlock is one parsed block of raw HTML before it is wrapped to
// preview rows. A table block carries cells that appendTableBlock renders.
type htmlPreviewBlock struct {
	Kind   string
	Marker string
	Runs   []segment
	Table  *htmlPreviewTable
}

// htmlPreviewTable is the header and body of a raw HTML table, plus any text
// that sat outside its cells so a rowless table still renders as prose.
type htmlPreviewTable struct {
	Header []string
	Rows   [][]string
	text   strings.Builder
}

// htmlPreviewList tracks one open ordered or unordered list.
type htmlPreviewList struct {
	ordered bool
	count   int
}

// htmlPreviewParser turns a raw HTML fragment into preview blocks. It is a
// small state machine: block elements start and flush blocks, table tags fill
// cells, and inline tags push and pop token classes.
type htmlPreviewParser struct {
	blocks  []htmlPreviewBlock
	current *htmlPreviewBlock
	styles  []string
	lists   []htmlPreviewList
	table   *htmlPreviewTable
	row     []string
	rowHead bool
	cell    strings.Builder
	inCell  bool
}

// htmlPreviewBlocks parses raw HTML into structured blocks.
func htmlPreviewBlocks(raw string) []htmlPreviewBlock {
	raw = stripHTMLHidden(raw)

	p := &htmlPreviewParser{}

	for i := 0; i < len(raw); {
		if raw[i] != '<' {
			next := strings.IndexByte(raw[i:], '<')
			if next < 0 {
				next = len(raw) - i
			}

			p.text(raw[i : i+next])
			i += next

			continue
		}

		if strings.HasPrefix(raw[i:], "<!--") {
			end := strings.Index(raw[i+len("<!--"):], "-->")
			if end < 0 {
				break
			}

			i += len("<!--") + end + len("-->")

			continue
		}

		end, name, rest, closing, ok := htmlTagAt(raw, i)
		if !ok {
			p.text("<")
			i++

			continue
		}

		selfClosing := strings.HasSuffix(strings.TrimSpace(rest), "/")
		p.tag(name, rest, closing, selfClosing)
		i = end
	}

	p.flush()
	p.finishTable()

	return p.blocks
}

// tag applies one HTML tag to the parser state.
func (p *htmlPreviewParser) tag(name, rest string, closing, selfClosing bool) {
	switch name {
	case "br":
		if p.table != nil {
			if p.inCell {
				p.cell.WriteString(" ")
			}

			return
		}

		if p.current == nil {
			p.current = &htmlPreviewBlock{Kind: "p"}
		}

		p.current.Runs = append(p.current.Runs, segment{Text: "\n"})

		return
	case "img":
		if !closing {
			if alt := strings.TrimSpace(decodeHTMLEntities(htmlAttr(rest, "alt"))); alt != "" {
				p.text(alt)
			}
		}

		return
	case "hr":
		if !closing {
			p.flush()
			p.blocks = append(p.blocks, htmlPreviewBlock{Kind: "rule"})
		}

		return
	case "table":
		if closing {
			p.finishTable()
		} else {
			p.openTable()
		}

		return
	}

	if p.table != nil {
		p.tableTag(name, closing)

		return
	}

	if isHTMLBlockTag(name) {
		if closing {
			p.closeBlock(name)
		} else if !selfClosing {
			p.openBlock(name)
		}

		return
	}

	if class := htmlTagClass(name); class != "" {
		switch {
		case closing:
			p.styles = popHTMLStyle(p.styles, class)
		case !selfClosing:
			p.styles = append(p.styles, class)
		}
	}
}

// text appends one text run to the current block or table cell.
func (p *htmlPreviewParser) text(chunk string) {
	if chunk == "" {
		return
	}

	if p.table != nil {
		if p.inCell {
			p.cell.WriteString(chunk)
		} else {
			p.table.text.WriteString(chunk)
		}

		return
	}

	if p.current == nil {
		p.current = &htmlPreviewBlock{Kind: "p"}
	}

	class := ""
	if len(p.styles) > 0 {
		class = p.styles[len(p.styles)-1]
	}

	p.current.Runs = append(p.current.Runs, htmlInlineRuns(chunk, class)...)
}

// openBlock starts the block a block-level opening tag introduces.
func (p *htmlPreviewParser) openBlock(name string) {
	switch name {
	case "ul", "ol":
		p.flush()
		p.lists = append(p.lists, htmlPreviewList{ordered: name == "ol"})
	case "li":
		p.flush()

		marker := "•"
		if n := len(p.lists); n > 0 && p.lists[n-1].ordered {
			p.lists[n-1].count++
			marker = strconv.Itoa(p.lists[n-1].count) + "."
		}

		p.current = &htmlPreviewBlock{Kind: "li", Marker: marker}
	case "summary":
		p.start("summary", "▸")
	case "blockquote":
		p.start("quote", "")
	case "h1", "h2", "h3", "h4", "h5", "h6":
		p.start(name, "")
	default:
		p.start("p", "")
	}
}

// closeBlock flushes the block a block-level closing tag ends.
func (p *htmlPreviewParser) closeBlock(name string) {
	p.flush()

	if name == "ul" || name == "ol" {
		if n := len(p.lists); n > 0 {
			p.lists = p.lists[:n-1]
		}
	}
}

// start flushes the current block and begins a new one.
func (p *htmlPreviewParser) start(kind, marker string) {
	p.flush()
	p.current = &htmlPreviewBlock{Kind: kind, Marker: marker}
}

// flush appends the current block when it has visible text.
func (p *htmlPreviewParser) flush() {
	if p.current == nil {
		return
	}

	block := *p.current
	p.current = nil

	if block.Kind == "" {
		block.Kind = "p"
	}

	if !segmentsHaveText(block.Runs) {
		return
	}

	p.blocks = append(p.blocks, block)
}

// openTable starts a raw HTML table.
func (p *htmlPreviewParser) openTable() {
	p.flush()

	p.table = &htmlPreviewTable{}
	p.row = nil
	p.rowHead = false
	p.inCell = false
}

// tableTag applies one tag while a raw HTML table is open.
func (p *htmlPreviewParser) tableTag(name string, closing bool) {
	switch name {
	case "tr":
		if closing {
			p.finishRow()
		} else {
			p.finishRow()
			p.row = nil
			p.rowHead = false
			p.inCell = false
		}
	case "td", "th":
		if closing {
			p.finishCell()
		} else {
			p.finishCell()

			p.inCell = true
			p.cell.Reset()

			if name == "th" {
				p.rowHead = true
			}
		}
	}
}

// finishCell appends the current cell to its row.
func (p *htmlPreviewParser) finishCell() {
	if !p.inCell {
		return
	}

	p.row = append(p.row, strings.Join(strings.Fields(decodeHTMLEntities(p.cell.String())), " "))
	p.cell.Reset()
	p.inCell = false
}

// finishRow appends the current row to the table header or body.
func (p *htmlPreviewParser) finishRow() {
	p.finishCell()

	if len(p.row) == 0 {
		p.row = nil

		return
	}

	if p.table != nil && p.table.Header == nil && p.rowHead {
		p.table.Header = p.row
	} else if p.table != nil {
		p.table.Rows = append(p.table.Rows, p.row)
	}

	p.row = nil
}

// finishTable promotes the first row when the table has no header cells.
func (p *htmlPreviewParser) finishTable() {
	if p.table == nil {
		return
	}

	p.finishRow()

	table := p.table
	p.table = nil

	if len(table.Header) == 0 && len(table.Rows) > 0 {
		table.Header = table.Rows[0]
		table.Rows = table.Rows[1:]
	}

	if len(table.Header) == 0 && len(table.Rows) == 0 {
		if text := strings.TrimSpace(table.text.String()); text != "" {
			p.blocks = append(p.blocks, htmlPreviewBlock{Kind: "p", Runs: htmlInlineRuns(text, "")})
		}

		return
	}

	p.blocks = append(p.blocks, htmlPreviewBlock{Table: table})
}

// htmlAttr returns one attribute value from a tag body.
func htmlAttr(rest, want string) string {
	lower := strings.ToLower(rest)

	for pos := 0; pos < len(lower); {
		idx := strings.Index(lower[pos:], want)
		if idx < 0 {
			return ""
		}

		idx += pos
		after := idx + len(want)

		if (idx > 0 && isHTMLNameChar(lower[idx-1])) || (after < len(lower) && isHTMLNameChar(lower[after])) {
			pos = after

			continue
		}

		for after < len(lower) && isHTMLSpace(lower[after]) {
			after++
		}

		if after >= len(lower) || lower[after] != '=' {
			pos = after

			continue
		}

		after++
		for after < len(lower) && isHTMLSpace(lower[after]) {
			after++
		}

		if after >= len(lower) {
			return ""
		}

		if quote := lower[after]; quote == '"' || quote == '\'' {
			end := strings.IndexByte(rest[after+1:], quote)
			if end < 0 {
				return ""
			}

			return rest[after+1 : after+1+end]
		}

		end := after
		for end < len(rest) && !isHTMLSpace(rest[end]) && rest[end] != '>' {
			end++
		}

		return rest[after:end]
	}

	return ""
}

// isHTMLNameStart reports whether c can start an HTML tag name.
func isHTMLNameStart(c byte) bool {
	return c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z'
}

// isHTMLNameChar reports whether c can continue an HTML tag or attribute name.
func isHTMLNameChar(c byte) bool {
	return isHTMLNameStart(c) || c >= '0' && c <= '9' || c == '-' || c == '_'
}

// isHTMLSpace reports whether c is HTML whitespace.
func isHTMLSpace(c byte) bool {
	return c == ' ' || c == '\t' || c == '\n' || c == '\r'
}

// decodeHTMLEntities turns the basic named and numeric entities into text.
func decodeHTMLEntities(text string) string {
	if !strings.ContainsRune(text, '&') {
		return text
	}

	var b strings.Builder

	b.Grow(len(text))

	for i := 0; i < len(text); {
		if text[i] != '&' {
			b.WriteByte(text[i])
			i++

			continue
		}

		end := strings.IndexByte(text[i:], ';')
		if end < 0 || end > previewEntityBytes {
			b.WriteByte('&')
			i++

			continue
		}

		value, ok := htmlEntityValue(text[i+1 : i+end])
		if !ok {
			b.WriteByte('&')
			i++

			continue
		}

		b.WriteString(value)
		i += end + 1
	}

	return b.String()
}

// htmlEntityValue returns the replacement for one entity name.
func htmlEntityValue(name string) (string, bool) {
	switch name {
	case "amp":
		return "&", true
	case "lt":
		return "<", true
	case "gt":
		return ">", true
	case "quot":
		return `"`, true
	case "apos":
		return "'", true
	case "nbsp":
		return " ", true
	case "ndash":
		return "–", true
	case "mdash":
		return "—", true
	case "hellip":
		return "…", true
	}

	if !strings.HasPrefix(name, "#") {
		return "", false
	}

	digits := name[1:]
	base := previewEntityBase

	if strings.HasPrefix(digits, "x") || strings.HasPrefix(digits, "X") {
		digits = digits[1:]
		base = previewEntityHex
	}

	if digits == "" {
		return "", false
	}

	value, err := strconv.ParseInt(digits, base, previewEntityBits)
	if err != nil || value <= 0 || !utf8.ValidRune(rune(value)) {
		return "", false
	}

	return string(rune(value)), true
}
