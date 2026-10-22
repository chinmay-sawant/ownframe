package atspi

import "github.com/chinmay-sawant/ownframe/internal/host"

// Publish maps tree and, when the bus is live, claims the ownframe
// name and replaces the remote tree. A dead bus still stores the map
// so tests can inspect it, then returns ErrLiveUnavailable.
func (p *Publisher) Publish(tree host.AccessibilityTree) error {
	same := tree.Document == p.doc && tree.Revision == p.rev && tree.Focused == p.focus
	if same && len(p.app.Objects) > 0 {
		if !p.Live() {
			return ErrLiveUnavailable
		}
		return nil
	}
	app := Map(tree)
	p.app = app
	p.doc = tree.Document
	p.rev = tree.Revision
	p.focus = tree.Focused
	if !p.Live() {
		return ErrLiveUnavailable
	}
	if err := p.bus.Claim(app.Name); err != nil {
		return err
	}
	if err := p.bus.Emit(Event{Path: app.RootPath, Kind: "children-changed"}); err != nil {
		return err
	}
	return p.bus.Publish(app)
}
