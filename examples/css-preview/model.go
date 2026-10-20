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

func targets() []target {
	targetIndex.once.Do(loadTargets)

	return targetIndex.all
}

func targetByButton(id string) (target, bool) {
	targetIndex.once.Do(loadTargets)
	t, ok := targetIndex.byButton[id]

	return t, ok
}

func targetByKey(key string) (target, bool) {
	targetIndex.once.Do(loadTargets)
	t, ok := targetIndex.byKey[key]

	return t, ok
}
