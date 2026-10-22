package tv

import (
	"sync/atomic"
	"testing"
)

func TestHotstarDiscoversAndCachesInstalledID(t *testing.T) {
	var lists, launches atomic.Int32
	c := testClient(t, func(req map[string]any) map[string]any {
		switch req["uri"] {
		case "ssap://com.webos.applicationManager/listApps":
			lists.Add(1)
			return map[string]any{"apps": []any{map[string]any{"id": "regional-hotstar-id", "title": "JioHotstar"}}}
		case "ssap://com.webos.applicationManager/launch":
			p, ok := req["payload"].(map[string]any)
			if !ok {
				t.Error("payload is not a map")
				return map[string]any{}
			}
			if p["id"] != "regional-hotstar-id" {
				t.Error("guessed app ID")
			}
			launches.Add(1)
			return map[string]any{"returnValue": true}
		}
		return map[string]any{}
	})
	for range 2 {
		if _, err := c.Exec("hotstar:"); err != nil {
			t.Fatal(err)
		}
	}
	if lists.Load() != 1 || launches.Load() != 2 {
		t.Fatal("Hotstar discovery was not cached")
	}
}

func TestHotstarUnavailableDoesNotLaunchAnotherApp(t *testing.T) {
	c := testClient(t, func(req map[string]any) map[string]any {
		if req["uri"] != "ssap://com.webos.applicationManager/listApps" {
			t.Error("unexpected launch")
		}
		return map[string]any{"apps": []any{map[string]any{"id": "netflix", "title": "Netflix"}}}
	})
	if _, err := c.Exec("hotstar:"); err == nil {
		t.Fatal("missing app was accepted")
	}
}
