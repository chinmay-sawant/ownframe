package main

// view is the template data. The page renders it on every click, so the card
// classes and the readout are ordinary template fields instead of patches
// into a parsed tree.
type view struct {
	// Started marks a card whose Start button ran at least once.
	Started map[string]bool
	// Restart is the keyframe-name variant of a card. It flips on every
	// Start after the first, so the target binds a new animation instance.
	Restart map[string]bool
	// Restarting hides a card target for the publish that starts a restart,
	// so the runtime drops the old player instead of carrying its clock into
	// the new keyframe name.
	Restarting map[string]bool
	Catalog    []catalogRow
	Rows       []catalogRow
	Readout    string
	Selected   string
	TopPad     int
	BotPad     int
}

// newView returns a view with the click-state maps allocated.
func newView() *view {
	return &view{Started: map[string]bool{}, Restart: map[string]bool{}, Restarting: map[string]bool{}}
}

// target names one card's preview target and Start button.
type target struct{ key, id, button, property, kind string }

func targets() []target {
	targetIndex.once.Do(loadTargets)

	return targetIndex.all
}

func targetByKey(key string) (target, bool) {
	targetIndex.once.Do(loadTargets)
	t, ok := targetIndex.byKey[key]

	return t, ok
}
