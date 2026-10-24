# Compare with Rust GPUI

GPUI defines screens in Rust. OwnFrame defines screens in HTML and CSS, with app logic in Go.
These examples display the same plain text, `Hello world`, with no added styles.

## The same screen

Both render the same plain text, `Hello world`, with no added styles.
The GPUI program below is complete and follows the official example, trimmed to unstyled text.
The OwnFrame program below is complete.

### Rust GPUI

```rust
use gpui::{
    div, prelude::*, px, size, App, Application, Bounds, Context,
    Window, WindowBounds, WindowOptions,
};

struct HelloWorld;

impl Render for HelloWorld {
    fn render(&mut self, _window: &mut Window, _cx: &mut Context<Self>) -> impl IntoElement {
        div()
            .flex()
            .size_full()
            .justify_center()
            .items_center()
            .child("Hello world")
    }
}

fn main() {
    Application::new().run(|cx: &mut App| {
        let bounds = Bounds::centered(None, size(px(500.), px(500.0)), cx);
        cx.open_window(
            WindowOptions {
                window_bounds: Some(WindowBounds::Windowed(bounds)),
                ..Default::default()
            },
            |_, cx| cx.new(|_| HelloWorld),
        )
        .unwrap();
    });
}
```

GPUI's [official example](https://gpui.rs/) shows `Render`, `div`, and `child` in a complete application.
Its [text element source](https://github.com/zed-industries/zed/blob/main/crates/gpui/src/elements/text.rs) accepts a static string as an element.

### OwnFrame

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
		HTML:   "<div>Hello world</div>",
		Width:  480,
		Height: 320,
	})
	if err != nil {
		log.Fatal(err)
	}
	if err := ownframe.Run(context.Background(), page); err != nil {
		log.Fatal(err)
	}
}
```

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
