//go:build windows && (amd64 || arm64)

package uia

import "testing"

// TestAnswersObject pins which WM_GETOBJECT requests the window answers.
// UI Automation sends object IDs from zero up when it delivers an event; a
// window that answers only the root leaves those to the default window
// procedure and every event it raises is dropped.
func TestAnswersObject(t *testing.T) {
	answered := []int32{uiaRootObject, objidClient, 0, 1, 7, 1 << 20}
	for _, id := range answered {
		if !answersObject(id) {
			t.Errorf("object ID %d must be answered by the UI Automation provider", id)
		}
	}

	// OBJID_WINDOW, OBJID_SYSMENU, OBJID_TITLEBAR, OBJID_MENU, OBJID_HSCROLL,
	// OBJID_VSCROLL, OBJID_SIZEGRIP stay with the window.
	left := []int32{-1, -2, -3, -5, -6, -7, -8}
	for _, id := range left {
		if answersObject(id) {
			t.Errorf("object ID %d belongs to the window, not to the provider", id)
		}
	}
}

// TestProviderOptionsAreServerSide pins the value get_ProviderOptions
// returns. The enum is ClientSideProvider = 1, ServerSideProvider = 2; a
// provider hosted in the window it serves that reports 1 gets none of its
// events delivered.
func TestProviderOptionsAreServerSide(t *testing.T) {
	if providerOptionsServerSide != 0x2 {
		t.Fatalf("providerOptionsServerSide = %#x, want 0x2 (ProviderOptions_ServerSideProvider)", providerOptionsServerSide)
	}
}
