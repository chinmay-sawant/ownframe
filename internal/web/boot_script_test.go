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
		`closest("[data-action], button, input[type=submit], a[href], [role=button]")`,
	} {
		if !strings.Contains(bootScript, want) {
			t.Errorf("bootScript is missing %q", want)
		}
	}
}

func TestBootScriptPrefersInteractiveClickAncestorToDescendantID(t *testing.T) {
	t.Parallel()

	interactiveTarget := `el.closest("[data-action], button, input[type=submit], a[href], [role=button]")`
	idFallback := `if (!target) { target = el.closest("[id]"); }`
	interactiveAt := strings.Index(bootScript, interactiveTarget)
	idFallbackAt := strings.Index(bootScript, idFallback)
	if interactiveAt < 0 || idFallbackAt < 0 {
		t.Fatalf("bootScript is missing interactive click target or id fallback")
	}
	if interactiveAt >= idFallbackAt {
		t.Fatalf("bootScript resolves descendant ids before interactive ancestors")
	}
}

func TestBootScriptForwardsWheelEvents(t *testing.T) {
	t.Parallel()

	for _, want := range []string{
		`function wheelUnits(delta, mode)`,
		`post("/wheel", body)`,
		`document.addEventListener("wheel", onWheel, { passive: false })`,
		`document.scrollingElement || document.documentElement`,
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
		`function postCaret(el)`,
		`if (hasCaret(el))`,
		`selectionStart`,
		`post("/change", body)`,
	} {
		if !strings.Contains(bootScript, want) {
			t.Errorf("bootScript is missing %q", want)
		}
	}
}

func TestBootScriptAppliesTickGeometryPatches(t *testing.T) {
	t.Parallel()

	for _, want := range []string{
		`op.geometry.left === "string"`,
		`el.style.left = op.geometry.left`,
		`el.style.top = op.geometry.top`,
		`el.style.width = op.geometry.width`,
		`el.style.height = op.geometry.height`,
	} {
		if !strings.Contains(bootScript, want) {
			t.Errorf("bootScript is missing %q", want)
		}
	}
}
