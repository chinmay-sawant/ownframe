package atspi

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/ownframe/internal/host"
)

func actionResult(t *testing.T, result callResult) bool {
	t.Helper()
	if result.errName != "" || result.sig != "b" {
		t.Fatalf("action reply failed: %s", result.errName)
	}
	d := dbusDecoder{body: result.body}
	value, err := d.uint32Val()
	if err != nil {
		t.Fatal(err)
	}
	return value != 0
}

func TestServeActionsDeliverFocusAndActivation(t *testing.T) {
	view, _ := testServerView(t)
	requests := []host.AccessibilityRequest{}
	view.actions = func(_ context.Context, request host.AccessibilityRequest) error {
		requests = append(requests, request)
		return nil
	}
	name := objectPathFor(t, view, "id:name")
	if !actionResult(t, callView(view, name, atspiComponent, "GrabFocus", nil)) {
		t.Fatal("focus was refused")
	}
	button := objectPathFor(t, view, "id:go")
	var index dbusEncoder
	index.int32Val(0)
	if !actionResult(t, callView(view, button, atspiAction, "DoAction", index.bytes())) {
		t.Fatal("button action was refused")
	}
	if len(requests) != 2 || requests[0].Node != "id:name" || requests[0].Kind != host.AccessibilityFocus ||
		requests[1].Node != "id:go" || requests[1].Document != view.app.Document {
		t.Fatal("action request did not preserve its document and node identity")
	}
	view.actions = func(context.Context, host.AccessibilityRequest) error { return host.ErrAccessibilityStale }
	if actionResult(t, callView(view, name, atspiComponent, "GrabFocus", nil)) {
		t.Fatal("stale document was reported successful")
	}
}

func TestServeActionsRejectInvalidDisabledAndUnsupported(t *testing.T) {
	view, _ := testServerView(t)
	view.actions = func(context.Context, host.AccessibilityRequest) error {
		t.Fatal("refused action reached the page")
		return nil
	}
	name := objectPathFor(t, view, "id:name")
	var invalid dbusEncoder
	invalid.int32Val(-1)
	if result := callView(view, name, atspiAction, "DoAction", invalid.bytes()); result.errName != "org.a11y.atspi.Error.InvalidIndex" {
		t.Fatal("invalid action index was accepted")
	}
	var setText dbusEncoder
	setText.int32Val(1)
	if actionResult(t, callView(view, name, atspiAction, "DoAction", setText.bytes())) {
		t.Fatal("text editing was accepted by a focus/activation-only bridge")
	}
	obj := view.objects[name]
	obj.State.Enabled = false
	view.objects[name] = obj
	if actionResult(t, callView(view, name, atspiComponent, "GrabFocus", nil)) {
		t.Fatal("disabled focus was accepted")
	}
	delete(view.objects, name)
	if result := callView(view, name, atspiComponent, "GrabFocus", nil); result.errName != "org.freedesktop.DBus.Error.UnknownObject" {
		t.Fatal("removed object was accepted")
	}
}
