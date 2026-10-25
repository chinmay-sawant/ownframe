package main

import "sync"

// targetIndex caches the immutable target list and its key lookup. A click
// asks for a target by sample key, and rebuilding all 43 entries there costs
// more than the lookup.
var targetIndex struct {
	once  sync.Once
	all   []target
	byKey map[string]target
}

func loadTargets() {
	all := catalogTargets()
	byKey := make(map[string]target, len(all))
	for _, item := range all {
		byKey[item.key] = item
	}
	targetIndex.all, targetIndex.byKey = all, byKey
}
