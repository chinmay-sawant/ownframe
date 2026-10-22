package page

import "fmt"

func (p *Page) patchReady(id string) error {
	if p == nil {
		return ErrNilPage
	}
	if id == "" {
		return fmt.Errorf("empty id: %w", ErrPatchRefused)
	}
	if p.cache == nil || p.draw.Display == nil {
		return fmt.Errorf("no cache: %w", ErrPatchRefused)
	}
	if !p.retainState {
		return fmt.Errorf("retain state off: %w", ErrPatchRefused)
	}

	return nil
}

func (p *Page) applyElementEdits(id string, edits []ElementEdit) bool {
	changed := false
	for _, e := range edits {
		target := e.ID
		if target == "" {
			target = id
		}
		switch {
		case e.Text != nil:
			changed = p.cache.SetText(target, *e.Text) || changed
		case e.Attribute != "":
			changed = p.cache.SetAttribute(target, e.Attribute, e.Value) || changed
		}
	}

	return changed
}

func (p *Page) refusePatch(err error) error {
	p.rollbackPatch()
	if err == nil {
		return ErrPatchRefused
	}

	return fmt.Errorf("%w: %w", ErrPatchRefused, err)
}
