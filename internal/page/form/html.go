package form

import (
	"html"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
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
	return RewriteCached(source, HashSource(source), spans, live, st)
}

// RewriteCached is RewriteState with the source hash threaded through, so a
// caller that already hashed the source (the page sync) pays no second hash
// and reuses the cached head offsets instead of scanning for them again.
// Output is identical to RewriteState for the same inputs.
func RewriteCached(source string, hash uint64, spans []Span, live map[string]Control, st Caret) string {
	if !hasFieldSpans(spans) {
		return source
	}

	scanMemo.rewrites.Add(1)

	var b strings.Builder
	b.Grow(len(source) + len(formCSS))

	cssAt := cachedHead(source, hash)
	kept := keepSpans(source, spans)
	if cssAt >= 0 && spanCovers(kept, cssAt) {
		cssAt = -1
	}

	prev := 0
	placed := false
	write := func(end int) {
		if !placed && cssAt >= prev && cssAt <= end {
			b.WriteString(source[prev:cssAt])
			b.WriteString(formCSS)
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
		return formCSS + b.String()
	}

	return b.String()
}

// cachedHead returns the form-style insertion offset for source: HeadOpen,
// or FindHead when there is no head open tag. The offset is memoized per
// source hash; a miss scans once and stores the offset for the next redraw
// of the same body.
func cachedHead(source string, hash uint64) int {
	scanMemo.Lock()
	for i := range scanMemo.entries {
		if e := &scanMemo.entries[i]; e.hash == hash && e.source == source && e.headValid {
			at := e.head
			scanMemo.Unlock()

			return at
		}
	}
	scanMemo.Unlock()

	at := headScan(source)

	scanMemo.Lock()
	for i := range scanMemo.entries {
		if e := &scanMemo.entries[i]; e.hash == hash && e.source == source {
			e.head, e.headValid = at, true

			break
		}
	}
	scanMemo.Unlock()

	return at
}

// headScan evaluates the style insertion offset without the cache. It backs
// the head part of the telemetry: every evaluation is one HeadOpen scan
// plus a FindHead scan only when there is no head open tag.
func headScan(source string) int {
	at := HeadOpen(source)
	scanMemo.heads.Add(1)
	if at < 0 {
		at = FindHead(source)
	}

	return at
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

const (
	// fnvOffset64 and fnvPrime64 are the FNV-1a 64-bit parameters used by
	// HashSource. Named so the lint gate does not read them as magic numbers.
	fnvOffset64 = 14695981039346656037
	fnvPrime64  = 1099511628211

	// probeCacheSize bounds the per-source scan memo. Four entries cover a
	// page alternating between a few bodies without retaining a history.
	probeCacheSize = 4
)

// HashSource hashes source with FNV-1a. The page sync hashes each executed
// body once and threads the hash through the cached scan and rewrite, so no
// pass hashes the source twice. It allocates nothing.
func HashSource(source string) uint64 {
	h := uint64(fnvOffset64)
	for i := 0; i < len(source); i++ {
		h ^= uint64(source[i])
		h *= fnvPrime64
	}

	return h
}

// probeResult memoizes one executed body: the control probe, the control
// spans, and the form-style insertion offset. Spans are a pure function of
// the source, so a repeat body reuses them without walking the source. The
// cached spans are shared read-only; callers must not mutate them.
type probeResult struct {
	hash        uint64
	source      string
	hasControls bool
	hasButtons  bool
	spans       []Span
	head        int
	headValid   bool
}

// ScanTelemetry counts source passes so tests can prove skipped passes are
// skipped: a repeat body must not move Probes, Scans, or HeadScans, and a
// span-free body must not move Rewrites.
type ScanTelemetry struct {
	Probes   uint64
	Scans    uint64
	Heads    uint64
	Rewrites uint64
}

// its pass counters. The entries key on the source hash with an exact source
// compare, so a collision degrades to a rescan instead of a wrong page. The
// counters are the only way for tests to observe skipped passes.
//
//nolint:gochecknoglobals // Bounded single-process memo (four entries) plus
var scanMemo = struct {
	sync.Mutex
	entries  [probeCacheSize]probeResult
	next     int
	probes   atomic.Uint64
	scans    atomic.Uint64
	heads    atomic.Uint64
	rewrites atomic.Uint64
}{}

// Telemetry returns the source-pass counters.
func Telemetry() ScanTelemetry {
	return ScanTelemetry{
		Probes:   scanMemo.probes.Load(),
		Scans:    scanMemo.scans.Load(),
		Heads:    scanMemo.heads.Load(),
		Rewrites: scanMemo.rewrites.Load(),
	}
}

// HasFieldSpans reports whether any span needs paintable field markup.
// Plain buttons stay in the source, so the page skips the rewrite for them.
func HasFieldSpans(spans []Span) bool { return hasFieldSpans(spans) }

// ScanSingle reads input, textarea, select, and button controls from source
// in one walk, mirroring Scan exactly, and also reports whether source holds
// a submit, button, or reset input that the button rewrite would replace.
// The button predicate matches rewriteButtons tag for tag (same open-tag
// read, same type attribute read, same comment and script skipping), so a
// false lets the caller skip that rewrite knowing it is the identity.
// Unlike Scan it performs no HasControls pre-probe: the walk is the probe.
func ScanSingle(source string) (spans []Span, buttons bool) {
	var out []Span
	i := 0
	for i < len(source) {
		next := strings.IndexByte(source[i:], '<')
		if next < 0 {
			break
		}
		i += next
		if strings.HasPrefix(source[i:], "<!--") {
			end := strings.Index(source[i+CommentOpenLen:], "-->")
			if end < 0 {
				break
			}
			i += end + CommentTotalLen

			continue
		}

		name, raw, end, ok := ReadOpen(source, i)
		if !ok {
			i++

			continue
		}

		low := strings.ToLower(name)
		switch low {
		case "script", "style":
			_, i = CloseSpan(source, end, low)
		case "input":
			switch strings.ToLower(AttrValue(raw, "type")) {
			case "submit", "button", "reset":
				buttons = true
			}
			if sp, add := makeInput(i, end, readAttrs(raw)); add {
				out = append(out, sp)
			}
			i = end
		case "textarea":
			sp, at := makeArea(source, i, end, readAttrs(raw))
			if sp.Control.ID != "" {
				out = append(out, sp)
			}
			i = at
		case "select":
			sp, at := makeSelect(source, i, end, readAttrs(raw))
			if sp.Control.ID != "" {
				out = append(out, sp)
			}
			i = at
		case "button":
			_, at := CloseSpan(source, end, low)
			a := readAttrs(raw)
			if a.id != "" {
				c := controlFrom(low, a)
				c.Type = "button"
				if a.kind == "submit" || a.kind == "reset" {
					c.Type = a.kind
				}
				out = append(out, Span{Control: c, Start: i, End: at})
			}
			i = at
		default:
			i = end
		}
	}

	return out, buttons
}

// ScanCached returns the control probe, the control spans, and the
// button-input flag for source, memoized per source hash. The first redraw
// of a body pays one probe plus one scan walk; repeats reuse the entry
// without touching the source. Entries key on the hash and validate with an
// exact source compare, so a hash collision only costs a rescan.
func ScanCached(source string, hash uint64) (bool, []Span, bool) {
	scanMemo.Lock()
	for i := range scanMemo.entries {
		if e := &scanMemo.entries[i]; e.hash == hash && e.source == source {
			has, spans, buttons := e.hasControls, e.spans, e.hasButtons
			scanMemo.Unlock()

			return has, spans, buttons
		}
	}
	scanMemo.Unlock()

	has := HasControls(source)
	scanMemo.probes.Add(1)
	spans, buttons := ScanSingle(source)
	scanMemo.scans.Add(1)

	scanMemo.Lock()
	scanMemo.entries[scanMemo.next] = probeResult{
		hash:        hash,
		source:      source,
		hasControls: has,
		hasButtons:  buttons,
		spans:       spans,
		head:        -1,
	}
	scanMemo.next = (scanMemo.next + 1) % probeCacheSize
	scanMemo.Unlock()

	return has, spans, buttons
}
