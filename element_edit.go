package ownframe

import "github.com/chinmay-sawant/ownframe/internal/page"

// ElementEdit is one change to an element in the retained tree: an
// attribute to set, or the text to replace. A nil Text and an empty
// Attribute change nothing.
type ElementEdit = page.ElementEdit
