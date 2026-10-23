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

// Edit is one change applied to the cached markup root.
// A nil Text leaves the children. An empty Attribute leaves attributes.
// An empty ID uses the id argument.
type Edit struct {
	ID        string
	Attribute string
	Value     string
	Text      *string
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
		for _, n := range names.IDs[e.ID] {
			if e.Text != nil {
				n.Children = []*markup.Node{{Type: markup.TypeText, Text: *e.Text}}
			}
			if e.Attribute != "" {
				if n.Attrs == nil {
					n.Attrs = map[string]string{}
				}
				n.Attrs[strings.ToLower(e.Attribute)] = e.Value
			}
		}
	}
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
