package player

import (
	"testing"
	"time"
)

func TestVirtualClockStaysOffWhileEngineMoves(t *testing.T) {
	app, voice := newFakeApp(t)
	redraw(t, app)

	click(t, app, "play")
	waitAudio(t, app)

	app.view.Playing = true
	voice.pos = 10 * time.Second

	if _, _, live := app.clock(); !live {
		t.Fatal("live engine wrongly treated as virtual")
	}

	if app.vengaged {
		t.Fatal("virtual clock engaged while engine advances")
	}
}

func TestVirtualClockEngagesAfterFreeze(t *testing.T) {
	app, voice := newFakeApp(t)
	redraw(t, app)

	click(t, app, "play")
	waitAudio(t, app)

	app.view.Playing = true
	voice.pos = 30 * time.Second

	// The first sample sees the new engine position and snaps to it.
	if _, _, live := app.clock(); !live {
		t.Fatal("fresh engine motion wrongly treated as virtual")
	}

	for i := 0; i < virtualFreezeTicks; i++ {
		if _, _, live := app.clock(); !live && i < virtualFreezeTicks-1 {
			t.Fatalf("engaged early at tick %d", i)
		}
	}

	if !app.vengaged {
		t.Fatal("virtual clock did not engage on frozen engine")
	}

	if pos, _, live := app.clock(); live || pos < 30*time.Second || pos > 31*time.Second {
		t.Fatalf("clock = %v live=%v, want frozen-then-virtual ~30s", pos, live)
	}
}

func TestNoteSeekArmsImmediately(t *testing.T) {
	app, _ := newFakeApp(t)
	redraw(t, app)

	app.view.Playing = true
	app.noteSeek(0.7)

	if !app.vengaged {
		t.Fatal("seek did not engage virtual clock")
	}

	want := time.Duration(0.7 * float64(app.vdur()))
	if pos, _, live := app.clock(); live || pos != want {
		t.Fatalf("clock = %v live=%v, want %v virtual", pos, live, want)
	}
}

func TestAdvanceVirtualNeedsPlaying(t *testing.T) {
	app, _ := newFakeApp(t)

	app.view.Playing = true
	app.advanceVirtual(5 * time.Second)

	if app.vpos != 5*time.Second {
		t.Fatalf("vpos = %v, want 5s", app.vpos)
	}

	app.view.Playing = false
	app.advanceVirtual(5 * time.Second)

	if app.vpos != 5*time.Second {
		t.Fatalf("paused virtual clock moved: %v", app.vpos)
	}
}

func TestVirtualEndedAtTrackEnd(t *testing.T) {
	app, _ := newFakeApp(t)

	if app.virtualEnded() {
		t.Fatal("fresh clock reports ended")
	}

	app.view.Playing = true
	app.noteSeek(1)
	app.advanceVirtual(time.Second)

	if !app.virtualEnded() {
		t.Fatalf("end of track not detected: vpos=%v vdur=%v", app.vpos, app.vdur())
	}
}

func TestSeekClickEngagesVirtual(t *testing.T) {
	app, voice := newFakeApp(t)
	redraw(t, app)

	click(t, app, "play")
	waitAudio(t, app)
	click(t, app, "seek-60")

	if !app.vengaged {
		t.Fatal("seek click did not engage virtual clock")
	}

	if voice.pos != time.Duration(0.6*float64(voice.dur)) {
		t.Fatalf("engine did not seek: pos=%v", voice.pos)
	}
}
