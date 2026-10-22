package window

import (
	"encoding/json"
	"fmt"
	"image/color"
	"reflect"
	"sort"
	"strconv"
	"strings"
)

// devJSONRows renders v as pretty-printed JSON rows with syntax colours.
// Structs keep declaration order, maps render keys sorted, slices render in
// order, and pointers or interfaces render their contents. An object or
// array opens on a row carrying devActJSON and the node's dotted key path,
// so a click can flip collapsed[key] to fold the node. A folded node renders
// as {...} or [...] on one row.
func devJSONRows(v any, collapsed map[string]bool) []devRow {
	p := &devJSONPrinter{collapsed: collapsed}
	p.emit(p.build(reflect.ValueOf(v), "", ""), 0, devLine{}, "", true)

	return p.rows
}

// buildStruct renders exported fields in declaration order.
func (p *devJSONPrinter) buildStruct(v reflect.Value, key, path string) *devJSONNode {
	node := &devJSONNode{kind: 'o', key: key, path: path}
	t := v.Type()

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if f.PkgPath != "" {
			continue
		}

		name, opts := devJSONTag(f.Tag.Get("json"))
		if name == "-" {
			continue
		}

		if name == "" {
			name = f.Name
		}

		fv := v.Field(i)
		if opts == "omitempty" && devJSONEmpty(fv) {
			continue
		}

		node.kids = append(node.kids, p.build(fv, name, devJSONJoin(path, name)))
	}

	return node
}

// buildArray renders elements in order, named by index.
func (p *devJSONPrinter) buildArray(v reflect.Value, key, path string) *devJSONNode {
	node := &devJSONNode{kind: 'a', key: key, path: path}

	for i := 0; i < v.Len(); i++ {
		ik := strconv.Itoa(i)
		node.kids = append(node.kids, p.build(v.Index(i), ik, devJSONJoin(path, ik)))
	}

	return node
}

// devJSONPrinter accumulates the rows for one document.
type devJSONPrinter struct {
	rows      []devRow
	collapsed map[string]bool
}

// emit appends the rows for one node. depth is the nesting level, lead is
// the member prefix, tail is a trailing comma, and root keeps the top node
// unfoldable.
func (p *devJSONPrinter) emit(n *devJSONNode, depth int, lead devLine, tail string, root bool) {
	indent := strings.Repeat("  ", depth)

	if n.kind == 's' {
		p.rows = append(p.rows, devRow{line: devJSONLine(indent, lead, n.scalar, n.ink, tail)})

		return
	}

	open, close := "{", "}"
	if n.kind == 'a' {
		open, close = "[", "]"
	}

	if len(n.kids) == 0 {
		p.rows = append(p.rows, devRow{line: devJSONLine(indent, lead, open+close, devPunctInk, tail)})

		return
	}

	if !root && p.collapsed[n.path] {
		p.rows = append(p.rows, devRow{
			line: devJSONLine(indent, lead, open+"..."+close, devPunctInk, tail),
			act:  devActJSON,
			key:  n.path,
		})

		return
	}

	row := devRow{line: devJSONLine(indent, lead, open, devPunctInk, "")}
	if !root {
		row.act, row.key = devActJSON, n.path
	}

	p.rows = append(p.rows, row)

	for i, kid := range n.kids {
		kidLead, kidTail := devLine{}, ","

		if i == len(n.kids)-1 {
			kidTail = ""
		}

		if n.kind == 'o' {
			kidLead = devJSONLead(kid.key)
		}

		p.emit(kid, depth+1, kidLead, kidTail, false)
	}

	p.rows = append(p.rows, devRow{line: devJSONPlain(indent, close, tail)})
}

// devJSONLeaf renders one scalar value in its JSON form.
func devJSONLeaf(v reflect.Value, key, path string) *devJSONNode {
	switch v.Kind() {
	case reflect.Bool:
		return devJSONScalar(strconv.FormatBool(v.Bool()), devBoolInk, key, path)
	case reflect.String:
		return devJSONScalar(string(devJSONMarshal(v.String())), devStrInk, key, path)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return devJSONScalar(strconv.FormatInt(v.Int(), 10), devNumInk, key, path)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return devJSONScalar(strconv.FormatUint(v.Uint(), 10), devNumInk, key, path)
	case reflect.Float32, reflect.Float64:
		return devJSONScalar(string(devJSONMarshal(v.Interface())), devNumInk, key, path)
	}

	return devJSONScalar("null", devBoolInk, key, path)
}

// devJSONExtra is the indent + text spans devJSONLine always adds.
const devJSONExtra = 2

// devJSONLine builds indent + lead + text + tail.
func devJSONLine(indent string, lead devLine, text string, ink color.RGBA, tail string) devLine {
	spans := make([]devSpan, 0, len(lead.spans)+devJSONExtra)
	spans = append(spans, devSpan{text: indent, ink: devPunctInk})
	spans = append(spans, lead.spans...)
	spans = append(spans, devSpan{text: text, ink: ink})

	if tail != "" {
		spans = append(spans, devSpan{text: tail, ink: devPunctInk})
	}

	return devLine{spans: spans}
}

// devJSONPlain builds indent + text + tail in punctuation ink.
func devJSONPlain(indent, text, tail string) devLine {
	return devJSONLine(indent, devLine{}, text, devPunctInk, tail)
}

// devJSONLead is the `"key": ` member prefix, key in the key ink.
func devJSONLead(key string) devLine {
	q := string(devJSONMarshal(key))

	return devLine{spans: []devSpan{
		{text: `"`, ink: devPunctInk},
		{text: q[1 : len(q)-1], ink: devKeyInk},
		{text: `": `, ink: devPunctInk},
	}}
}

// devJSONPair is one map entry waiting for its sorted place.
type devJSONPair struct {
	label string
	value reflect.Value
}

// buildMap renders entries sorted by their member name.
func (p *devJSONPrinter) buildMap(v reflect.Value, key, path string) *devJSONNode {
	node := &devJSONNode{kind: 'o', key: key, path: path}
	pairs := make([]devJSONPair, 0, v.Len())

	for _, k := range v.MapKeys() {
		pairs = append(pairs, devJSONPair{label: devJSONMapKey(k), value: v.MapIndex(k)})
	}

	sort.Slice(pairs, func(i, j int) bool { return pairs[i].label < pairs[j].label })

	for _, pair := range pairs {
		node.kids = append(node.kids, p.build(pair.value, pair.label, devJSONJoin(path, pair.label)))
	}

	return node
}

// devJSONMapKey is the member name for one map key.
func devJSONMapKey(k reflect.Value) string {
	if k.Kind() == reflect.String {
		return k.String()
	}

	return fmt.Sprint(k.Interface())
}

// devJSONNode is one value in the rendered document: an object, an array,
// or a scalar, plus the key and dotted path that name it.
type devJSONNode struct {
	kind   byte
	key    string
	path   string
	scalar string
	ink    color.RGBA
	kids   []*devJSONNode
}

// devJSONScalar builds a leaf node.
func devJSONScalar(text string, ink color.RGBA, key, path string) *devJSONNode {
	return &devJSONNode{kind: 's', key: key, path: path, scalar: text, ink: ink}
}

// devJSONNumber reads a json.Number through its string kind.
func devJSONNumber(v reflect.Value) (string, bool) {
	if v.Kind() != reflect.String {
		return "", false
	}

	n, ok := v.Interface().(json.Number)

	return n.String(), ok
}

// build turns one reflected value into a node. key is the member name and
// path the dotted location in the document.
func (p *devJSONPrinter) build(v reflect.Value, key, path string) *devJSONNode {
	if !v.IsValid() {
		return devJSONScalar("null", devBoolInk, key, path)
	}

	if n, ok := devJSONNumber(v); ok {
		return devJSONScalar(n, devNumInk, key, path)
	}

	switch v.Kind() {
	case reflect.Interface, reflect.Pointer:
		if v.IsNil() {
			return devJSONScalar("null", devBoolInk, key, path)
		}

		return p.build(v.Elem(), key, path)
	case reflect.Struct:
		return p.buildStruct(v, key, path)
	case reflect.Slice, reflect.Array:
		if v.Kind() == reflect.Slice && v.Type().Elem().Kind() == reflect.Uint8 {
			return devJSONScalar(string(devJSONMarshal(v.Interface())), devStrInk, key, path)
		}

		return p.buildArray(v, key, path)
	case reflect.Map:
		return p.buildMap(v, key, path)
	}

	return devJSONLeaf(v, key, path)
}

// devJSONMarshal is json.Marshal with a null fallback for values it refuses.
func devJSONMarshal(v any) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		return []byte("null")
	}

	return b
}

// devJSONTag splits a json struct tag into its name and options.
func devJSONTag(tag string) (string, string) {
	if i := strings.IndexByte(tag, ','); i >= 0 {
		return tag[:i], tag[i+1:]
	}

	return tag, ""
}

// devJSONEmpty reports the values omitempty drops.
func devJSONEmpty(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Pointer:
		return v.IsNil()
	}

	return false
}

// devJSONJoin names a child node under its parent.
func devJSONJoin(path, key string) string {
	if path == "" {
		return key
	}

	return path + "." + key
}
