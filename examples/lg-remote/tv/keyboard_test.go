package tv

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func TestNoteKeyboardReadsFocus(t *testing.T) {
	for _, c := range []struct {
		name string
		msg  Message
		want KeyboardState
	}{
		{"focus", Message{Type: "response", Payload: json.RawMessage(`{"currentWidget":{"focus":true}}`)}, KeyboardOpen},
		{"blur", Message{Type: "response", Payload: json.RawMessage(`{"currentWidget":{"focus":false}}`)}, KeyboardClosed},
		{"none", Message{Type: "response", Payload: json.RawMessage(`{"subscribed":true}`)}, KeyboardClosed},
		{"error", Message{Type: "error", Error: "404 no such service"}, KeyboardUnreported},
	} {
		c0 := &Client{}
		c0.noteKeyboard(c.msg)
		if c0.imeState != c.want {
			t.Errorf("%s: state %d, want %d", c.name, c0.imeState, c.want)
		}
	}
}

func TestSessionKeyboardStateWithoutClient(t *testing.T) {
	if got := (&Session{}).KeyboardState(); got != KeyboardUnknown {
		t.Fatalf("state %d", got)
	}
}

func TestIMEErrorExplainsWhatToDo(t *testing.T) {
	if imeError(nil) != nil {
		t.Fatal("nil error changed")
	}
	if got := imeError(errors.New("the TV did not answer")); got.Error() != "the TV did not answer" {
		t.Fatalf("offline text changed: %v", got)
	}
	if got := imeError(errors.New("401 insufficient permissions")); !strings.Contains(got.Error(), "connect again") {
		t.Fatalf("permission text %v", got)
	}
	if got := imeError(errors.New("no active input")); !strings.Contains(got.Error(), "Open a search box on the TV") {
		t.Fatalf("generic text %v", got)
	}
}
