package csstrigger

import (
	"fmt"
	"time"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/animation"
)

// Dispatch applies matching trigger actions to each box's animation player.
func Dispatch(boxes []layout.Box, event Event, players map[int]*animation.Player, now time.Time) error {
	for _, action := range Resolve(boxes, event) {
		player := players[action.BoxIndex]
		if player == nil {
			return fmt.Errorf("csstrigger: no player for box %d", action.BoxIndex)
		}
		if err := player.TriggerEffect(action.AnimationIndex, action.Value, now); err != nil {
			return fmt.Errorf("csstrigger: box %d animation %d: %w", action.BoxIndex, action.AnimationIndex, err)
		}
	}
	return nil
}
