package atspi

import (
	"reflect"

	"github.com/chinmay-sawant/ownframe/internal/host"
)

// Publish maps tree and, when the bus is live, claims the ownframe
// name and replaces the remote tree. A dead bus still stores the map
// so tests can inspect it, then returns ErrLiveUnavailable.
func (p *Publisher) Publish(tree host.AccessibilityTree) error {
	app := Map(tree)
	same := reflect.DeepEqual(app, p.app)
	if same && p.published {
		if !p.Live() {
			return ErrLiveUnavailable
		}
		return nil
	}
	p.app = app
	p.published = false
	p.doc = tree.Document
	if !p.Live() {
		return ErrLiveUnavailable
	}
	if err := p.bus.Claim(app.Name); err != nil {
		return err
	}
	if err := p.bus.Emit(Event{Path: app.RootPath, Kind: "children-changed"}); err != nil {
		return err
	}
	if err := p.bus.Publish(app); err != nil {
		return err
	}
	p.published = true
	return nil
}
