package window

import (
	"github.com/chinmay-sawant/ownframe/internal/clipboard"
	"github.com/chinmay-sawant/ownframe/internal/host"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"strings"
	"time"
)

func (s *shell) keys() error {
	mods := readModifiers()

	ate, err := s.devtoolsKeys(mods)
	if err != nil {
		return err
	}

	if err := s.keyEvents(mods); err != nil {
		return err
	}

	if typingSuppressed(mods) && backspaceDue() {
		return s.kbd().DeleteWord(s.ctx)
	}

	if chord := s.chords.pressed(mods); chord != chordNone {
		return s.applyChord(chord)
	}

	if typingSuppressed(mods) || ate || s.imeHandled() {
		return s.submitIfEnter()
	}

	s.chars = ebiten.AppendInputChars(s.chars[:0])
	s.chars = s.chords.filter(s.chars)
	text := strings.ReplaceAll(string(s.chars), "\r", "")
	text = strings.ReplaceAll(text, "\n", "")

	if text != "" {
		if err := s.kbd().Type(s.ctx, text); err != nil {
			return err
		}
	}

	if backspaceDue() {
		if err := s.kbd().Backspace(s.ctx); err != nil {
			return err
		}
	}

	return s.submitIfEnter()
}

func readModifiers() modifiers {
	return modifiers{
		Control: ebiten.IsKeyPressed(ebiten.KeyControl),
		Meta:    ebiten.IsKeyPressed(ebiten.KeyMeta),
		Alt:     ebiten.IsKeyPressed(ebiten.KeyAlt),
		Shift:   ebiten.IsKeyPressed(ebiten.KeyShift),
	}
}

// keyEvents sends each key press and release to the screen in the frame it
// happens. A key handler does not draw; the screen paints on its own.
// The window handles Tab and Escape before the screen sees them.
func (s *shell) keyEvents(mods modifiers) error {
	for key := ebiten.Key(0); key <= ebiten.KeyMax; key++ {
		down, up := s.pageKeyStep(key, ebiten.IsKeyPressed(key))
		if !down && !up {
			continue
		}

		name := keyEventName(key, mods)
		if name == "" {
			continue
		}

		if down {
			taken, err := s.windowKey(key, true, mods)
			if err != nil {
				return err
			}

			if !taken {
				if err := s.kbd().KeyDown(s.ctx, name); err != nil {
					return err
				}
			}
		}

		if up {
			taken, err := s.windowKey(key, false, mods)
			if err != nil {
				return err
			}

			if !taken {
				if err := s.kbd().KeyUp(s.ctx, name); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// keyEventName is the name the screen sees. The caret keys carry the
// shift+, ctrl+, and alt+ prefixes the page understands, so the window can
// extend a selection or jump by word. Every other key keeps its plain name.
func keyEventName(key ebiten.Key, mods modifiers) string {
	name := keyName(key)
	if name == "" {
		return ""
	}

	switch name {
	case "arrowleft", "arrowright", "home", "end":
	default:
		return name
	}

	if mods.Shift {
		name = "shift+" + name
	}

	if mods.command() {
		name = "ctrl+" + name
	} else if mods.Alt {
		name = "alt+" + name
	}

	return name
}

// keyCount is one slot per key Ebiten can report.
const keyCount = int(ebiten.KeyMax) + 1

// keyGuardFrames is how many frames a key must be seen up before another
// press counts, and how many frames it must be held before its release
// counts. A server that reports auto-repeat as a release and a press leaves
// only one frame between the two, so the guard drops the pair while a real
// tap still reports both events.
const keyGuardFrames = 2

// keyWatch reports one press and one release per real key event, not per
// auto-repeat pulse.
type keyWatch struct {
	down [keyCount]int
	up   [keyCount]int
}

// newKeyWatch returns a watch whose keys start up, so the first press fires.
func newKeyWatch() keyWatch {
	var w keyWatch

	for i := range w.up {
		w.up[i] = keyGuardFrames
	}

	return w
}

// step advances one key by one frame and reports the press and the release
// a page should see in this frame.
func (w *keyWatch) step(key ebiten.Key, pressed bool) (down, up bool) {
	i := int(key)
	if i < 0 || i >= keyCount {
		return false, false
	}

	if pressed {
		fire := w.down[i] == 0 && w.up[i] >= keyGuardFrames
		w.up[i] = 0
		w.down[i]++

		return fire, false
	}

	fire := w.down[i] >= keyGuardFrames
	w.down[i] = 0

	if w.up[i] < keyGuardFrames {
		w.up[i]++
	}

	return false, fire
}

// keyName is the name a key handler sees: the Ebiten key name lowercased,
// with the Digit prefix removed, so "Space" is "space", "ArrowUp" is
// "arrowup", "A" is "a", and "Digit1" is "1". An unknown key is "".
func keyName(key ebiten.Key) string {
	name := strings.ToLower(key.String())

	return strings.TrimPrefix(name, "digit")
}

// windowKey lets the window act on a key before the page handler sees it.
// It reports whether the key was consumed, so the handler stays unaware of
// the keys the window owns.
func (s *shell) windowKey(key ebiten.Key, down bool, mods modifiers) (bool, error) {
	switch key {
	case ebiten.KeyTab:
		return s.tabKey(down, mods)
	case ebiten.KeyEscape:
		if down {
			wasOpen := s.menu.open
			if err := s.escape(); err != nil {
				return false, err
			}

			return wasOpen, nil
		}
	case ebiten.KeyF11:
		return s.f11(down), nil
	case ebiten.KeyEqual, ebiten.KeyKPAdd,
		ebiten.KeyMinus, ebiten.KeyKPSubtract,
		ebiten.Key0, ebiten.KeyKP0:
		return s.zoomKey(key, down, mods), nil
	}

	return false, nil
}

// tabKey moves focus with Tab and Shift+Tab. The key is consumed only when a
// field took the focus, so a page without fields keeps the key.
func (s *shell) tabKey(down bool, mods modifiers) (bool, error) {
	if !down {
		taken := s.tabEaten
		s.tabEaten = false

		return taken, nil
	}

	focuser, ok := s.app.(host.Focuser)
	if !ok {
		return false, nil
	}

	var err error
	if mods.Shift {
		err = focuser.FocusPrev(s.ctx)
	} else {
		err = focuser.FocusNext(s.ctx)
	}

	if err != nil {
		return false, err
	}

	if focuser.FocusID() == "" {
		return false, nil
	}

	s.tabEaten = true

	return true, nil
}

// escape closes the context menu first, then clears keyboard focus.
func (s *shell) escape() error {
	if s.menu.open {
		s.closeMenu()

		return nil
	}

	focuser, ok := s.app.(host.Focuser)
	if !ok || focuser.FocusID() == "" {
		return nil
	}
	if err := s.imeCancelSession(); err != nil {
		return err
	}

	return focuser.Focus(s.ctx, "")
}

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

func (s *shell) submitIfEnter() error {
	enter := inpututil.IsKeyJustPressed(ebiten.KeyEnter)
	numpad := inpututil.IsKeyJustPressed(ebiten.KeyNumpadEnter)
	if enter || numpad {
		return s.kbd().Submit(s.ctx)
	}

	return nil
}

func backspaceDue() bool {
	held := inpututil.KeyPressDuration(ebiten.KeyBackspace)
	if held == 1 {
		return true
	}

	if held <= backspaceRepeatAt {
		return false
	}

	return (held-backspaceRepeatAt)%backspaceRepeatEvery == 0
}

// filter drops the runes of a chord key that already fired and is still
// down. The key can keep repeating after the modifier goes up, and those
// repeats must not reach the focused field. A key is forgotten only after
// its release has been quiet for the repeat gap, so a repeat reported as a
// release and a press keeps the key down.
func (w *chordWatch) filter(runes []rune) []rune {
	kept := runes[:0]

	for _, r := range runes {
		if !w.swallows(r) {
			kept = append(kept, r)
		}
	}

	for i, key := range chordKeys {
		if ebiten.IsKeyPressed(key) {
			continue
		}

		if w.released[i].IsZero() || time.Since(w.released[i]) >= chordRepeatGap {
			w.eaten[i] = ""
		}
	}

	return kept
}

// swallows reports whether r belongs to a fired key that is still down.
func (w *chordWatch) swallows(r rune) bool {
	for _, name := range w.eaten {
		if name != "" && strings.EqualFold(string(r), name) {
			return true
		}
	}

	return false
}

// chordKeys are the keys that start a text chord.
var chordKeys = [...]ebiten.Key{
	ebiten.KeyC,
	ebiten.KeyV,
	ebiten.KeyX,
	ebiten.KeyA,
	ebiten.KeyZ,
	ebiten.KeyY,
	ebiten.KeyInsert,
	ebiten.KeyDelete,
}

// chordRepeatGap is the shortest time a key can be up between two real
// presses. A server that reports auto-repeat as a release and a press
// leaves far less than this between the two.
const chordRepeatGap = 50 * time.Millisecond

// chordWatch fires a chord once per press, and keeps a fired key from
// typing while it stays down. A server can keep repeating the key after
// the modifier is released, and those repeats must not turn into text.
type chordWatch struct {
	released [len(chordKeys)]time.Time
	eaten    [len(chordKeys)]string
}

// newChordWatch returns a watch ready for the first press.
func newChordWatch() chordWatch {
	return chordWatch{}
}

// pressed returns the chord one just-pressed key starts, if any.
func (w *chordWatch) pressed(mods modifiers) chord {
	now := time.Now()

	for i, key := range chordKeys {
		justPressed := inpututil.IsKeyJustPressed(key)
		justReleased := inpututil.IsKeyJustReleased(key)

		if !w.step(i, justPressed, justReleased, now) {
			continue
		}

		if c := shortcutChord(mods, key); c != chordNone {
			w.eaten[i] = ebiten.KeyName(key)

			return c
		}
	}

	return chordNone
}

// step records one key's frame and reports a press that is not a repeat.
func (w *chordWatch) step(i int, justPressed, justReleased bool, now time.Time) bool {
	if justReleased {
		w.released[i] = now
	}

	if !justPressed {
		return false
	}

	return w.released[i].IsZero() || now.Sub(w.released[i]) >= chordRepeatGap
}
