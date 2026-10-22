package page

import "fmt"

// statNote reports a stat or read error once per distinct message.
func (p *Page) statNote(wp *watchPath, err error) error {
	if wp.statErr == err.Error() {
		return nil
	}

	wp.statErr = err.Error()

	return p.sourceErr(wp.path, err)
}

func (p *Page) sourceErr(path string, err error) error {
	return fmt.Errorf("%s: %w", cleanReload(path), err)
}

// resolveStates clears hover and active ids the new layout does not have.
// syncForm does the same for focus when Redraw runs.
func (p *Page) resolveStates() {
	ids := map[string]bool{}
	for _, box := range p.boxes {
		if box.ID != "" {
			ids[box.ID] = true
		}
	}

	if p.hoverID() != "" && !ids[p.hoverID()] {
		p.clearHoverID()
	}

	if p.activeID() != "" && !ids[p.activeID()] {
		p.setActiveID("")
	}
}
