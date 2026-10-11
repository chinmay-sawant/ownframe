# Go desktop app frameworks compared with ownframe

Checked on 2026-10-11 against the projects' own documentation and repositories. This is a representative comparison, not a census of every Go GUI project.

## The main options

| Project | UI model | What renders the UI | Where it fits |
|---|---|---|---|
| [Wails v2](https://wails.io/docs/howdoesitwork/) | Go backend with an HTML, CSS, and JavaScript frontend. The frontend calls bound Go methods. | The operating system's web view: WebView2 on Windows, WebKit on macOS, and WebKitGTK on Linux. | The closest conventional choice for a Go app built like a web app. It provides the browser platform, including JavaScript, and expects frontend assets and tooling. |
| [Fyne](https://fyne.io/) | Go widgets, layouts, and canvas objects. You can write custom widgets. | Fyne's own canvas and widget renderer. It does not use an HTML browser runtime. | A cross-platform Go toolkit for standard desktop and mobile interfaces. It supports Windows, macOS, Linux, iOS, and Android. |
| [Gio](https://gioui.org/doc/architecture) | Go immediate-mode UI. The app handles events, updates its state, then lays out and emits drawing operations for each frame. | Gio's drawing operations and platform drivers. It does not use HTML or a browser runtime. | A Go UI library for desktop, mobile, and WebAssembly when you want to build the interface from Go code. |
| [webview/webview_go](https://github.com/webview/webview_go) | A Go binding to the `webview` library. It hosts HTML and supports a JavaScript-to-Go bridge. | The native system web view: WebView2 on Windows, WebKit on macOS, and WebKitGTK on Linux. | A smaller, lower-level route to a web-based desktop window. The binding exposes the webview API; you build more of the app shell yourself. |
| [Ebitengine](https://ebitengine.org/) | A Go 2D game engine with an update and draw loop, graphics, input, and audio APIs. | Ebitengine draws into its window or canvas. It is not an HTML engine or web view. | A rendering and input layer for custom Go interfaces, games, and tools. It targets desktop, mobile, and WebAssembly. |

The [webview core project](https://github.com/webview/webview) describes its purpose as a common HTML5 UI layer and documents the platform web-view backends. Wails builds a fuller Go application framework around a web frontend and Go bindings. Both use an actual browser engine supplied by the platform.

## How ownframe fits

Ownframe combines HTML templates and CSS with Go handlers, then opens the result in an Ebitengine window. Its current [module file](../go.mod) directly requires [Blinkless](https://github.com/chinmay-sawant/blinkless) and Ebitengine. Blinkless parses HTML, applies CSS, lays out the styled document, and emits a display list. Ownframe replays supported display operations through Ebitengine; it can use a bitmap fallback for operations its replay path cannot draw. See the [rendering path](screen.md) and [platform guide](platforms.md).

No embedded web view is involved. Ebitengine supplies the application loop, window, input, and drawing surface. Blinkless supplies HTML/CSS layout. The app does not get a live browser DOM or JavaScript runtime. The same page can also run as WebAssembly, through the mobile binding, or as an HTTP page that serves its rendered image.

Among the projects listed here, ownframe has a distinctive combination: a Go HTML/CSS layout engine, a separate Go drawing layer, and desktop, mobile, and WebAssembly targets without embedding a browser runtime. That describes this shortlist; it does not prove that no other Go project combines these parts.

## What the trade-off means

Ownframe is a good fit when you want to author an app screen in HTML/CSS, keep behavior in Go, control rendering, and avoid embedding a browser engine. The Teams and Pet Care demos show how that model handles application screens ([showcase](../showcase.md)). The fit is stronger when you control the templates and the CSS they use.

Wails is a better fit when you need JavaScript, frontend frameworks, browser APIs, or compatibility with existing web pages. JavaScript is a scope choice in ownframe; native apps can use JavaScript through a web view, as Wails does. Fyne or Gio is a better fit when you want a Go-defined interface with widgets or an immediate-mode programming model. Choose based on whether you need browser behavior, a Go-drawn interface, or Ownframe's HTML/CSS authoring with Go logic.

Ownframe's documented gaps still matter for production apps: it has no JavaScript or live DOM, no browser-grade compatibility guarantee, and no native screen-reader bridge. It exposes a page-level accessibility tree, but that tree does not give the Ebitengine window to assistive technology. Ownframe also lacks some common desktop shell features such as multiple windows, a system tray, and notifications. The [feature list](features.md) tracks current support and limits. A CSS-property count does not by itself show browser compatibility; the result depends on the exact engine revision, accepted values, selectors, and observed layout and painting.

## Note on the CSS engine and counts

The checkout used for this comparison depends on Blinkless in [go.mod](../go.mod). Its compatibility matrix, measured on 2026-10-10, lists 389 implemented properties, 1 partial, 387 unsupported, and 8 intentionally ignored in a catalog of 785. "Implemented" requires a handler, a consumer, and a behavior test. The matrix says this does not establish browser rendering parity.

The separate [gowkhtmltopdf compatibility matrix](https://github.com/chinmay-sawant/gowkhtmltopdf/blob/master/documentation/compatibility-matrix.md), audited on 2026-09-21, lists 407 implemented properties, 0 partial, and 411 unsupported out of 818. Its target is HTML-to-PDF, not an interactive app runtime. Ownframe's current `go.mod` does not depend on it. That matrix is historical context and does not measure Blinkless.

When quoting a CSS number, attach it to the engine, revision, and counting method. Property names, accepted values, selectors, HTML behavior, and browser rendering parity are different measures.

## Primary sources

- [Wails v2 architecture](https://wails.io/docs/howdoesitwork/) and [installation requirements](https://wails.io/docs/gettingstarted/installation/)
- [Fyne project](https://fyne.io/) and [widget architecture](https://docs.fyne.io/architecture/widgets/)
- [Gio project](https://gioui.org/) and [architecture](https://gioui.org/doc/architecture)
- [`webview` core README](https://github.com/webview/webview) and [Go binding README](https://github.com/webview/webview_go)
- [Ebitengine project](https://ebitengine.org/) and [repository](https://github.com/hajimehoshi/ebiten)
- [Blinkless repository](https://github.com/chinmay-sawant/blinkless) and [CSS compatibility matrix](https://github.com/chinmay-sawant/blinkless/blob/master/documentation/compatibility-matrix.md)
