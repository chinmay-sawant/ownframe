package a11y

import (
	"strings"

	"github.com/chinmay-sawant/blinkless/markup"
)

// State is the accessibility document retained between redraws.
// A zero document means the tree is not available yet.
type State struct {
	source   string
	root     *markup.Node
	document uint64
	retire   bool
}

// Commit records source as the current document. The first commit, and
// any commit after Invalidate, advances the document number. A different
// source drops the parsed root.
func (s *State) Commit(source string) {
	if s.document == 0 || s.retire {
		s.document++
		s.retire = false
	}
	if s.source != source {
		s.root = nil
	}
	s.source = source
}

// Invalidate marks the current document stale. A nil state is a no-op.
func (s *State) Invalidate() {
	if s == nil {
		return
	}
	s.retire = true
}

// Source is the last committed document source.
func (s *State) Source() string { return s.source }

// Attached reports whether a parsed document root is held. A patch mutates
// that root; without one the edits have nowhere to land, so the page skips
// the whole refresh instead of rebuilding the edit slice for a no-op.
func (s *State) Attached() bool { return s != nil && s.root != nil }

// Edit is one change applied to the cached markup root.
// Leaving all action fields empty/nil leaves the node untouched.
// An empty ID uses the id argument.
type Edit struct {
	ID              string
	Attribute       string
	Value           string
	Text            *string
	RemoveAttribute string
	AddClass        string
	RemoveClass     string
	ToggleClass     string
	Remove          bool
	InnerHTML       *string
	AppendHTML      string
}

// Apply writes edits into the cached root. A nil state or a root that
// has not been parsed yet is left alone, so the next tree parses source.
func (s *State) Apply(id string, edits []Edit) {
	if s == nil || s.root == nil {
		return
	}
	names := NewNames()
	names.Index(s.root, nil)
	for _, e := range edits {
		if e.ID == "" {
			e.ID = id
		}
		if e.Remove {
			detachMarkupNodes(s.root, names.IDs[e.ID])
			names = NewNames()
			names.Index(s.root, nil)
			continue
		}
		needsReindex := false
		for _, n := range names.IDs[e.ID] {
			if e.InnerHTML != nil {
				n.Children = parseMarkupFragment(*e.InnerHTML)
				needsReindex = true
			}
			if e.AppendHTML != "" {
				n.Children = append(n.Children, parseMarkupFragment(e.AppendHTML)...)
				needsReindex = true
			}
			if e.Text != nil {
				n.Children = []*markup.Node{{Type: markup.TypeText, Text: *e.Text}}
				needsReindex = true
			}
			if e.Attribute != "" {
				if n.Attrs == nil {
					n.Attrs = map[string]string{}
				}
				n.Attrs[strings.ToLower(e.Attribute)] = e.Value
			}
			if e.RemoveAttribute != "" && n.Attrs != nil {
				delete(n.Attrs, strings.ToLower(e.RemoveAttribute))
			}
			if e.AddClass != "" {
				applyMarkupClass(n, e.AddClass, classAdd)
			}
			if e.RemoveClass != "" {
				applyMarkupClass(n, e.RemoveClass, classRemove)
			}
			if e.ToggleClass != "" {
				applyMarkupClass(n, e.ToggleClass, classToggle)
			}
		}
		if needsReindex {
			names = NewNames()
			names.Index(s.root, nil)
		}
	}
}

type classOp uint8

const (
	classAdd classOp = iota
	classRemove
	classToggle
)

func applyMarkupClass(n *markup.Node, token string, op classOp) {
	tok := strings.TrimSpace(token)
	if tok == "" || strings.ContainsAny(tok, " \t\n\r\f") {
		return
	}
	classes := strings.Fields(n.Attrs["class"])
	idx := -1
	for i, c := range classes {
		if c == tok {
			idx = i
			break
		}
	}
	switch op {
	case classAdd:
		if idx >= 0 {
			return
		}
		classes = append(classes, tok)
	case classRemove:
		if idx < 0 {
			return
		}
		classes = append(classes[:idx], classes[idx+1:]...)
	case classToggle:
		if idx >= 0 {
			classes = append(classes[:idx], classes[idx+1:]...)
		} else {
			classes = append(classes, tok)
		}
	}
	if n.Attrs == nil {
		n.Attrs = map[string]string{}
	}
	if len(classes) == 0 {
		delete(n.Attrs, "class")
		return
	}
	n.Attrs["class"] = strings.Join(classes, " ")
}

func detachMarkupNodes(root *markup.Node, targets []*markup.Node) {
	if root == nil || len(targets) == 0 {
		return
	}
	drop := make(map[*markup.Node]bool, len(targets))
	for _, t := range targets {
		drop[t] = true
	}
	var walk func(n *markup.Node)
	walk = func(n *markup.Node) {
		if len(n.Children) == 0 {
			return
		}
		kept := n.Children[:0]
		for _, c := range n.Children {
			if drop[c] {
				continue
			}
			walk(c)
			kept = append(kept, c)
		}
		n.Children = kept
	}
	walk(root)
}

func parseMarkupFragment(fragment string) []*markup.Node {
	if fragment == "" {
		return nil
	}
	parsed, err := markup.Parse([]byte("<html><body>" + fragment + "</body></html>"))
	if err != nil || parsed == nil {
		return nil
	}
	var body *markup.Node
	var findBody func(n *markup.Node)
	findBody = func(n *markup.Node) {
		if body != nil || n == nil {
			return
		}
		if n.Type == markup.TypeElement && strings.EqualFold(n.Name, "body") {
			body = n
			return
		}
		for _, c := range n.Children {
			findBody(c)
		}
	}
	findBody(parsed)
	if body != nil {
		return body.Children
	}
	return parsed.Children
}

func (s *State) ensureRoot() (*markup.Node, error) {
	if s.root != nil {
		return s.root, nil
	}
	root, err := markup.Parse([]byte(s.source))
	if err != nil {
		return nil, err
	}
	s.root = root
	return root, nil
}
