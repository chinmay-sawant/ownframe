package page

import "github.com/chinmay-sawant/ownframe/internal/page/form"

func wordBounds(s string, pos int) (int, int) { return form.WordBounds(s, pos) }

func lineBounds(s string, pos int) (int, int) { return form.LineBounds(s, pos) }

func runeAtByte(value string, off int) int { return form.RuneAtByte(value, off) }
