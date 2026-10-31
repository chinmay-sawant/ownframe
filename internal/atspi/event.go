package atspi

import "fmt"

// encodeEvent builds the member, signature, and body for an accessibility
// event signal. It has no build tag so the package tests compile on every host.
func encodeEvent(ev Event, busName, root string) (string, []byte, string, error) {
	switch ev.Kind {
	case "children-changed":
		detail := ev.Detail
		if detail == "" {
			detail = "add"
		}
		child := ev.Child
		if child == "" {
			child = ev.Path
		}
		if child == "" {
			child = root
		}
		var e dbusEncoder
		e.stringVal(detail)
		e.int32Val(ev.Index)
		e.int32Val(0)
		e.variant("(so)", func(e *dbusEncoder) {
			e.structValue(func(e *dbusEncoder) {
				e.stringVal(busName)
				e.stringVal(child)
			})
		})
		e.array("{sv}", func(*dbusEncoder) {})
		return "siiva{sv}", e.bytes(), "ChildrenChanged", nil
	case "state-changed":
		detail := ev.Detail
		if detail == "" {
			detail = "focused"
		}
		enabled := int32(0)
		if ev.Enabled {
			enabled = 1
		}
		var e dbusEncoder
		e.stringVal(detail)
		e.int32Val(enabled)
		e.int32Val(0)
		e.variant("i", func(e *dbusEncoder) { e.int32Val(0) })
		e.array("{sv}", func(*dbusEncoder) {})
		return "siiva{sv}", e.bytes(), "StateChanged", nil
	}
	return "", nil, "", fmt.Errorf("atspi: unsupported event %q", ev.Kind)
}
