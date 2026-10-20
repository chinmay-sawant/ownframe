package main

type view struct {
	Started  map[string]bool
	Catalog  []catalogRow
	Rows     []catalogRow
	Readout  string
	Selected string
	TopPad   int
	BotPad   int
}

type target struct{ key, id, button, property, kind string }

func targets() []target { return catalogTargets() }

func targetByButton(id string) (target, bool) {
	for _, t := range targets() {
		if t.button == id {
			return t, true
		}
	}
	return target{}, false
}

func targetByKey(key string) (target, bool) {
	for _, t := range targets() {
		if t.key == key {
			return t, true
		}
	}
	return target{}, false
}
