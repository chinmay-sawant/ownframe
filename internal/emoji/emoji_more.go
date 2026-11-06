package emoji

// tableMore extends the set with faces and symbols people type in
// messages. Without these the font paints nothing or a clipped gold
// fallback, which reads as a cut-off picture.
var tableMore = []entry{
	{seq: "😱", file: "1f631"},
	{seq: "😭", file: "1f62d"},
	{seq: "🤔", file: "1f914"},
	{seq: "💀", file: "1f480"},
	{seq: "🫡", file: "1fae1"},
	{seq: "🥲", file: "1f972"},
	{seq: "😴", file: "1f634"},
	{seq: "❓", file: "2753"},
	{seq: "❗", file: "2757"},
	{seq: "💡", file: "1f4a1"},
	{seq: "👀", file: "1f440"},
	{seq: "🧠", file: "1f9e0"},
	{seq: "📌", file: "1f4cc"},
	{seq: "🔋", file: "1f50b"},
	{seq: "🌮", file: "1f32e"},
	{seq: "🍜", file: "1f35c"},
	{seq: "🧈", file: "1f9c8"},
	{seq: "🦀", file: "1f980"},
	{seq: "🎨", file: "1f3a8"},
	{seq: "🎧", file: "1f3a7"},
	{seq: "👥", file: "1f465"},
}

// all returns every entry across the tables.
func all() []entry {
	out := make([]entry, 0, len(table)+len(tableMore))
	out = append(out, table...)
	out = append(out, tableMore...)

	return out
}
