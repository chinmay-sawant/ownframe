package window

import (
	"context"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

func TestPaceWindowOption(t *testing.T) {
	sh := &shell{ctx: context.Background()}

	out, stop := paceWindow(context.Background(), sh, Options{NoIdlePace: true, MaxFPS: -1})
	stop()

	if out != sh {
		t.Fatal("NoIdlePace uncapped must not wrap the game")
	}

	out, stop = paceWindow(context.Background(), sh, Options{})
	defer stop()

	if _, ok := out.(*paceGame); !ok {
		t.Fatalf("default pacing wrapped %T, want *paceGame", out)
	}
}

func TestPaceWindowCappedNoIdleStillWraps(t *testing.T) {
	sh := &shell{ctx: context.Background()}

	// paceWindow with a cap mutates the global TPS; restore it so later
	// tests in this package see the TPS they started with.
	saved := ebiten.TPS()
	defer ebiten.SetTPS(saved)

	out, stop := paceWindow(context.Background(), sh, Options{NoIdlePace: true, MaxFPS: 60})
	defer stop()

	pg, ok := out.(*paceGame)
	if !ok {
		t.Fatalf("capped NoIdlePace wrapped %T, want *paceGame", out)
	}
	if pg.maxFPS != 60 {
		t.Fatalf("capped maxFPS = %d, want 60", pg.maxFPS)
	}
	if !pg.noIdle {
		t.Fatal("capped NoIdlePace must set noIdle")
	}
}
