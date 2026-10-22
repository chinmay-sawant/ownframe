package runtimebehavior

import "testing"

func TestAccessibilityAndFocusCoverage(t *testing.T) {
	a := AccessibilityFallback{}
	a.Bind("node", false)
	if !a.IsFallback() {
		t.Fatal("fallback broken")
	}
	a.Update(true)
	if a.IsFallback() {
		t.Fatal("update broken")
	}
	f := FocusRestorer{}
	f.Save("input", 2, [2]int{1, 3})
	f.SetCaret(5)
	f.SetSelection([2]int{0, 1})
	if !f.HasSaved() || !f.Restore("input") {
		t.Fatal("focus restore broken")
	}
	if got := WrapPromiseError("page", nil).Error(); got != "page: " {
		t.Fatalf("WrapPromiseError = %q, want %q", got, "page: ")
	}
	_ = NewObserver(nil)
}
