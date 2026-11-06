// Package ui provides an experimental fluent HTML/CSS builder for OwnFrame.
//
// The builder produces ordinary HTML that OwnFrame passes to Blinkless. It does
// not implement layout, replace templates, or run a JavaScript/DOM environment.
package ui

import (
	"context"
	"fmt"
	"slices"
	"sort"
	"strconv"
	"strings"
	"sync"
)

// Element is one node in an experimental declarative UI tree.
//
// Builder calls mutate the receiver. Do not share a tree between goroutines
// while editing or rendering it.
type Element struct {
	tag      string
	attrs    *attrSet
	styles   styleSet
	text     string
	children []*Element
	onClick  func(context.Context) error
	err      error
}

// Tag creates an element. Unsupported HTML tags are rejected by Render.
func Tag(name string) *Element { return &Element{tag: strings.ToLower(strings.TrimSpace(name))} }

func Div() *Element                { return Tag("div") }
func Span() *Element               { return Tag("span") }
func P() *Element                  { return Tag("p") }
func Button(label string) *Element { return Tag("button").Text(label) }
func Ul() *Element                 { return Tag("ul") }
func Li() *Element                 { return Tag("li") }
func Input() *Element              { return Tag("input") }
func Section() *Element            { return Tag("section") }
func Nav() *Element                { return Tag("nav") }
func Header() *Element             { return Tag("header") }

// Text constructs a span with escaped text, suitable for use in Children.
func Text(value string) *Element { return Span().Text(value) }

// Text replaces the receiver's own text, which is written before its children.
func (e *Element) Text(value string) *Element {
	if e != nil {
		e.text = value
	}
	return e
}

// Children appends the given nodes in order. Nil children are ignored.
func (e *Element) Children(children ...*Element) *Element {
	if e != nil {
		// Grow once for the whole call instead of letting append double the
		// backing array from cap 1 through 2, 4, and 8.
		e.children = slices.Grow(e.children, len(children))
		for _, child := range children {
			if child != nil {
				e.children = append(e.children, child)
			}
		}
	}
	return e
}

// inlineAttrEntries is the number of attributes an element keeps without
// another allocation. Elements with no attributes carry no set at all.
const inlineAttrEntries = 4

// attrEntry is one HTML attribute.
type attrEntry struct {
	name  string
	value string
}

// attrSet stores an element's attributes in insertion order. It is allocated
// on the first Attr call, and the first inlineAttrEntries fit in that one
// allocation. Reads are nil-safe so an element with no attributes needs no
// empty set.
type attrSet struct {
	inline [inlineAttrEntries]attrEntry
	extra  []attrEntry
	n      int
}

func (s *attrSet) len() int {
	if s == nil {
		return 0
	}
	return s.n
}

// at returns the entry at index i, which must be within len.
func (s *attrSet) at(i int) attrEntry {
	if i < len(s.inline) {
		return s.inline[i]
	}
	return s.extra[i-len(s.inline)]
}

// get returns the value of the named attribute.
func (s *attrSet) get(name string) (string, bool) {
	if s == nil {
		return "", false
	}
	for i := 0; i < s.n; i++ {
		if entry := s.at(i); entry.name == name {
			return entry.value, true
		}
	}
	return "", false
}

// value returns the named attribute, or "" when it is absent.
func (s *attrSet) value(name string) string {
	value, _ := s.get(name)
	return value
}

// set adds an attribute, replacing one with the same name.
func (s *attrSet) set(name, value string) {
	for i := 0; i < s.n; i++ {
		if s.at(i).name != name {
			continue
		}
		if i < len(s.inline) {
			s.inline[i].value = value
		} else {
			s.extra[i-len(s.inline)].value = value
		}
		return
	}
	entry := attrEntry{name: name, value: value}
	if s.n < len(s.inline) {
		s.inline[s.n] = entry
	} else {
		s.extra = append(s.extra, entry)
	}
	s.n++
}

// equal reports whether two sets hold the same attributes in any order.
func (s *attrSet) equal(other *attrSet) bool {
	if s.len() != other.len() {
		return false
	}
	for i := 0; i < s.len(); i++ {
		entry := s.at(i)
		if value, ok := other.get(entry.name); !ok || value != entry.value {
			return false
		}
	}
	return true
}

// Attr adds or replaces an HTML attribute. Inline styles must use generated style methods or CSSRaw.
func (e *Element) Attr(name, value string) *Element {
	if e == nil {
		return e
	}
	name = strings.ToLower(strings.TrimSpace(name))
	if !validAttribute(name) || name == "style" || strings.HasPrefix(name, "on") {
		e.recordError(fmt.Errorf("ui: disallowed HTML attribute %q", name))
		return e
	}
	if e.attrs == nil {
		e.attrs = &attrSet{}
	}
	e.attrs.set(name, value)
	return e
}

func (e *Element) ID(id string) *Element       { return e.Attr("id", id) }
func (e *Element) Class(class string) *Element { return e.Attr("class", class) }

// OnClick registers an OwnFrame callback. NewPage connects it to the host's
// existing data-action / Handlers.Click path; Render alone only emits markup.
func (e *Element) OnClick(handler func(context.Context) error) *Element {
	if e != nil {
		e.onClick = handler
	}
	return e
}

func (e *Element) recordError(err error) {
	if e.err == nil {
		e.err = err
	}
}

// Render serializes a component tree to an escaped HTML fragment. The caller
// can use the fragment alongside existing HTML/CSS templates. Use NewPage
// instead when the tree has OnClick callbacks that must be connected.
func Render(root *Element) (string, error) {
	r := getRenderer()
	defer putRenderer(r)
	r.output.Grow(estimateSize(root))
	if err := r.visit(root, 0); err != nil {
		return "", err
	}
	return r.output.String(), nil
}

type renderer struct {
	output  strings.Builder
	actions map[string]func(context.Context) error
	// stack is the current visit path. It replaces the old active map:
	// membership is a linear scan, which stays cheap because depth is
	// capped, and it needs no per-node map operations or defers.
	stack []*Element
	// keys is scratch for sorted key emission, shared by the style and
	// attribute phases of one node. A parent finishes emitting before it
	// recurses, so children reusing the scratch never disturb it.
	keys []string
	// scratch holds one formatted numeric style value at a time. The
	// element stores lengths as numbers; they become text only here.
	scratch    []byte
	nextID     int
	scope      string
	identities map[string]int
}

func newRenderer() *renderer {
	return &renderer{
		actions:    make(map[string]func(context.Context) error),
		identities: make(map[string]int),
	}
}

// rendererPool reuses builders across renders. The output buffer keeps its
// capacity (one 600KB backing instead of regrowing per render); maps are
// cleared on checkout so no state leaks between documents.
var rendererPool = sync.Pool{New: func() any { return newRenderer() }}

// reset prepares a pooled renderer for a new document. Callers that keep
// r.actions (Mount, Update) must detach it first; reset clears whatever map
// the renderer currently holds.
func (r *renderer) reset() {
	r.output.Reset()
	// A view with no OnClick handlers can pool a renderer whose actions map
	// was detached as nil. Recreate it on checkout so visit can always store
	// a callback.
	if r.actions == nil {
		r.actions = make(map[string]func(context.Context) error)
	} else {
		clear(r.actions)
	}
	r.stack = r.stack[:0]
	r.keys = r.keys[:0]
	r.scratch = r.scratch[:0]
	r.nextID = 0
	r.scope = ""
	clear(r.identities)
}

// getRenderer checks out a reset renderer. Put it back with putRenderer
// after detaching anything the caller keeps (the actions map on success).
func getRenderer() *renderer {
	r := rendererPool.Get().(*renderer) //nolint:forcetypeassert // pool holds only *renderer
	r.reset()
	return r
}

// putRenderer returns a renderer whose maps the caller has detached or
// cleared. The output buffer and key scratch keep their capacity.
func putRenderer(r *renderer) {
	if r == nil {
		return
	}
	rendererPool.Put(r)
}

const (
	maxDepth = 1024
	// estimateNodeOverhead budgets the markup around one node: the open
	// and close tags, quotes, separators, and synthesized attributes.
	estimateNodeOverhead = 64
	// estimateAttrOverhead budgets the `="..."` separators around one
	// attribute or style declaration value.
	estimateAttrOverhead = 8
	// maxEstimateNodes caps the render size pre-walk so a cyclic tree
	// still terminates quickly; the walk stops counting, it never fails.
	// The cap covers a 9k-node document, so a page-sized tree grows the
	// output buffer once instead of doubling it partway through.
	maxEstimateNodes = 16384
)

// estimateSize bounds the fragment length with one cheap pointer walk, so
// Render can Grow the output builder once instead of regrowing it per node.
// Shared subtrees may be counted twice and cyclic trees stop at the cap;
// both only make the hint larger, never wrong.
func estimateSize(root *Element) int {
	if root == nil {
		return 0
	}
	total := 0
	counted := 0
	stack := []*Element{root}
	var scratch []byte
	for len(stack) > 0 && counted < maxEstimateNodes {
		e := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if e == nil {
			continue
		}
		counted++
		total += estimateNodeOverhead + len(e.tag)*2 + len(e.text)
		for i := 0; i < e.attrs.len(); i++ {
			entry := e.attrs.at(i)
			total += len(entry.name) + len(entry.value) + estimateAttrOverhead
		}
		for i := 0; i < e.styles.len(); i++ {
			entry := e.styles.at(i)
			total += len(entry.name) + estimateAttrOverhead
			if entry.unit == unitText {
				total += len(entry.text)
				continue
			}
			scratch = appendScalar(scratch[:0], entry.num, entry.unit)
			total += len(scratch)
		}
		stack = append(stack, e.children...)
	}
	return total
}

// isVoidTag reports whether tag is written without a closing tag.
func isVoidTag(tag string) bool {
	switch tag {
	case "input", "img", "br", "hr":
		return true
	default:
		return false
	}
}

// actionName is the data-action value for the n-th OnClick node in document
// order. The incremental diff reuses it to rebuild the callback registry
// without emitting HTML.
func actionName(n int) string {
	return "ownframe-ui-" + strconv.Itoa(n)
}

func supportedTag(tag string) bool {
	switch tag {
	case "div", "span", "p", "br", "hr", "h1", "h2",
		"h3", "h4", "h5", "h6", "ul", "ol", "li",
		"table", "thead", "tbody", "tfoot", "tr", "th", "td",
		"caption", "img", "a", "strong", "em", "b", "i",
		"u", "small", "pre", "code", "blockquote", "header", "footer",
		"main", "section", "article", "aside", "nav", "input", "textarea",
		"select", "option", "optgroup", "button", "label", "form":
		return true
	default:
		return false
	}
}

func (r *renderer) visit(e *Element, depth int) error {
	if e == nil {
		return nil
	}
	if e.err != nil {
		return e.err
	}
	if depth >= maxDepth {
		return fmt.Errorf("ui: element nesting exceeds %d", maxDepth)
	}
	if !supportedTag(e.tag) {
		return fmt.Errorf("ui: unsupported HTML tag %q", e.tag)
	}
	if slices.Contains(r.stack, e) {
		return fmt.Errorf("ui: cycle in component tree at <%s>", e.tag)
	}
	void := isVoidTag(e.tag)
	if void && (e.text != "" || len(e.children) != 0) {
		return fmt.Errorf("ui: void HTML element <%s> cannot have children or text", e.tag)
	}

	r.stack = append(r.stack, e)

	action, id := "", ""
	hasAction := false
	if e.onClick != nil {
		if _, exists := e.attrs.get("data-action"); exists {
			r.stack = r.stack[:len(r.stack)-1]
			return fmt.Errorf("ui: cannot combine OnClick with a custom data-action")
		}
		r.nextID++
		action = actionName(r.nextID)
		hasAction = true
		r.actions[action] = e.onClick
	} else if dataAction := e.attrs.value("data-action"); dataAction != "" {
		action = dataAction
		hasAction = true
	}
	if userID := e.attrs.value("id"); userID != "" {
		id = userID
	} else if hasAction {
		id = r.controlID(e)
	}

	previousScope := r.scope
	if id != "" {
		r.scope = id
	}

	// Style keys occupy the head of the scratch so the style value can be
	// serialized directly in sorted order; attribute names follow and are
	// sorted as their own region. Single-key regions skip the sort.
	r.keys = r.keys[:0]
	styleCount := e.styles.len()
	for i := 0; i < styleCount; i++ {
		r.keys = append(r.keys, e.styles.at(i).name)
	}
	if styleCount > 1 {
		sort.Strings(r.keys)
	}
	attrStart := len(r.keys)
	for i := 0; i < e.attrs.len(); i++ {
		r.keys = append(r.keys, e.attrs.at(i).name)
	}
	// Synthesized attributes read back the matching value below, so they
	// sort with the author's attributes instead of living in a copied map.
	if hasAction {
		if _, ok := e.attrs.get("data-action"); !ok {
			r.keys = append(r.keys, "data-action")
		}
	}
	if id != "" {
		if _, ok := e.attrs.get("id"); !ok {
			r.keys = append(r.keys, "id")
		}
	}
	if styleCount != 0 {
		if _, ok := e.attrs.get("style"); !ok {
			r.keys = append(r.keys, "style")
		}
	}
	if len(r.keys)-attrStart > 1 {
		sort.Strings(r.keys[attrStart:])
	}

	r.output.WriteByte('<')
	r.output.WriteString(e.tag)
	for _, name := range r.keys[attrStart:] {
		r.output.WriteByte(' ')
		r.output.WriteString(name)
		r.output.WriteString(`="`)
		switch name {
		case "data-action":
			writeEscaped(&r.output, action)
		case "id":
			writeEscaped(&r.output, id)
		case "style":
			for _, sname := range r.keys[:styleCount] {
				entry, _ := e.styles.get(sname)
				writeEscaped(&r.output, sname)
				r.output.WriteByte(':')
				if entry.unit == unitText {
					writeEscaped(&r.output, entry.text)
				} else {
					r.scratch = entry.appendValue(r.scratch[:0])
					_, _ = r.output.Write(r.scratch)
				}
				r.output.WriteByte(';')
			}
		default:
			writeEscaped(&r.output, e.attrs.value(name))
		}
		r.output.WriteByte('"')
	}
	r.output.WriteByte('>')
	if !void {
		writeEscaped(&r.output, e.text)
		for _, child := range e.children {
			if err := r.visit(child, depth+1); err != nil {
				r.scope = previousScope
				r.stack = r.stack[:len(r.stack)-1]
				return err
			}
		}
		r.output.WriteString("</")
		r.output.WriteString(e.tag)
		r.output.WriteByte('>')
	}
	r.scope = previousScope
	r.stack = r.stack[:len(r.stack)-1]
	return nil
}

// writeEscaped streams s into sb in one pass, escaping the HTML metacharacters
// and the template brace. Byte-identical to applying html.EscapeString and then
// replacing "{": every replacement target is ASCII, so a byte loop never
// splits a multi-byte rune, and replacements contain no escapable bytes.
func writeEscaped(sb *strings.Builder, s string) {
	start := 0
	for i := 0; i < len(s); i++ {
		var replacement string
		switch s[i] {
		case '&':
			replacement = "&amp;"
		case '<':
			replacement = "&lt;"
		case '>':
			replacement = "&gt;"
		case '"':
			replacement = "&#34;"
		case '\'':
			replacement = "&#39;"
		case '{':
			replacement = "&#123;"
		default:
			continue
		}
		sb.WriteString(s[start:i])
		sb.WriteString(replacement)
		start = i + 1
	}
	sb.WriteString(s[start:])
}

// Escape template braces as character references too. The serialized HTML is
// given to html/template as a template *source*, so raw "{{" text must never
// turn into an executable Go template action.
func escape(value string) string {
	var sb strings.Builder
	sb.Grow(len(value))
	writeEscaped(&sb, value)
	return sb.String()
}

const (
	// fnvOffset64 and fnvPrime64 are the FNV-1a 64-bit parameters, matching
	// hash/fnv New64a. controlID folds bytes manually so it never allocates
	// hash state, format buffers, or string conversions.
	fnvOffset64 = 14695981039346656037
	fnvPrime64  = 1099511628211
)

// hashString folds s into an FNV-1a 64-bit hash.
func hashString(h uint64, s string) uint64 {
	for i := 0; i < len(s); i++ {
		h ^= uint64(s[i])
		h *= fnvPrime64
	}
	return h
}

// controlID derives the generated id for a control.
//
// Stability contract: the id hashes the enclosing scope, the tag, the
// action-ish attributes, and the element's own subtree text, so inserting or
// removing a conditional sibling earlier in the document does not renumber
// the surviving controls; genuine duplicates share a base and are numbered in
// document order. The full subtree walk is required for that stability: two
// same-labeled buttons in different cards must hash differently even when
// their own attributes match, which hashing only the node itself would miss.
func (r *renderer) controlID(e *Element) string {
	h := hashString(fnvOffset64, r.scope)
	h = hashString(h, " ")
	h = hashString(h, e.tag)
	h = hashString(h, " ")
	h = hashString(h, e.attrs.value("data-action"))
	h = hashString(h, " ")
	h = hashString(h, e.attrs.value("name"))
	h = hashString(h, " ")
	h = hashString(h, e.attrs.value("aria-label"))
	h = hashString(h, "\n")
	seen := map[*Element]bool{}
	var text func(*Element)
	text = func(node *Element) {
		if node == nil || seen[node] {
			return
		}
		seen[node] = true
		h = hashString(h, node.tag)
		h = hashString(h, " ")
		h = hashString(h, node.text)
		h = hashString(h, "\n")
		for _, child := range node.children {
			text(child)
		}
	}
	text(e)
	base := "ownframe-ui-control-" + strconv.FormatUint(h, 16)
	r.identities[base]++
	return base + "-" + strconv.Itoa(r.identities[base])
}

// Document assembly pieces for a complete builder page. Mount and
// View.Update should adopt Document; the older document helper in mount.go
// concatenates the same pieces with intermediate strings.
const (
	docHead = "<!doctype html><html><head><meta charset=\"utf-8\"><title>"
	docMid  = "</title><style>html,body{margin:0;padding:0}</style></head><body>"
	docTail = "</body></html>"
)

// Document assembles a complete builder page in a single write sequence:
// one Grow for the known pieces, the title escaped straight into the buffer,
// then the fragment. It emits byte-identical output to the mount.go document
// helper while copying the payload once instead of two to three times.
func Document(title, fragment string) string {
	var sb strings.Builder
	sb.Grow(len(docHead) + len(title) + len(docMid) + len(fragment) + len(docTail))
	sb.WriteString(docHead)
	writeEscaped(&sb, title)
	sb.WriteString(docMid)
	sb.WriteString(fragment)
	sb.WriteString(docTail)
	return sb.String()
}

func validAttribute(name string) bool {
	if len(name) == 0 || (name[0] < 'a' || name[0] > 'z') {
		return false
	}
	for i := 1; i < len(name); i++ {
		c := name[i]
		if (c < 'a' || c > 'z') && (c < '0' || c > '9') && c != '_' && c != '-' && c != ':' && c != '.' {
			return false
		}
	}
	return true
}
