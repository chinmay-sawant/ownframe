package motion

import (
	"fmt"
	"time"

	"github.com/chinmay-sawant/blinkless/layout"
)

// DispatchTrigger applies matching trigger actions to each box's animation player.
func DispatchTrigger(boxes []layout.Box, event TriggerEvent, players map[int]*Player, now time.Time) error {
	for _, action := range resolveTriggers(boxes, event) {
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
