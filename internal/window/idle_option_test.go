package window

import (
	"context"
	"testing"
)

func TestPaceWindowOption(t *testing.T) {
	sh := &shell{ctx: context.Background()}

	out, stop := paceWindow(context.Background(), sh, Options{NoIdlePace: true})
	stop()

	if out != sh {
		t.Fatal("NoIdlePace must not wrap the game")
	}

	out, stop = paceWindow(context.Background(), sh, Options{})
	defer stop()

	if _, ok := out.(*paceGame); !ok {
		t.Fatalf("default pacing wrapped %T, want *paceGame", out)
	}
}
