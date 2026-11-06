package draw

import (
	"sync"

	"github.com/chinmay-sawant/ownframe/internal/emoji"
	"github.com/chinmay-sawant/ownframe/internal/page/form"
)

// emojiCSS sizes the replacement images to the surrounding text. The
// negative alignment sinks the picture onto the baseline like a glyph.
const emojiCSS = `<style>img[data-ownframe-emoji]{width:1em;height:1em;vertical-align:-0.125em}</style>`

// EmojiTelemetry counts emoji passes so tests can prove the probe is
// skipped: a repeat source must move Hits without moving Probes.
type EmojiTelemetry struct {
	Probes uint64
	Hits   uint64
}

// its pass counters. The entry keys on the source hash with an exact source
// compare, so a collision degrades to a rescan instead of a stale page. The
// probe counter is the only way for tests to observe the skipped probe.
//
//nolint:gochecknoglobals // Single-entry memo (one source plus its output)
var emojiMemo = struct {
	sync.Mutex
	hash   uint64
	source string
	out    string
	valid  bool
	probes uint64
	hits   uint64
}{}

// EmojiStats returns the emoji pass counters.
func EmojiStats() EmojiTelemetry {
	emojiMemo.Lock()
	defer emojiMemo.Unlock()

	return EmojiTelemetry{Probes: emojiMemo.probes, Hits: emojiMemo.hits}
}

// EmojiPass swaps supported emoji text for images and adds the sizing rule
// when anything changed. Field values and painted text share one pass.
// Caret spans are untouched because the replacer skips tags.
//
// The probe is hash-gated: a repeat of the last source returns the memoized
// output without running the emoji scan again. The gate keys on the source
// hash and validates with an exact source compare.
func EmojiPass(source string) string {
	hash := form.HashSource(source)

	emojiMemo.Lock()
	if emojiMemo.valid && emojiMemo.hash == hash && emojiMemo.source == source {
		out := emojiMemo.out
		emojiMemo.hits++
		emojiMemo.Unlock()

		return out
	}
	emojiMemo.Unlock()

	out, changed := emoji.Replace(source)
	if !changed {
		emojiMemo.Lock()
		emojiMemo.hash, emojiMemo.source, emojiMemo.out, emojiMemo.valid = hash, source, out, true
		emojiMemo.probes++
		emojiMemo.Unlock()

		return out
	}

	at := form.HeadOpen(out)
	if at < 0 {
		at = form.FindHead(out)
	}
	if at < 0 {
		out = emojiCSS + out
	} else {
		out = out[:at] + emojiCSS + out[at:]
	}

	emojiMemo.Lock()
	emojiMemo.hash, emojiMemo.source, emojiMemo.out, emojiMemo.valid = hash, source, out, true
	emojiMemo.probes++
	emojiMemo.Unlock()

	return out
}
