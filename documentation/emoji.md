# Emoji

Thirty-one emoji paint in full color in the window and in every picture it
produces. The bundled text faces draw a few smileys in
monochrome and tofu for the rest, so the page replaces supported
sequences in text with inline images backed by the bundled Twemoji 72px
PNGs (CC-BY 4.0, Twitter Inc; attribution lives in
`internal/emoji/emoji_table.go`).

Supported sequences, in display order: ❤️ (bare `❤` too) ⚠️ (bare
`⚠` too) 😂 😮 😢 😍 😊 😀 🎉 👍 🔥 🚀 ⭐ 👏 🎂 💯 ✅ ❌ 👋 🙏 🤍
😱 😭 🤔 💀 🫡 🥲 😴 ❓ ❗ 💡.
Everything else keeps the font's own paint: ZWJ families, skin-tone
modifiers, and flags stay monochrome or tofu.

The swap is paint-only and needs no page code. `Page.Redraw` runs it
after the control rewrite, so chat text, reaction chips, and typed field
values all show images while `FormValue`, copy, and select-all keep the
raw runes. Each picture is `img[data-ownframe-emoji]` at `1em`, sunk onto
the baseline; a theme that wants a different size styles that selector.
The window draws images with linear filtering, so a downscaled picture
keeps smooth edges.

Limits: the set is fixed in `internal/emoji`. A page cannot register
its own emoji through `SetImage`, though a template can always carry its
own `<img>` icons the usual way. Clicking exactly on an emoji picture in
a field lands the caret beside it.
