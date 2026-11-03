//go:build !ownframe_debug

package window

import (
	"context"
	"testing"
)

func TestDebugServerStaysOutOfRelease(t *testing.T) {
	game := newDevShell(newDevScreen())
	bound, stop, err := startDebugServer(context.Background(), game, "127.0.0.1:0")
	defer stop()
	if err != nil {
		t.Fatalf("stub: %v, want nil", err)
	}
	if bound != "" {
		t.Fatalf("bound = %q, want empty without the tag", bound)
	}
	if game.debugEnabled {
		t.Fatal("debug enabled without the tag")
	}
}
