# Compare with Rust GPUI

GPUI defines screens in Rust. OwnFrame defines screens in HTML and CSS, with app logic in Go.
These examples display the same plain text, `Hello world`, with no added styles.

## The same screen

These are UI excerpts. Both omit imports, application initialization, window setup, and error handling.
Complete setup is linked separately for each framework.

### Rust GPUI

The body of a view's `render` method:

```rust
div().child("Hello world")
```

GPUI's [official example](https://gpui.rs/) shows `Render`, `div`, and `child` in a complete application.
Its [text element source](https://github.com/zed-industries/zed/blob/main/crates/gpui/src/elements/text.rs) accepts a static string as an element.

### OwnFrame

The HTML field in a page configuration:

```go
ownframe.Config{
	HTML: `<div>Hello world</div>`,
}
```

A complete app also supplies a window size, creates the page with `New`, and opens it with `Run`.
See the [complete OwnFrame program](../README.md#get-started) and [window setup](window.md).

## Where the code lives

| You write | Rust GPUI | OwnFrame |
|---|---|---|
| Screen | Rust elements | HTML templates |
| Styles | Rust builder methods | CSS |
| App logic | Rust state and event handlers | Go data and handlers |

## A Go HTML/CSS renderer plus an Ebitengine window

OwnFrame uses our own HTML and CSS view written in Go.
We built [Blinkless](https://github.com/chinmay-sawant/blinkless) from scratch as the shared engine across OwnFrame screens.
Blinkless parses HTML, applies CSS, and calculates layout.
OwnFrame handles data, actions, and input. Ebitengine opens the window and draws the result.
The [screen guide](screen.md) describes that flow.

This view supports a documented HTML and CSS subset, with no live DOM or JavaScript runtime.
Read the [features and limits](features.md) for the supported behavior.
Desktop, Android, WebAssembly, and `-web` setup is in the [platform guide](platforms.md). iOS is not a 1.0 target.
