package window

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

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
