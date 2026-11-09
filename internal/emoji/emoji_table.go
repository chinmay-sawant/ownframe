// Package emoji paints forty-two color emoji as inline images. The bundled
// text faces draw a few smileys in monochrome and tofu for the rest, so
// supported sequences are replaced with <img> tags that resolve to the
// bundled Twemoji 72px PNGs through the page image lookup. Form values
// keep the raw runes; only the paint carries images.
//
// The graphics are Twemoji by Twitter, Inc and other contributors,
// licensed CC-BY 4.0 (https://creativecommons.org/licenses/by/4.0/).
// The set is fixed: six channel reactions, fifteen common faces,
// gestures, and symbols, plus twenty-one message faces and marks. ZWJ families
// and skin-tone modifiers are out of scope and keep the font's own paint.
package emoji

// entry maps one rune sequence to its PNG file stem.
type entry struct {
	seq  string
	file string
}

// table lists every supported sequence, two-rune forms first. The bare
// hearts match with or without the emoji variation selector.
var table = []entry{
	{seq: "❤️", file: "2764"},
	{seq: "⚠️", file: "26a0"},
	{seq: "❤", file: "2764"},
	{seq: "⚠", file: "26a0"},
	{seq: "😂", file: "1f602"},
	{seq: "😮", file: "1f62e"},
	{seq: "😢", file: "1f622"},
	{seq: "😍", file: "1f60d"},
	{seq: "😊", file: "1f60a"},
	{seq: "😀", file: "1f600"},
	{seq: "🎉", file: "1f389"},
	{seq: "👍", file: "1f44d"},
	{seq: "🔥", file: "1f525"},
	{seq: "🚀", file: "1f680"},
	{seq: "⭐", file: "2b50"},
	{seq: "👏", file: "1f44f"},
	{seq: "🎂", file: "1f382"},
	{seq: "💯", file: "1f4af"},
	{seq: "✅", file: "2705"},
	{seq: "❌", file: "274c"},
	{seq: "👋", file: "1f44b"},
	{seq: "🙏", file: "1f64f"},
	{seq: "🤍", file: "1f90d"},
}

// bySeq maps each sequence to its file stem for the replacer.
var bySeq = func() map[string]string {
	all := make([]entry, 0, len(table)+len(tableMore))
	all = append(all, table...)
	all = append(all, tableMore...)

	m := make(map[string]string, len(all))
	for _, e := range all {
		m[e.seq] = e.file
	}

	return m
}()
