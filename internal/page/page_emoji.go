package page

import "github.com/chinmay-sawant/ownframe/internal/page/draw"

// emojiPass swaps supported emoji text for images and adds the sizing rule
// when anything changed. It runs after the control rewrite, so field
// values and painted text share one pass and caret spans are untouched.
func emojiPass(source string) string { return draw.EmojiPass(source) }
