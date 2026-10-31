// Package ui provides an experimental fluent HTML/CSS builder for OwnFrame.
//
// The builder produces ordinary HTML that OwnFrame passes to Blinkless. It does
// not implement layout, replace templates, or run a JavaScript/DOM environment.
package ui

import (
	"context"
	"fmt"
	"html"
	"sort"
	"strings"
)

// Element is one node in an experimental declarative UI tree.
//
// Builder calls mutate the receiver. Do not share a tree between goroutines
// while editing or rendering it.
type Element struct {
	tag      string
	attrs    map[string]string
	styles   map[string]string
	text     string
	children []*Element
	onClick  func(context.Context) error
	err      error
}

// Tag creates an element. Unsupported HTML tags are rejected by Render.
func Tag(name string) *Element { return &Element{tag: strings.ToLower(strings.TrimSpace(name))} }

func Div() *Element     { return Tag("div") }
func Span() *Element    { return Tag("span") }
func P() *Element       { return Tag("p") }
func Button(label string) *Element { return Tag("button").Text(label) }
func Ul() *Element      { return Tag("ul") }
func Li() *Element      { return Tag("li") }
func Input() *Element   { return Tag("input") }
func Section() *Element { return Tag("section") }
func Nav() *Element     { return Tag("nav") }
func Header() *Element  { return Tag("header") }

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
		for _, child := range children {
			if child != nil {
				e.children = append(e.children, child)
			}
		}
	}
	return e
}

// Attr adds or replaces an HTML attribute. Inline styles must use CSS().
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
		e.attrs = make(map[string]string)
	}
	e.attrs[name] = value
	return e
}

func (e *Element) ID(id string) *Element       { return e.Attr("id", id) }
func (e *Element) Class(class string) *Element { return e.Attr("class", class) }

// CSS adds a CSS property by its standard, vendor-prefixed, or matrix name.
// All 555 properties marked Implemented in the compatibility matrix are
// accepted here. Values are passed to Blinkless, which decides their semantics.
func (e *Element) CSS(property, value string) *Element {
	if e == nil {
		return e
	}
	property = strings.ToLower(strings.TrimSpace(property))
	if !SupportsProperty(property) {
		e.recordError(fmt.Errorf("ui: CSS property %q is not implemented by the pinned compatibility matrix", property))
		return e
	}
	if strings.TrimSpace(value) == "" || strings.IndexByte(value, 0) >= 0 {
		e.recordError(fmt.Errorf("ui: invalid value for CSS property %q", property))
		return e
	}
	if e.styles == nil {
		e.styles = make(map[string]string)
	}
	e.styles[property] = value
	return e
}

// Style is an alias for CSS, useful for unfamiliar and vendor-prefixed names.
func (e *Element) Style(property, value string) *Element { return e.CSS(property, value) }

// Styles applies a map of CSS properties in deterministic key order.
func (e *Element) Styles(values map[string]string) *Element {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		e.CSS(key, values[key])
	}
	return e
}

// Common fluent helpers all use the same matrix-backed CSS path.
func (e *Element) BackgroundColor(value string) *Element { return e.CSS("background-color", value) }
func (e *Element) Padding(value string) *Element         { return e.CSS("padding", value) }
func (e *Element) BorderRadius(value string) *Element    { return e.CSS("border-radius", value) }
func (e *Element) Display(value string) *Element         { return e.CSS("display", value) }
func (e *Element) FlexDirection(value string) *Element   { return e.CSS("flex-direction", value) }
func (e *Element) AlignItems(value string) *Element      { return e.CSS("align-items", value) }
func (e *Element) Gap(value string) *Element             { return e.CSS("gap", value) }
func (e *Element) Color(value string) *Element           { return e.CSS("color", value) }
func (e *Element) FontSize(value string) *Element        { return e.CSS("font-size", value) }
func (e *Element) FontWeight(value string) *Element      { return e.CSS("font-weight", value) }
func (e *Element) Margin(value string) *Element          { return e.CSS("margin", value) }
func (e *Element) Width(value string) *Element           { return e.CSS("width", value) }
func (e *Element) Height(value string) *Element          { return e.CSS("height", value) }

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
	r := newRenderer()
	if err := r.visit(root, 0); err != nil {
		return "", err
	}
	return r.output.String(), nil
}

type renderer struct {
	output  strings.Builder
	actions map[string]func(context.Context) error
	active  map[*Element]bool
	nextID  int
}

func newRenderer() *renderer {
	return &renderer{
		actions: make(map[string]func(context.Context) error),
		active:  make(map[*Element]bool),
	}
}

const maxDepth = 1024

func supportedTag(tag string) bool {
	switch tag {
	case "div", "span", "p", "br", "hr", "h1", "h2",
		"h3", "h4", "h5", "h6", "ul", "ol", "li",
		"table", "thead", "tbody", "tfoot", "tr", "th", "td",
		"caption", "img", "a", "strong", "em", "b", "i",
		"u", "small", "pre", "code", "blockquote", "header", "footer",
		"main", "section", "article", "aside", "nav", "input", "textarea",
		"select", "button", "label", "form":
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
	if r.active[e] {
		return fmt.Errorf("ui: cycle in component tree at <%s>", e.tag)
	}
	r.active[e] = true
	defer delete(r.active, e)

	void := e.tag == "input" || e.tag == "img" || e.tag == "br" || e.tag == "hr"
	if void && (e.text != "" || len(e.children) != 0) {
		return fmt.Errorf("ui: void HTML element <%s> cannot have children or text", e.tag)
	}

	attrs := make(map[string]string, len(e.attrs)+3)
	for name, value := range e.attrs {
		attrs[name] = value
	}
	if e.onClick != nil {
		if _, exists := attrs["data-action"]; exists {
			return fmt.Errorf("ui: cannot combine OnClick with a custom data-action")
		}
		r.nextID++
		action := fmt.Sprintf("ownframe-ui-%d", r.nextID)
		attrs["data-action"] = action
		if attrs["id"] == "" {
			attrs["id"] = action
		}
		r.actions[action] = e.onClick
	}
	if len(e.styles) != 0 {
		keys := make([]string, 0, len(e.styles))
		for name := range e.styles {
			keys = append(keys, name)
		}
		sort.Strings(keys)
		var style strings.Builder
		for _, name := range keys {
			style.WriteString(name)
			style.WriteByte(':')
			style.WriteString(e.styles[name])
			style.WriteByte(';')
		}
		attrs["style"] = style.String()
	}

	r.output.WriteByte('<')
	r.output.WriteString(e.tag)
	names := make([]string, 0, len(attrs))
	for name := range attrs {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		r.output.WriteByte(' ')
		r.output.WriteString(name)
		r.output.WriteString("=\"")
		r.output.WriteString(escape(attrs[name]))
		r.output.WriteByte('"')
	}
	r.output.WriteByte('>')
	if void {
		return nil
	}
	r.output.WriteString(escape(e.text))
	for _, child := range e.children {
		if err := r.visit(child, depth+1); err != nil {
			return err
		}
	}
	r.output.WriteString("</")
	r.output.WriteString(e.tag)
	r.output.WriteByte('>')
	return nil
}

// Escape template braces as character references too. The serialized HTML is
// given to html/template as a template *source*, so raw "{{" text must never
// turn into an executable Go template action.
func escape(value string) string {
	return strings.ReplaceAll(html.EscapeString(value), "{", "&#123;")
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
