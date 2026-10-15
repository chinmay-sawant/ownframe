package window

import "strings"

import "github.com/chinmay-sawant/ownframe/internal/clipboard"

func (s *shell) applyChord(chord chord) error {
	switch chord {
	case chordCopy:
		return s.copy(false)
	case chordCut:
		return s.copy(true)
	case chordPaste:
		text, err := clipboard.Read()
		if err != nil {
			return nil
		}

		text = strings.ReplaceAll(text, "\r", "")
		text = strings.ReplaceAll(text, "\n", "")

		return s.clip().Paste(s.ctx, text)
	case chordSelectAll:
		return s.clip().SelectAll(s.ctx)
	case chordUndo:
		return s.clip().Undo(s.ctx)
	case chordRedo:
		return s.clip().Redo(s.ctx)
	default:
		return nil
	}
}

func (s *shell) copy(cut bool) error {
	var text string
	var ok bool
	var err error

	if cut {
		text, ok, err = s.clip().Cut(s.ctx)
	} else {
		text, ok, err = s.clip().Copy(s.ctx)
	}

	if err != nil || !ok || text == "" {
		return err
	}

	clipboard.Write(text)

	return nil
}
