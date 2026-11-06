package ui

import (
	"context"
	"slices"
	"sort"
	"strings"

	"github.com/chinmay-sawant/ownframe"
)

// maxPatches bounds one incremental update. PatchElements re-lays out every
// patched element, so a long batch costs more than one full rebuild. A tree
// with more changed elements than this takes the full-render path instead.
const maxPatches = 64

// treeDiff is the outcome of comparing the retained tree with a rebuilt one.
//
// ok reports that the two trees have the same structure and that every changed
// node can be addressed by a stable id. In that case patches holds one entry
// per changed element and actions holds the callback registry the fresh tree
// would publish. ok is false when Update must render the whole document
// again; the patches and actions are then meaningless.
type treeDiff struct {
	patches []ownframe.ElementPatch
	actions map[string]func(context.Context) error
	// nextID mirrors the renderer's action counter: one ownframe-ui-N key
	// per element with an OnClick handler, in document order.
	nextID int
	ok     bool
}

// diffTree compares last, the tree the page was rendered from, with root, the
// tree a caller just built. Same tag, same child count, and same child order
// at every node means the structure matches; text, attributes, and styles are
// compared as values. ok is false when the caller must take the full-render
// path: a structural change, a changed node without an id, an attribute that
// disappeared, more than maxPatches changes, or any tree the renderer would
// reject.
func diffTree(last, root *Element) treeDiff {
	d := treeDiff{ok: last != nil && root != nil}
	if !d.ok {
		return d
	}
	d.walk(nil, last, root, 0)
	if !d.ok {
		d.patches = nil
		d.actions = nil
	}
	return d
}

// walk compares one node and its subtree. path carries the ancestors of both
// trees so a cycle falls back to the renderer, which reports it as an error.
func (d *treeDiff) walk(path []*Element, last, root *Element, depth int) {
	if !d.ok {
		return
	}
	if last == nil || root == nil || depth >= maxDepth {
		d.ok = false
		return
	}
	if slices.Contains(path, last) || slices.Contains(path, root) {
		d.ok = false
		return
	}
	// The renderer owns validation. Any error it would report falls back to
	// the full render, which returns the canonical error unchanged.
	if root.err != nil {
		d.ok = false
		return
	}
	if last.tag != root.tag || len(last.children) != len(root.children) {
		d.ok = false
		return
	}
	lastID, rootID := last.attrs["id"], root.attrs["id"]
	// A rename cannot be patched safely: the batch locates the element by
	// its old id and the relayout resolves it again after the edit.
	if lastID != rootID {
		d.ok = false
		return
	}
	// Action ids count OnClick nodes in document order. The rendered DOM
	// keeps its data-action values across a patch, so the new tree must
	// carry a handler exactly where the retained tree did.
	if (last.onClick == nil) != (root.onClick == nil) {
		d.ok = false
		return
	}
	if root.onClick != nil {
		if _, custom := root.attrs["data-action"]; custom {
			// renderDocument reports this combination as an error.
			d.ok = false
			return
		}
		d.nextID++
		if d.actions == nil {
			d.actions = make(map[string]func(context.Context) error)
		}
		d.actions[actionName(d.nextID)] = root.onClick
	}
	// A void element cannot hold text or children; renderDocument refuses it.
	if isVoidTag(root.tag) && (root.text != "" || len(root.children) != 0) {
		d.ok = false
		return
	}

	attrEdits, attrsOK := changedAttributes(last.attrs, root.attrs)
	if !attrsOK {
		d.ok = false
		return
	}
	textChanged := last.text != root.text
	stylesChanged := !sameMap(last.styles, root.styles)
	if stylesChanged && len(root.styles) == 0 {
		// Removing the last declaration would leave style="" in the tree
		// while the rebuilt source omits the attribute, so rebuild.
		d.ok = false
		return
	}
	changed := textChanged || len(attrEdits) != 0 || stylesChanged

	if changed {
		if lastID == "" {
			d.ok = false
			return
		}
		// SetText replaces the element's children, so only a leaf can take
		// a text edit; a node that also owns child elements keeps its text
		// before them and needs the full render.
		if textChanged && len(root.children) != 0 {
			d.ok = false
			return
		}
		if len(d.patches) >= maxPatches {
			d.ok = false
			return
		}
		edits := attrEdits
		if stylesChanged {
			edits = append(edits, ownframe.ElementEdit{
				Attribute: "style",
				Value:     inlineStyle(root),
			})
		}
		if textChanged {
			text := root.text
			edits = append(edits, ownframe.ElementEdit{Text: &text})
		}
		d.patches = append(d.patches, ownframe.ElementPatch{ID: lastID, Edits: edits})
	}

	path = append(path, last, root)
	for i := range root.children {
		d.walk(path, last.children[i], root.children[i], depth+1)
		if !d.ok {
			return
		}
	}
}

// changedAttributes returns one edit for every attribute that was added or
// whose value changed. ok is false when an attribute disappeared: the patch
// API sets values but cannot remove an attribute, so the caller rebuilds.
func changedAttributes(last, root map[string]string) ([]ownframe.ElementEdit, bool) {
	if sameMap(last, root) {
		return nil, true
	}
	edits := make([]ownframe.ElementEdit, 0, len(root))
	for name, value := range root {
		if oldValue, exists := last[name]; exists && oldValue == value {
			continue
		}
		edits = append(edits, ownframe.ElementEdit{Attribute: name, Value: value})
	}
	for name := range last {
		if _, exists := root[name]; !exists {
			return nil, false
		}
	}
	return edits, true
}

// sameMap reports whether two string maps hold the same keys and values.
func sameMap(last, root map[string]string) bool {
	if len(last) != len(root) {
		return false
	}
	for name, value := range root {
		if last[name] != value {
			return false
		}
	}
	return true
}

// inlineStyle serializes an element's declarations the way visit writes the
// style attribute: names sorted, each as name:value;. The builder escapes the
// result when it writes HTML; this value lands in the parsed tree, which
// already holds decoded text, so it is not escaped a second time.
func inlineStyle(e *Element) string {
	if len(e.styles) == 0 {
		return ""
	}
	names := make([]string, 0, len(e.styles))
	for name := range e.styles {
		names = append(names, name)
	}
	sort.Strings(names)
	var style strings.Builder
	for _, name := range names {
		style.WriteString(name)
		style.WriteByte(':')
		style.WriteString(e.styles[name])
		style.WriteByte(';')
	}
	return style.String()
}
