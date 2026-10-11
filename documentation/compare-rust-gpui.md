# Compare with the Rust GPUI framework

Recorded on 2026-10-03, updated on 2026-10-06. The left column comes from [gpui.rs](https://www.gpui.rs/), the `gpui` crate page on crates.io, and the example on that site. The right column is what this repo does today. For the Electron gap list see [compare-electron.md](compare-electron.md).

Rust GPUI is a GPU-accelerated UI framework that builds a UI in Rust code. ownframe lays out HTML and paints it in a Go window. The projects are separate and use different approaches.

## Hello world

### Rust GPUI

`main.rs`

```rust
use gpui::{
    div, prelude::*, px, rgb, App, Application, Context, SharedString, Window, WindowOptions,
};

struct HelloWorld {
    text: SharedString,
}

impl Render for HelloWorld {
    fn render(&mut self, _window: &mut Window, _cx: &mut Context<Self>) -> impl IntoElement {
        div()
            .flex()
            .flex_col()
            .gap_3()
            .bg(rgb(0x505050))
            .size(px(500.0))
            .justify_center()
            .items_center()
            .text_xl()
            .text_color(rgb(0xffffff))
            .child(format!("Hello, {}!", &self.text))
    }
}

fn main() {
    Application::new().run(|cx: &mut App| {
        cx.open_window(WindowOptions::default(), |_, cx| {
            cx.new(|_| HelloWorld {
                text: "World".into(),
            })
        })
        .unwrap();
    });
}
```

### ownframe

`main.go`

```go
package main

import (
	"context"
	"log"

	"github.com/chinmay-sawant/ownframe"
)

func main() {
	page, err := ownframe.New(ownframe.Config{
		Title:  "Hello",
		HTML:   `<h1>{{.Title}}</h1>`,
		Width:  480,
		Height: 640,
	})
	if err != nil {
		log.Fatal(err)
	}
	page.SetData(struct{ Title string }{"Hello"})
	if err := ownframe.Run(context.Background(), page); err != nil {
		log.Fatal(err)
	}
}
```

The Rust side builds the whole UI in code and owns a render trait. This side hands over HTML and a `Run` call. That one difference runs through every row below.

## What differs

| Aspect | Rust GPUI | ownframe |
|---|---|---|
| What it is | a GPU-accelerated UI framework, hybrid immediate and retained mode | a Go library that lays out HTML and paints it in an Ebiten window |
| UI language | a Rust builder chain, `div().flex().bg(rgb(...)).child(...)` | an HTML template plus CSS |
| Data into the UI | struct fields read by `render`, updates through contexts | `page.SetData(any)`, then handlers run |
| Input | closures plus key-dispatch contexts | one `Handlers` struct: `Click`, `Submit`, `Type`, `Undo`, `Redo`, and the rest |
| Layout | its own layout engine, written for code-defined UI | Blinkless HTML and CSS layout, replayed through Ebiten ([screen.md](screen.md)) |
| Painting | GPU, custom shaders, retained and immediate | Ebiten canvas, display-list replay with a bitmap fallback ([screen.md](screen.md#replay)) |
| Targets | macOS, Linux, Windows | desktop, wasm in a browser, Android and iOS bind, plus a `-web` PNG page ([platforms.md](platforms.md), [web.md](web.md)) |
| Windows | multi-window, full native integration | one window per page |
| Already in it | animations, IME, deep text editing, multi-window | a page-level semantic tree, themes, IPC, HTML history, fetch, clipboard, crash reports, forms, a file open dialog; v0.0.2 adds DevTools, drag and drop, printing, hot reload, caret editing, and window options ([features.md](features.md)) |
| Not in it | HTML rendering, web and mobile targets | native screen-reader bridges, video and canvas, WebGL, multi-window |
| Maturity | runs Zed every day, published on crates.io as `gpui` 0.2.2 under Apache-2.0 | v0.0.2, an example for each feature |
| Build cost | a heavy Rust compile, the crate tracks Zed closely | a small Go build, two direct dependencies |

## What the Rust framework does better

- Performance. It draws on the GPU through its own renderer, with text shaping and effects it controls end to end.
- Maturity. Zed is the proof. Large team, years of work, thousands of lines of Rust plus Metal, HLSL, and WGSL shaders.
- Input depth. IME, dead keys, and text editing internals that a browser or a picture-based renderer does not get for free.
- Multi-window, with real native integration on every desktop platform.
- Animations and a retained mode for large, fast-updating screens.

## What this repo does better

- Familiarity of the UI layer. HTML and CSS, not a builder chain in a systems language.
- Language. Go for the handlers, `go test`, `go build`. No Rust toolchain and no borrow-checker-shaped API.
- Reach. Desktop, wasm, Android, and iOS from one page definition. That framework is desktop only.
- Size of the mental model. `New`, `SetData`, `Handle`, `Run`, `RunWithOptions`, `Serve`, `BindMobile`. The examples cover the rest.
- No render trait to implement. A page is a template string and a data value.

## When to pick which

- Build a Zed-like editor, or any fast custom UI in Rust: that framework.
- Show an HTML template in a window from Go, and ship one small binary: this one.

A Rust team that wants what this repo does ends up wrapping a webview crate, and a Go team that wants a Rust GPUI ends up rewriting the UI in a builder chain. Neither detour is worth it.

## Sources

- <https://www.gpui.rs/>
- <https://crates.io/crates/gpui>
- <https://github.com/zed-industries/zed/tree/main/crates/gpui>
