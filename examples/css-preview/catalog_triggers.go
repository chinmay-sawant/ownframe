package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/internal/animation"
	"github.com/chinmay-sawant/ownframe/internal/csstrigger"
)

func dispatchAnimationTriggers(boxes []ownframe.Box, target int, sourceID string, player *animation.Player, now time.Time) error {
	if target < 0 || target >= len(boxes) || player == nil {
		return nil
	}
	names := triggerNames(boxes[target])
	if len(names) == 0 {
		return nil
	}
	source, ok := findBoxIndex(boxes, sourceID)
	if !ok {
		return fmt.Errorf("csspreview: trigger source %q is missing", sourceID)
	}
	players := map[int]*animation.Player{target: player}
	for _, name := range names {
		if err := csstrigger.Dispatch(boxes, csstrigger.Event{Name: name, Source: source, Active: true}, players, now); err != nil {
			return err
		}
	}
	return nil
}

func triggerNames(box ownframe.Box) []string {
	seen := map[string]bool{}
	var names []string
	for _, item := range box.Animations {
		for _, token := range strings.Fields(strings.ReplaceAll(item.Trigger, ",", " ")) {
			if strings.HasPrefix(token, "--") && !seen[token] {
				seen[token] = true
				names = append(names, token)
			}
		}
	}
	return names
}

func findBoxIndex(boxes []ownframe.Box, id string) (int, bool) {
	for i, box := range boxes {
		if box.ID == id {
			return i, true
		}
	}
	return -1, false
}
