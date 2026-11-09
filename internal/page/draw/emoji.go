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
// outHash is the hash of out, so a caller that threads the input hash can
// reuse the output hash for the source memo without rehashing.
//
//nolint:gochecknoglobals // Single-entry memo (one source plus its output)
var emojiMemo = struct {
	sync.Mutex
	hash    uint64
	source  string
	out     string
	outHash uint64
	valid   bool
	probes  uint64
	hits    uint64
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
	out, _ := EmojiPassWithHash(source, form.HashSource(source))

	return out
}

// EmojiPassWithHash is EmojiPass when the caller already hashed source (the
// page sync threads one hash through the scan, the rewrite, and the emoji
// pass). Output is byte-identical to EmojiPass for the same source. It also
// returns the hash of the output: the input hash when the source carries no
// emoji, otherwise a single hash of the final output, so the caller can note
// the source without rehashing. The memo stays hash-keyed with an exact
// source compare.
func EmojiPassWithHash(source string, hash uint64) (string, uint64) {
	emojiMemo.Lock()
	if emojiMemo.valid && emojiMemo.hash == hash && emojiMemo.source == source {
		out, outHash := emojiMemo.out, emojiMemo.outHash
		emojiMemo.hits++
		emojiMemo.Unlock()

		return out, outHash
	}
	emojiMemo.Unlock()

	out, changed := emoji.Replace(source)
	if !changed {
		emojiMemo.Lock()
		emojiMemo.hash, emojiMemo.source, emojiMemo.out, emojiMemo.outHash, emojiMemo.valid = hash, source, out, hash, true
		emojiMemo.probes++
		emojiMemo.Unlock()

		return out, hash
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

	outHash := form.HashSource(out)

	emojiMemo.Lock()
	emojiMemo.hash, emojiMemo.source, emojiMemo.out, emojiMemo.outHash, emojiMemo.valid = hash, source, out, outHash, true
	emojiMemo.probes++
	emojiMemo.Unlock()

	return out, outHash
}
