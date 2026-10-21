package main

import "sync"

// targetIndex caches the immutable target list and its lookup maps. The
// catalog asks for a target by key or button on every click and rebind, and
// rebuilding all 43 entries there costs more than the lookup.
var targetIndex struct {
	once     sync.Once
	all      []target
	byKey    map[string]target
	byButton map[string]target
}

func loadTargets() {
	all := catalogTargets()
	byKey := make(map[string]target, len(all))
	byButton := make(map[string]target, len(all))
	for _, item := range all {
		byKey[item.key] = item
		byButton[item.button] = item
	}
	targetIndex.all, targetIndex.byKey, targetIndex.byButton = all, byKey, byButton
}
