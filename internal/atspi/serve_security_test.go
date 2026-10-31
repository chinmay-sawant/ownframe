package atspi

import (
	"encoding/binary"
	"strings"
	"testing"
)

func TestServeApplicationIDWriteRequiresRegistry(t *testing.T) {
	view, appID := testServerView(t)
	setID := func(sender string, value int32) callResult {
		var body dbusEncoder
		body.stringVal(atspiApplication)
		body.stringVal("Id")
		body.variant("i", func(e *dbusEncoder) { e.int32Val(value) })
		return callViewFrom(view, sender, accessibleRootPath, dbusProperties, "Set", body.bytes())
	}

	for _, sender := range []string{"", ":1.99"} {
		res := setID(sender, 11)
		if res.errName != "org.freedesktop.DBus.Error.PropertyReadOnly" {
			t.Fatalf("sender %q: error = %q %q", sender, res.errName, res.errText)
		}
	}
	if *appID != 0 {
		t.Fatalf("foreign write changed id to %d", *appID)
	}

	if res := setID(":1.1", 11); res.errName != "" {
		t.Fatalf("registry write: %s: %s", res.errName, res.errText)
	}
	if *appID != 11 {
		t.Fatalf("registry write left id at %d", *appID)
	}
}

func TestServePasswordTextRedacted(t *testing.T) {
	app := Map(sampleTree())
	for i := range app.Objects {
		if app.Objects[i].ID == "id:pw" {
			// Bypass Map's own redaction to prove the serving layer is the
			// boundary that enforces the invariant.
			app.Objects[i].Value = "secret"
		}
	}
	id := int32(0)
	view := newServerView(app, DefaultName, dbusRef{}, 0, 0, 800, 600, &id)
	pwPath := objectPathFor(t, view, "id:pw")
	namePath := objectPathFor(t, view, "id:name")

	var body dbusEncoder
	body.int32Val(0)
	body.int32Val(-1)
	if got := decodeStringResult(t, callView(view, pwPath, atspiText, "GetText", body.bytes())); got != "" {
		t.Fatalf("password text = %q", got)
	}

	var char dbusEncoder
	char.int32Val(0)
	if got := decodeInt32Result(t, callView(view, pwPath, atspiText, "GetCharacterAtOffset", char.bytes())); got != 0 {
		t.Fatalf("password char = %d", got)
	}

	if got := decodeStringResult(t, callView(view, namePath, atspiText, "GetText", body.bytes())); got != "Ada" {
		t.Fatalf("plain text = %q", got)
	}
}

func TestServeCacheItems(t *testing.T) {
	view, _ := testServerView(t)
	res := callView(view, atspiCachePath, atspiCache, "GetItems", nil)
	if res.errName != "" {
		t.Fatalf("%s: %s", res.errName, res.errText)
	}
	if len(res.body) < 8 {
		t.Fatalf("cache body is %d bytes", len(res.body))
	}
	payload := int(binary.LittleEndian.Uint32(res.body[0:4]))
	if payload != len(res.body)-8 {
		t.Fatalf("cache payload = %d, body = %d", payload, len(res.body))
	}
}

func TestServeCacheItemsRejectsOversizedPayload(t *testing.T) {
	view, _ := testServerView(t)
	view.app.Objects[0].Name = strings.Repeat("a", int(maxBodyLen))
	res := callView(view, atspiCachePath, atspiCache, "GetItems", nil)
	if res.errName != "org.freedesktop.DBus.Error.LimitsExceeded" {
		t.Fatalf("error = %q %q", res.errName, res.errText)
	}
}
