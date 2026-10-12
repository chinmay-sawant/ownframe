# Theming

The template's own `<style>` elements and `style=` attributes are the base
look. `Config.Theme` and `Page.SetTheme` add one extra stylesheet after the
template's `<style>` elements. Any property the engine implements can appear
in the theme, including custom properties the template reads with `var()`.
The engine walks the extra sheet's rules last, but source order is numbered
per sheet: a theme rule wins a tie of equal specificity only when its index
in the theme sheet is at least the template rule's index. A theme declaration
marked `!important` beats any normal template declaration; a `style=`
attribute beats a normal theme declaration.

```go
page, err := gpui.New(gpui.Config{
    Title:  "Hello",
    HTML:   templateHTML, // reads var(--accent, #1a56db)
    Theme:  `:root { --accent: #7aa2f7; }`,
    Width:  480,
    Height: 640,
})
if err != nil {
    return err
}

if err := page.SetTheme(darkTheme); err != nil {
    return err
}
if err := page.Redraw(ctx); err != nil {
    return err
}
```

`Config.Theme` empty means no theme. `SetTheme` stores the sheet the next
`Redraw` applies and does not draw by itself; `SetTheme("")` removes the
theme. `Click` redraws after its handler, so a click can switch a theme
without an extra `Redraw`. A theme with unbalanced braces returns a parse
error from `New` or `SetTheme` and leaves the current theme in place; other
malformed input is skipped by the engine's CSS parser. The theme reaches both
the display-list path and the bitmap fallback, and `Page.PNG` paints with the
current theme. It applies to whichever template is current; `Load` does not
change it. Media queries in the theme are evaluated against the screen: width,
height, and orientation match; `@media print` and `prefers-color-scheme` never
do. A theme can set `background-image: url(name)`; `Page.SetImage` resolves
that name like a template source.

The [theme example](../examples/theme) opens a page, applies a light
stylesheet through `Config.Theme`, and swaps in a dark one on a click. Both
of its sheets keep the same geometry, so the switch does not move the page.
Keep geometry in the template and let a theme change paint: two themes that
set different padding, borders, or font sizes reflow the content when the
theme changes.

## Custom properties

The recommended shape is a template that declares defaults with
`var(--name, fallback)` and a theme sheet that sets `--name`. Because the
theme sheet is a plain Go string parsed by the engine, its values are never
touched by `html/template` escaping. That matters for data-driven CSS: a
value such as `rgb(26, 86, 219)`, `hsl(...)`, a gradient, or a shadow
interpolated through template data becomes `ZgotmplZ` unless the Go field is
typed `template.CSS`. Reaching for a custom property set by `SetTheme` avoids
that.

## How much CSS

The engine's 2026-09-21 catalog audit maps all 818 webref properties and marks
407 implemented, 0 partial, and 411 unsupported. Every implemented property can
be set from a theme. Unsupported properties are ignored; the permanent
non-goals (animation, transition, 3D transforms, scroll snap, and the
print-noop UI names such as `cursor` and `user-select`) never apply. Selector,
at-rule, function, and unit coverage is smaller than property coverage. The
engine's own `documentation/compatibility-matrix.md` is the detailed list.
