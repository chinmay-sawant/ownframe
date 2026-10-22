package page

import "github.com/chinmay-sawant/ownframe/internal/page/form"

const (
	// DefaultMaxFieldLen caps one field value when the template sets no
	// maxlength (GEN-28).
	DefaultMaxFieldLen = form.DefaultMaxFieldLen
	// MaxPasteRunes caps one paste or type call before it merges.
	MaxPasteRunes = form.MaxPasteRunes
)

func textLike(kind string) bool { return form.TextLike(kind) }

func canEdit(c Control) bool { return form.CanEdit(c) }

func isFileInput(c Control) bool { return form.IsFileInput(c) }

func willChange(c Control) bool { return form.WillChange(c) }

func fieldCap(c Control) int { return form.FieldCap(c) }

func clampField(s string, n int) string { return form.ClampField(s, n) }

func rangeOf(a, b, n int) (int, int) { return form.RangeOf(a, b, n) }

func clampPos(pos, n int) int { return form.ClampPos(pos, n) }

func runeLen(s string) int { return form.RuneLen(s) }

func shownText(kind, value string) string { return form.ShownText(kind, value) }

func pickOption(opts []Option) string { return form.PickOption(opts) }

func nextOption(c Control) Control { return form.NextOption(c) }

func selectedIndex(opts []Option) int { return form.SelectedIndex(opts) }

func setRunes(c Control, runes []rune) Control { return form.SetRunes(c, runes) }
