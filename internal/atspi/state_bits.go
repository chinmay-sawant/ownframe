package atspi

// AT-SPI state numbers from AtspiStateType in atspi-constants.h.
const (
	stateActive    = 1
	stateChecked   = 4
	stateEditable  = 7
	stateEnabled   = 8
	stateFocusable = 11
	stateFocused   = 12
	stateSensitive = 24
	stateShowing   = 25
	stateVisible   = 30
	stateCheckable = 41
	stateReadOnly  = 43
	stateWordBits  = 32
	stateWordCount = 2
)

// stateWords packs state flags into the au bit set. Two words cover
// every state this mapping can set, including READ_ONLY at bit 43.
func stateWords(s State) []uint32 {
	words := make([]uint32, stateWordCount)
	set := func(bit int, on bool) {
		if !on {
			return
		}
		words[bit/stateWordBits] |= 1 << uint(bit%stateWordBits)
	}
	set(stateActive, s.Active)
	set(stateChecked, s.Checked)
	set(stateEditable, s.Editable)
	set(stateEnabled, s.Enabled)
	set(stateFocusable, s.Focusable)
	set(stateFocused, s.Focused)
	set(stateSensitive, s.Sensitive)
	set(stateShowing, true)
	set(stateVisible, true)
	set(stateCheckable, s.Checkable)
	set(stateReadOnly, s.ReadOnly)
	return words
}
