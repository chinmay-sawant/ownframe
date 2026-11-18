package main

import (
	"context"

	"github.com/chinmay-sawant/ownframe"
)

// Applier writes style patches into a page and remembers what it wrote, so a
// frame that changes nothing patches nothing, and an element whose driven
// value goes away returns to its original style.
type Applier struct {
	last map[string]string
	base map[string]string
}

// NewApplier keeps base as the original style of every element.
func NewApplier(base map[string]string) *Applier {
	return &Applier{last: map[string]string{}, base: base}
}

// Diff returns the patches that differ from what was written last time, and
// records them.
func (a *Applier) Diff(patches []StylePatch) []StylePatch {
	var out []StylePatch

	seen := map[string]bool{}

	for _, p := range patches {
		seen[p.ID] = true

		if a.last[p.ID] != p.Style {
			a.last[p.ID] = p.Style
			out = append(out, p)
		}
	}

	for id := range a.last {
		if !seen[id] {
			delete(a.last, id)

			out = append(out, StylePatch{ID: id, Style: a.base[id]})
		}
	}

	return out
}

// Apply patches the page with the changed styles in one batch.
func (a *Applier) Apply(ctx context.Context, page *ownframe.Page, patches []StylePatch) error {
	changed := a.Diff(patches)
	if len(changed) == 0 {
		return nil
	}

	batch := make([]ownframe.ElementPatch, len(changed))
	for i, p := range changed {
		batch[i] = ownframe.ElementPatch{
			ID:    p.ID,
			Edits: []ownframe.ElementEdit{{ID: p.ID, Attribute: "style", Value: p.Style}},
		}
	}

	return page.PatchElements(ctx, batch)
}
