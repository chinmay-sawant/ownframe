package window

import "github.com/hajimehoshi/ebiten/v2"

type chord int

const (
	chordNone chord = iota
	chordCopy
	chordCut
	chordPaste
	chordSelectAll
	chordUndo
	chordRedo
)

type modifiers struct {
	Control bool
	Meta    bool
	Alt     bool
	Shift   bool
}

// command is Control or Command, and not AltGr.
// AltGr is Control and Alt together, and it types a character.
func (m modifiers) command() bool {
	if m.Alt && !m.Meta {
		return false
	}

	return m.Control || m.Meta
}

// typingSuppressed reports that printable input belongs to a shortcut.
func typingSuppressed(m modifiers) bool {
	return m.command()
}

// shortcutChord maps one just-pressed key to an editing shortcut.
func shortcutChord(m modifiers, key ebiten.Key) chord {
	command := m.command()

	switch key {
	case ebiten.KeyC:
		if command {
			return chordCopy
		}
	case ebiten.KeyInsert:
		if command {
			return chordCopy
		}

		if m.Shift {
			return chordPaste
		}
	case ebiten.KeyV:
		if command {
			// A command with Shift is a page chord, not a text edit.
			// Only the unshifted chord pastes.
			if m.Shift {
				return chordNone
			}

			return chordPaste
		}
	case ebiten.KeyX:
		if command {
			return chordCut
		}
	case ebiten.KeyDelete:
		if m.Shift && !m.Alt {
			return chordCut
		}
	case ebiten.KeyA:
		if command {
			return chordSelectAll
		}
	case ebiten.KeyZ:
		if !command {
			return chordNone
		}

		if m.Shift {
			return chordRedo
		}

		return chordUndo
	case ebiten.KeyY:
		if command {
			return chordRedo
		}
	}

	return chordNone
}
