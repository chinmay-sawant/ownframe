package web

import (
	"strings"
	"testing"
)

func TestBootScriptForwardsDOMKeyboardAndIdentifiedClicks(t *testing.T) {
	t.Parallel()

	for _, want := range []string{
		`document.addEventListener("keydown", onKeyDown)`,
		`document.addEventListener("keyup", onKeyUp)`,
		`post("/key"`,
		`closest("[data-action], button, input[type=submit], a[href], [role=button], [id]")`,
	} {
		if !strings.Contains(bootScript, want) {
			t.Errorf("bootScript is missing %q", want)
		}
	}
}

func TestBootScriptForwardsWheelEvents(t *testing.T) {
	t.Parallel()

	for _, want := range []string{
		`function wheelUnits(delta, mode)`,
		`post("/wheel", body)`,
		`document.addEventListener("wheel", onWheel, { passive: false })`,
	} {
		if !strings.Contains(bootScript, want) {
			t.Errorf("bootScript is missing %q", want)
		}
	}
}

func TestBootScriptSyncsBrowserFocusAndCaret(t *testing.T) {
	t.Parallel()

	for _, want := range []string{
		`document.addEventListener("focusin", onFocusIn)`,
		`restoringFocus = true`,
		`body += "&caret="`,
		`post("/change", body)`,
	} {
		if !strings.Contains(bootScript, want) {
			t.Errorf("bootScript is missing %q", want)
		}
	}
}
