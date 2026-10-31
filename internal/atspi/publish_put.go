package atspi

import (
	"reflect"

	"github.com/chinmay-sawant/ownframe/internal/host"
)

// maxPublishEvents bounds one publish's event burst. Focus and state
// events are collected first, so a large tree replacement cannot flood
// the bus with per-node notifications.
const maxPublishEvents = 128

// Publish maps tree and, when the bus is live, claims the ownframe
// name and replaces the remote tree. A dead bus still stores the map
// so tests can inspect it, then returns ErrLiveUnavailable.
func (p *Publisher) Publish(tree host.AccessibilityTree) error {
	// A redraw that did not change the snapshot would map the whole tree
	// only to find it identical. Revision is zero for handwritten test
	// trees, which keep the DeepEqual fallback below.
	if p.published && p.Live() && tree.Revision != 0 && tree.Document == p.doc && tree.Revision == p.rev {
		return nil
	}
	app := Map(tree)
	if app.Name == "" {
		app.Name = ApplicationName
	}
	normalizeRootName(&app)
	same := false
	if tree.Revision == 0 {
		same = reflect.DeepEqual(app, p.app)
	}
	if same && p.published && p.Live() {
		p.doc = tree.Document
		p.rev = tree.Revision
		return nil
	}

	old := p.app
	first := !p.published || old.RootPath != app.RootPath
	p.app = app
	p.doc = tree.Document
	p.rev = tree.Revision
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

	// Focus first, then the rest of the state changes, then the child
	// notifications that may be cut by maxPublishEvents.
	events := make([]Event, 0, maxPublishEvents)
	oldFocus := focusPath(old)
	newFocus := focusPath(next)
	if oldFocus != newFocus {
		if oldFocus != "" {
			events = append(events, Event{Path: oldFocus, Kind: "state-changed", Detail: "focused", Enabled: false})
		}
		if newFocus != "" {
			events = append(events, Event{Path: newFocus, Kind: "state-changed", Detail: "focused", Enabled: true})
		}
	}
	for path, oldObj := range oldByPath {
		newObj, ok := newByPath[path]
		if !ok {
			continue
		}
		events = appendState(events, path, "active", oldObj.State.Active, newObj.State.Active)
		events = appendState(events, path, "checked", oldObj.State.Checked, newObj.State.Checked)
		events = appendState(events, path, "enabled", oldObj.State.Enabled, newObj.State.Enabled)
		events = appendState(events, path, "sensitive", oldObj.State.Sensitive, newObj.State.Sensitive)
		events = appendState(events, path, "editable", oldObj.State.Editable, newObj.State.Editable)
	}

	if replacedTree(old, next) {
		if oldRoot := rootChildPath(old); oldRoot != "" {
			events = append(events, Event{Path: old.RootPath, Kind: "children-changed", Detail: "remove", Child: oldRoot, Index: 0})
		}
		if newRoot := rootChildPath(next); newRoot != "" {
			events = append(events, Event{Path: next.RootPath, Kind: "children-changed", Detail: "add", Child: newRoot, Index: 0})
		}
	} else {
		events = append(events, childChanges(old, next, oldByPath, newByPath)...)
	}
	if len(events) > maxPublishEvents {
		events = events[:maxPublishEvents]
	}
	for _, ev := range events {
		p.emit(ev)
	}
}

// replacedTree reports that the page document is new, so one root child
// notification describes the whole changed subtree.
func replacedTree(old, next Application) bool {
	return old.Document != next.Document || rootChildPath(old) != rootChildPath(next)
}

// rootChildPath is the page document object under the synthetic
// application root.
func rootChildPath(app Application) string {
	for _, obj := range app.Objects {
		if obj.Path != app.RootPath {
			continue
		}
		if len(obj.Children) > 0 {
			return obj.Children[0]
		}
		return ""
	}
	return ""
}

// childChanges reports at most one children-changed event per parent.
// A parent that only lost children reports remove; a parent that gained
// any reports the first added child. A vanished ancestor hides its own
// subtree, which the parent event already identifies.
func childChanges(old, next Application, oldByPath, newByPath map[string]Object) []Event {
	type change struct {
		child string
		index int32
		add   bool
	}
	byParent := make(map[string]change)
	for path, obj := range oldByPath {
		if _, ok := newByPath[path]; ok {
			continue
		}
		parent := obj.Parent
		if parent == "" {
			parent = old.RootPath
		}
		if _, ok := newByPath[parent]; !ok {
			continue
		}
		if _, ok := byParent[parent]; !ok {
			byParent[parent] = change{child: path, index: indexIn(old, parent, path)}
		}
	}
	for path, obj := range newByPath {
		if _, ok := oldByPath[path]; ok {
			continue
		}
		parent := obj.Parent
		if parent == "" {
			parent = next.RootPath
		}
		prev, ok := byParent[parent]
		if !ok || !prev.add {
			byParent[parent] = change{child: path, index: indexIn(next, parent, path), add: true}
		}
	}
	out := make([]Event, 0, len(byParent))
	for parent, c := range byParent {
		detail := "remove"
		if c.add {
			detail = "add"
		}
		out = append(out, Event{Path: parent, Kind: "children-changed", Detail: detail, Child: c.child, Index: c.index})
	}
	return out
}

func appendState(events []Event, path, state string, was, now bool) []Event {
	if was == now {
		return events
	}
	return append(events, Event{Path: path, Kind: "state-changed", Detail: state, Enabled: now})
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
