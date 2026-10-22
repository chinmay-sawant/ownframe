package cat

import "testing"

func TestBidiStripped(t *testing.T) {
	t.Parallel()
	inbox := NewInbox()
	if err := inbox.Submit(Notification{Expression: "happy", Message: "a\u202eb\u2066c"}); err != nil {
		t.Fatal(err)
	}
	got, _ := inbox.Latest()
	if got.Message != "abc" {
		t.Fatalf("got %q", got.Message)
	}
}
