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
	if app.Name == "" {
		app.Name = ApplicationName
	}
	normalizeRootName(&app)
	same := reflect.DeepEqual(app, p.app)
	if same && p.published && p.Live() {
		return nil
	}

	old := p.app
	first := !p.published || old.RootPath != app.RootPath
	p.app = app
	p.doc = tree.Document
	p.published = false

	if err := p.bus.Claim(DefaultName); err != nil {
		return err
	}
	if err := p.bus.Publish(app); err != nil {
		return err
	}
	p.published = true
	p.emitChanges(old, app, first)
	return nil
}

// normalizeRootName gives the application root a readable name when
// the page document itself has none. Accerciser and Orca list the
// application by this name.
func normalizeRootName(app *Application) {
	for i := range app.Objects {
		if app.Objects[i].Path == app.RootPath && app.Objects[i].Name == "" {
			app.Objects[i].Name = app.Name
		}
	}
}

func (p *Publisher) emitChanges(old, next Application, first bool) {
	if first {
		p.emitFocus(next)
		return
	}
	oldByPath := objectMap(old)
	newByPath := objectMap(next)

	for path, obj := range oldByPath {
		if _, ok := newByPath[path]; ok {
			continue
		}
		parent := obj.Parent
		if parent == "" {
			parent = old.RootPath
		}
		if _, ok := newByPath[parent]; !ok {
			// An ancestor removal already identifies the vanished subtree.
			continue
		}
		p.emit(Event{Path: parent, Kind: "children-changed", Detail: "remove", Child: path, Index: indexIn(old, parent, path)})
	}
	for path, obj := range newByPath {
		if _, ok := oldByPath[path]; ok {
			continue
		}
		parent := obj.Parent
		if parent == "" {
			parent = next.RootPath
		}
		p.emit(Event{Path: parent, Kind: "children-changed", Detail: "add", Child: path, Index: indexIn(next, parent, path)})
	}

	oldFocus := focusPath(old)
	newFocus := focusPath(next)
	if oldFocus != newFocus {
		if oldFocus != "" {
			p.emit(Event{Path: oldFocus, Kind: "state-changed", Detail: "focused", Enabled: false})
		}
		if newFocus != "" {
			p.emit(Event{Path: newFocus, Kind: "state-changed", Detail: "focused", Enabled: true})
		}
	}
	for path, oldObj := range oldByPath {
		newObj, ok := newByPath[path]
		if !ok {
			continue
		}
		p.emitState(path, "active", oldObj.State.Active, newObj.State.Active)
		p.emitState(path, "checked", oldObj.State.Checked, newObj.State.Checked)
		p.emitState(path, "enabled", oldObj.State.Enabled, newObj.State.Enabled)
		p.emitState(path, "sensitive", oldObj.State.Sensitive, newObj.State.Sensitive)
		p.emitState(path, "editable", oldObj.State.Editable, newObj.State.Editable)
	}
}

func (p *Publisher) emitState(path, state string, was, now bool) {
	if was == now {
		return
	}
	p.emit(Event{Path: path, Kind: "state-changed", Detail: state, Enabled: now})
}

func (p *Publisher) emitFocus(app Application) {
	if path := focusPath(app); path != "" {
		p.emit(Event{Path: path, Kind: "state-changed", Detail: "focused", Enabled: true})
	}
}

func (p *Publisher) emit(ev Event) {
	if p.bus == nil {
		return
	}
	_ = p.bus.Emit(ev)
}

func objectMap(app Application) map[string]Object {
	out := make(map[string]Object, len(app.Objects))
	for _, obj := range app.Objects {
		out[obj.Path] = obj
	}
	return out
}

func focusPath(app Application) string {
	if app.Focused != "" {
		for _, obj := range app.Objects {
			if obj.ID == app.Focused {
				return obj.Path
			}
		}
	}
	for _, obj := range app.Objects {
		if obj.State.Focused {
			return obj.Path
		}
	}
	return ""
}

func indexIn(app Application, parent, child string) int32 {
	for _, obj := range app.Objects {
		if obj.Path != parent {
			continue
		}
		for i, path := range obj.Children {
			if path == child {
				return int32(i)
			}
		}
	}
	return -1
}
