package draw

import (
	"github.com/chinmay-sawant/ownframe/internal/emoji"
	"github.com/chinmay-sawant/ownframe/internal/page/form"
)

// emojiCSS sizes the replacement images to the surrounding text. The
// negative alignment sinks the picture onto the baseline like a glyph.
const emojiCSS = `<style>img[data-ownframe-emoji]{width:1em;height:1em;vertical-align:-0.125em}</style>`

// EmojiPass swaps supported emoji text for images and adds the sizing rule
// when anything changed. Field values and painted text share one pass.
// Caret spans are untouched because the replacer skips tags.
func EmojiPass(source string) string {
	out, changed := emoji.Replace(source)
	if !changed {
		return out
	}

	at := form.HeadOpen(out)
	if at < 0 {
		at = form.FindHead(out)
	}
	if at < 0 {
		return emojiCSS + out
	}

	return out[:at] + emojiCSS + out[at:]
}
