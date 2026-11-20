package window

import (
	"context"
	"github.com/hajimehoshi/ebiten/v2"
	"math"
	"testing"
	"time"
)

// TestChordWatchDropsRepeats checks the auto-repeat shapes a server can
// report: a release and a press that share one frame, and a press a frame
// or two after the release. A press that comes after the gap is real.
func TestChordWatchDropsRepeats(t *testing.T) {
	t.Parallel()

	w := newChordWatch()
	t0 := time.Unix(0, 0)

	// A real press fires once, and a held key does not fire again.
	if !w.step(0, true, false, t0) {
		t.Fatal("a real press did not fire")
	}

	if w.step(0, false, false, t0.Add(20*time.Millisecond)) {
		t.Fatal("a held key fired again")
	}

	// The release and a press in the same frame are one repeat pulse.
	if w.step(0, true, true, t0.Add(60*time.Millisecond)) {
		t.Fatal("a same-frame release and press fired")
	}

	// A press a frame after the release is another repeat pulse.
	if w.step(0, true, false, t0.Add(70*time.Millisecond)) {
		t.Fatal("a press right after a release fired")
	}

	// A press after the gap is a real press.
	if !w.step(0, true, false, t0.Add(200*time.Millisecond)) {
		t.Fatal("a press after the gap did not fire")
	}
}

// TestChordWatchSwallowsFiredKey checks that a chord key stops typing
// after its chord fired, and that it types again once it is up.
func TestChordWatchSwallowsFiredKey(t *testing.T) {
	t.Parallel()

	w := newChordWatch()
	w.eaten[0] = "c"

	if got := string(w.filter([]rune("cv"))); got != "v" {
		t.Fatalf("filter = %q, want v", got)
	}

	// The key is not pressed in this test, so the filter forgot it.
	if got := string(w.filter([]rune("c"))); got != "c" {
		t.Fatalf("filter after the key went up = %q, want c", got)
	}
}

func TestEscapeClearsFocus(t *testing.T) {
	t.Parallel()

	app := &focusScreen{fakeScreen: &fakeScreen{}, focus: "a"}
	s := &shell{app: app, ctx: context.Background()}

	if err := s.escape(); err != nil {
		t.Fatal(err)
	}

	if app.focus != "" {
		t.Fatalf("focus = %q", app.focus)
	}

	if err := s.escape(); err != nil {
		t.Fatal(err)
	}

	s = &shell{app: &fakeScreen{}, ctx: context.Background()}
	if err := s.escape(); err != nil {
		t.Fatal(err)
	}
}

func TestTabConsumesWhenAFieldTakesFocus(t *testing.T) {
	t.Parallel()

	app := &focusScreen{fakeScreen: &fakeScreen{}, ids: []string{"a", "b"}}
	s := &shell{app: app, ctx: context.Background()}

	taken, err := s.tabKey(true, modifiers{})
	if err != nil || !taken {
		t.Fatalf("tab = %v, %v", taken, err)
	}

	if app.focus != "a" || app.nexts != 1 {
		t.Fatalf("focus = %q nexts = %d", app.focus, app.nexts)
	}

	taken, err = s.tabKey(false, modifiers{})
	if err != nil || !taken {
		t.Fatalf("tab up = %v, %v", taken, err)
	}

	taken, err = s.tabKey(true, modifiers{Shift: true})
	if err != nil || !taken {
		t.Fatalf("shift tab = %v, %v", taken, err)
	}

	if app.focus != "b" || app.prevs != 1 {
		t.Fatalf("focus = %q prevs = %d", app.focus, app.prevs)
	}
}

func TestTabPassesThroughWithoutFields(t *testing.T) {
	t.Parallel()

	s := &shell{app: &fakeScreen{}, ctx: context.Background()}

	taken, err := s.tabKey(true, modifiers{})
	if err != nil || taken {
		t.Fatalf("plain tab = %v, %v", taken, err)
	}

	app := &focusScreen{fakeScreen: &fakeScreen{}}
	s = &shell{app: app, ctx: context.Background()}

	taken, err = s.tabKey(true, modifiers{})
	if err != nil || taken {
		t.Fatalf("empty tab = %v, %v", taken, err)
	}

	if app.nexts != 1 {
		t.Fatalf("nexts = %d", app.nexts)
	}
}

func TestKeyEventNamePrefixesCaretKeys(t *testing.T) {
	t.Parallel()

	cases := []struct {
		key  ebiten.Key
		mods modifiers
		want string
	}{
		{ebiten.KeyArrowLeft, modifiers{}, "arrowleft"},
		{ebiten.KeyArrowLeft, modifiers{Shift: true}, "shift+arrowleft"},
		{ebiten.KeyArrowRight, modifiers{Control: true}, "ctrl+arrowright"},
		{ebiten.KeyArrowRight, modifiers{Alt: true}, "alt+arrowright"},
		{ebiten.KeyHome, modifiers{Control: true, Shift: true}, "ctrl+shift+home"},
		{ebiten.KeyEnd, modifiers{Meta: true}, "ctrl+end"},
		{ebiten.KeyA, modifiers{Control: true}, "a"},
		{ebiten.KeySpace, modifiers{Shift: true}, "space"},
	}

	for _, c := range cases {
		if got := keyEventName(c.key, c.mods); got != c.want {
			t.Fatalf("keyEventName(%v, %+v) = %q, want %q", c.key, c.mods, got, c.want)
		}
	}
}

// TestKeyWatchDropsRepeatPulses checks the auto-repeat shape a server can
// report: a release and a press with only one frame between them.
func TestKeyWatchDropsRepeatPulses(t *testing.T) {
	t.Parallel()

	w := newKeyWatch()

	if down, _ := w.step(0, true); !down {
		t.Fatal("the first press did not fire")
	}

	for range keyGuardFrames {
		if down, up := w.step(0, false); down || up {
			t.Fatal("a release before the guard fired")
		}

		if down, _ := w.step(0, true); down {
			t.Fatal("a repeat pulse fired a press")
		}
	}
}

func TestKeyWatchFiresRealEvents(t *testing.T) {
	t.Parallel()

	w := newKeyWatch()

	if down, _ := w.step(0, true); !down {
		t.Fatal("the first press did not fire")
	}

	for range keyGuardFrames - 1 {
		if down, up := w.step(0, true); down || up {
			t.Fatal("a held key fired again")
		}
	}

	if _, up := w.step(0, false); !up {
		t.Fatal("the release did not fire")
	}

	for range keyGuardFrames {
		w.step(0, false)
	}

	if down, _ := w.step(0, true); !down {
		t.Fatal("the next real press did not fire")
	}
}

func TestKeyName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		key  ebiten.Key
		want string
	}{
		{key: ebiten.KeySpace, want: "space"},
		{key: ebiten.KeyArrowUp, want: "arrowup"},
		{key: ebiten.KeyArrowDown, want: "arrowdown"},
		{key: ebiten.KeyEnter, want: "enter"},
		{key: ebiten.KeyEscape, want: "escape"},
		{key: ebiten.KeyA, want: "a"},
		{key: ebiten.KeyZ, want: "z"},
		{key: ebiten.KeyDigit1, want: "1"},
		{key: ebiten.Key0, want: "0"},
		{key: ebiten.KeyF5, want: "f5"},
		{key: ebiten.Key(9999), want: ""},
	}

	for _, test := range tests {
		if got := keyName(test.key); got != test.want {
			t.Errorf("keyName(%v) = %q, want %q", test.key, got, test.want)
		}
	}
}

func TestShortcutChords(t *testing.T) {
	t.Parallel()

	ctrl := modifiers{Control: true}
	cmd := modifiers{Meta: true}
	altGr := modifiers{Control: true, Alt: true}

	tests := []struct {
		name string
		mods modifiers
		key  ebiten.Key
		want chord
	}{
		{name: "ctrl c", mods: ctrl, key: ebiten.KeyC, want: chordCopy},
		{name: "cmd c", mods: cmd, key: ebiten.KeyC, want: chordCopy},
		{name: "ctrl insert", mods: ctrl, key: ebiten.KeyInsert, want: chordCopy},
		{name: "ctrl v", mods: ctrl, key: ebiten.KeyV, want: chordPaste},
		{name: "ctrl shift v", mods: modifiers{Control: true, Shift: true}, key: ebiten.KeyV, want: chordNone},
		{name: "cmd shift v", mods: modifiers{Meta: true, Shift: true}, key: ebiten.KeyV, want: chordNone},
		{name: "shift insert", mods: modifiers{Shift: true}, key: ebiten.KeyInsert, want: chordPaste},
		{name: "ctrl x", mods: ctrl, key: ebiten.KeyX, want: chordCut},
		{name: "shift delete", mods: modifiers{Shift: true}, key: ebiten.KeyDelete, want: chordCut},
		{name: "ctrl a", mods: ctrl, key: ebiten.KeyA, want: chordSelectAll},
		{name: "ctrl z", mods: ctrl, key: ebiten.KeyZ, want: chordUndo},
		{name: "ctrl shift z", mods: modifiers{Control: true, Shift: true}, key: ebiten.KeyZ, want: chordRedo},
		{name: "ctrl y", mods: ctrl, key: ebiten.KeyY, want: chordRedo},
		{name: "plain c", mods: modifiers{}, key: ebiten.KeyC, want: chordNone},
		{name: "altgr c", mods: altGr, key: ebiten.KeyC, want: chordNone},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			if got := shortcutChord(test.mods, test.key); got != test.want {
				t.Fatalf("chord = %d, want %d", got, test.want)
			}
		})
	}
}

func TestAltGrStillTypes(t *testing.T) {
	t.Parallel()

	if typingSuppressed(modifiers{Control: true, Alt: true}) {
		t.Fatal("altgr suppressed typing")
	}

	if !typingSuppressed(modifiers{Control: true}) {
		t.Fatal("ctrl did not suppress typing")
	}
}

// TestZoomByClamps checks the keyboard zoom steps and the range clamp.
func TestZoomByClamps(t *testing.T) {
	s := &shell{pageZoom: 1}

	s.zoomBy(zoomStep)
	if math.Abs(s.zoom()-zoomStep) > 1e-9 {
		t.Fatalf("zoom after one step = %v", s.zoom())
	}

	for i := 0; i < 100; i++ {
		s.zoomBy(zoomStep)
	}

	if s.zoom() != maxZoom {
		t.Fatalf("zoom at the top = %v", s.zoom())
	}

	for i := 0; i < 200; i++ {
		s.zoomBy(1 / zoomStep)
	}

	if s.zoom() != minZoom {
		t.Fatalf("zoom at the bottom = %v", s.zoom())
	}
}

// TestZoomKey checks that only the command modifier consumes =, -, and 0.
func TestZoomKey(t *testing.T) {
	s := &shell{pageZoom: 1}

	if !s.zoomKey(ebiten.KeyEqual, true, modifiers{Control: true}) {
		t.Fatal("Control+= was not consumed")
	}

	if s.zoomKey(ebiten.KeyEqual, true, modifiers{}) {
		t.Fatal("plain = was consumed")
	}

	before := s.zoom()
	s.zoomKey(ebiten.KeyMinus, true, modifiers{Control: true})

	if s.zoom() >= before {
		t.Fatalf("Control+- zoom = %v, before %v", s.zoom(), before)
	}

	s.zoomKey(ebiten.Key0, true, modifiers{Control: true})

	if s.zoom() != 1 {
		t.Fatalf("zoom reset = %v", s.zoom())
	}
}

type noWheelZoomScreen struct{ fakeScreen }

func (noWheelZoomScreen) WheelZoomAllowed() bool { return false }

type noKeyZoomScreen struct{ fakeScreen }

func (noKeyZoomScreen) KeyZoomAllowed() bool { return false }

func TestDisabledWheelZoomStaysStill(t *testing.T) {
	off := &shell{app: &noWheelZoomScreen{}, pageZoom: 1}
	if off.wheelMayZoom(1) {
		t.Fatal("a page that refuses wheel zoom still accepted a notch")
	}

	if off.zoom() != 1 {
		t.Fatalf("zoom = %v", off.zoom())
	}

	on := &shell{app: &fakeScreen{}, pageZoom: 1}
	if !on.wheelMayZoom(1) {
		t.Fatal("the default page rejected wheel zoom")
	}
}

func TestDisabledKeyZoomStaysStill(t *testing.T) {
	off := &shell{app: &noKeyZoomScreen{}, pageZoom: 1}
	mods := modifiers{Control: true}
	keys := []ebiten.Key{
		ebiten.KeyEqual,
		ebiten.KeyKPAdd,
		ebiten.KeyMinus,
		ebiten.KeyKPSubtract,
		ebiten.Key0,
		ebiten.KeyKP0,
	}

	for _, key := range keys {
		if off.zoomKey(key, true, mods) {
			t.Fatalf("Control+%v was consumed", key)
		}
	}

	if off.zoom() != 1 {
		t.Fatalf("zoom = %v", off.zoom())
	}

	plain := &shell{pageZoom: 1}
	plain.zoomKey(ebiten.KeyEqual, true, mods)

	if plain.zoom() == 1 {
		t.Fatal("a page without the policy rejected Control+=")
	}
}

// TestWheelZoomFactor checks trackpad deltas scale by the same step.
func TestWheelZoomFactor(t *testing.T) {
	if got := wheelZoomFactor(0); got != 1 {
		t.Fatalf("wheelZoomFactor(0) = %v", got)
	}

	if got := wheelZoomFactor(1); math.Abs(got-zoomStep) > 1e-9 {
		t.Fatalf("wheelZoomFactor(1) = %v", got)
	}

	half := wheelZoomFactor(0.5)

	if math.Abs(half*half-zoomStep) > 1e-9 {
		t.Fatalf("half notch = %v", half)
	}
}

// TestZoomCombinesPinch checks the page zoom multiplies into the pinch scale.
func TestZoomCombinesPinch(t *testing.T) {
	s := &shell{pageZoom: 2}
	s.fingers.zoom = 1.5

	if got := s.zoom(); math.Abs(got-3) > 1e-9 {
		t.Fatalf("combined zoom = %v", got)
	}
}
