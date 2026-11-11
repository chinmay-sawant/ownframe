package atspi

import (
	"context"
	"slices"

	"github.com/chinmay-sawant/ownframe/internal/host"
)

func (v *serverView) performAction(ctx context.Context, obj Object, name string) callResult {
	if v.actions == nil || !obj.State.Enabled || !obj.State.Sensitive || !slices.Contains(obj.Actions, name) {
		return okResult("b", encodeBoolValue(false))
	}
	kind, ok := ActionKind(name)
	if !ok || (kind != host.AccessibilityFocus && kind != host.AccessibilityActivate) {
		return okResult("b", encodeBoolValue(false))
	}
	err := v.actions(ctx, host.AccessibilityRequest{Document: v.app.Document, Node: obj.ID, Kind: kind})
	return okResult("b", encodeBoolValue(err == nil))
}
