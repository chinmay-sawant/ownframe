# Compare hello world syntax across Go UI libraries

Recorded on 2026-10-06. Every snippet below is the smallest app the library puts in its own README, quoted from that README and kept close to the source. Electron and the Rust GPUI framework already have pages of their own, [compare-electron.md](compare-electron.md) and [compare-rust-gpui.md](compare-rust-gpui.md); this file only looks at the syntax of the hello world.

Four libraries, three answers to one question. Where does the UI live?

- In markup, with Go behind it: ownframe, and the MyGo web window.
- In Go code, checked at build time: go-gui, gogpu/ui, and the MyGo native window.
- MyGo puts both answers in one framework.

## ownframe

```go
page, err := ownframe.New(ownframe.Config{
    Title:  "Hello",
    HTML:   `<h1>{{.Title}}</h1>`,
    Width:  480,
    Height: 640,
})
page.SetData(struct{ Title string }{"Hello"})
ownframe.Run(context.Background(), page)
```

The UI is one HTML string and the CSS beside it. Go fills the template through `SetData` and one `Handlers` struct answers clicks and keys, with a `data-action` attribute routing a click to a Go function. Nothing crosses a bridge.

The supported CSS set is published instead of implied. The current Blinkless matrix lists 389 implemented properties, 1 partial, and 387 unsupported in a catalog of 785. That count describes property handlers and behavior tests; it does not establish browser rendering parity. Check the [Blinkless compatibility matrix](https://github.com/chinmay-sawant/blinkless/blob/master/documentation/compatibility-matrix.md) for the property, value, selector, and unit status.

## go-gui

The `get_started` example, one button and one counter, from the go-gui README:

```go
type App struct {
    Clicks int
}

func main() {
    w := gui.SimpleWindow("Get Started", 300, 300, &App{}, func(w *gui.Window) {
        w.SetView(mainView)
    })
    backend.Run(w)
}
```

```go
func mainView(w *gui.Window) gui.View {
    app := gui.State[App](w)

    return gui.Column(gui.ContainerCfg{
        Sizing: gui.FillFill,
        HAlign: gui.HAlignCenter,
        VAlign: gui.VAlignMiddle,
        Content: []gui.View{
            gui.Label("Hello GUI!", gui.CurrentTheme().TextStyleDisplay),
            gui.Label(fmt.Sprintf("%d Clicks", app.Clicks), gui.TextStyle{}),
            gui.TextButton("Click Me", func(ctx gui.EventCtx) {
                gui.State[App](ctx.Window).Clicks++
            }),
        },
    })
}
```

The view runs again for each frame and rebuilds the widget slice from the state. State is one typed slot per window, `gui.State[App](w)`, and the click is a closure that writes the slot. Style comes from the theme, `gui.CurrentTheme().TextStyleDisplay`, with `Cfg` structs such as `gui.ContainerCfg` for everything the theme does not carry.

## gogpu/ui

```go
gogpuApp := gogpu.NewApp(gogpu.DefaultConfig().
    WithTitle("My App").
    WithSize(800, 600))

uiApp := app.New(
    app.WithWindowProvider(gogpuApp),
    app.WithPlatformProvider(gogpuApp),
    app.WithEventSource(gogpuApp.EventSource()),
)

uiApp.SetRoot(
    primitives.Box(
        primitives.Text("Hello gogpu/ui!").
            FontSize(24).Bold().
            Color(widget.RGBA8(33, 33, 33, 255)),
        primitives.Text("Enterprise-grade GUI for Go").
            FontSize(16).
            Color(widget.RGBA8(100, 100, 100, 255)),
    ).Padding(24).Gap(12).Background(widget.RGBA8(255, 255, 255, 255)),
)

if err := desktop.Run(gogpuApp, uiApp); err != nil {
    log.Fatal(err)
}
```

The tree is built once and handed to `SetRoot`. Every style is a chained method. State lives in `state.Signal` values, and a signal write invalidates the widgets bound to it, so the window is patched instead of rebuilt.

## MyGo, web window

```go
type Greeter struct{}

func (Greeter) Greet(name string) string { return "Hello, " + name }

func main() {
    mygo.Bind(Greeter{})
    mygo.App.WhenReady(func() {
        mygo.NewWindow(mygo.WindowOptions{Title: "Hello", URL: "/"})
    })
    if err := mygo.App.Run(); err != nil {
        log.Fatal(err)
    }
}
```

```ts
import { Greeter } from "./mygo"; // generated

document.body.textContent = await Greeter.greet("Ada");
```

The page is a real web page in the webview the OS already has, WKWebView, WebKitGTK, or WebView2. The TypeScript client is generated from the bound Go service, so the bridge is typed on both sides, and the UI syntax is whatever the web toolchain already is.

## MyGo, native window

```go
type counter struct{ n int }

func (s *counter) view(c *ui.Context) {
    ui.Column(c).Fill().Center().Gap(12).Children(func() {
        ui.Text(c, fmt.Sprint(s.n)).FontSize(40).Bold()
        if ui.PrimaryButton(c, "Increment").Clicked() {
            s.n++
        }
    })
}
```

No markup at all. The view reads the receiver, and the click handling sits in the if that draws the button: `if ui.PrimaryButton(c, "Increment").Clicked()`. The native window is drawn on the GPU, so no webview opens for it.

## The syntax side by side

| Aspect | ownframe | go-gui | gogpu/ui | MyGo web | MyGo native |
|---|---|---|---|---|---|
| UI lives in | an HTML string or file, CSS beside it | a Go view function, rebuilt each frame | a Go widget tree built once | an HTML, CSS, and JS page | a Go view function |
| Styling | CSS, then `Config.Theme` layered after | theme styles and `Cfg` structs | chained methods, `FontSize(24).Bold()` | CSS from the web toolchain | chained methods, `.FontSize(40).Bold()` |
| State | `page.SetData`, read by handlers | `gui.State[App](w)`, one typed slot per window | `state.Signal` values bound to widgets | Go services behind the generated client | fields on the view receiver |
| A click | `data-action` in the markup, one Go handler | a closure per button | a closure per widget, or a signal write | `await Greeter.greet("Ada")` | `Clicked()` read in the same if |
| Layout | Blinkless, in process | the framework's Fit, Fixed, and Fill pass | flex, grid, and stack engines | the OS webview | flexbox and grid, in Go |
| Painting | Ebiten canvas, display list replay | Metal, OpenGL, WebGL | the gogpu GPU stack | the OS webview | the MyGo GPU stack |
| A typo in the UI | runtime, a selector that matches nothing is silent | build error | build error | build error on the bridge, runtime in the page | build error |

## What the syntax says

Markup or code is the first split. ownframe and the MyGo web window keep HTML and CSS, so a stylesheet is a file a non-Go hand can edit, and the price is that a selector that matches nothing fails quietly. The other three keep the UI in Go and fail the build instead.

How state reaches the screen is the second. go-gui rebuilds the view every frame from a typed slot. gogpu/ui builds once and patches through signals. The MyGo native view reads its receiver. ownframe replays a cached document and lays it out again when the data, size, or state changes, with no reparse and no DOM.

The bridge count is the third. Only the MyGo web window crosses a language boundary, and the generated TypeScript client is what keeps that bridge typed. go-gui and gogpu/ui never leave Go. ownframe keeps the page script-free, so there is no boundary to cross.

## Sources

- go-gui: <https://github.com/go-gui-org/go-gui>, example at <https://github.com/go-gui-org/go-gui/blob/main/examples/get_started/main.go>
- gogpu/ui: <https://github.com/gogpu/ui>
- MyGo: <https://github.com/egoist/mygo>
- ownframe: [../README.md](../README.md), [screen.md](screen.md)
- Engine CSS catalog: <https://github.com/chinmay-sawant/blinkless/blob/master/documentation/compatibility-matrix.md>
- All snippets recorded on 2026-10-06.
