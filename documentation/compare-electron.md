# Compare with Electron

Recorded on 2026-10-03, updated on 2026-10-06. The Electron column comes from the current Electron tutorial, the process model guide, and the API reference. The last column is what this repo does today. The v0.0.1 gap scan is [../plans/v0.0.1/compare.md](../plans/v0.0.1/compare.md); this file is the same gap list next to the framework it was written against.

## Hello world

### Electron

Create these files in the example directory.

`package.json`

```json
{
  "main": "main.js",
  "scripts": {
    "start": "electron ."
  }
}
```

`main.js`

```javascript
const { app, BrowserWindow } = require('electron/main')

const createWindow = () => {
  const win = new BrowserWindow({ width: 800, height: 600 })
  win.loadFile('index.html')
}

app.whenReady().then(() => {
  createWindow()
  app.on('activate', () => {
    if (BrowserWindow.getAllWindows().length === 0) createWindow()
  })
})

app.on('window-all-closed', () => {
  if (process.platform !== 'darwin') app.quit()
})
```

`index.html`

```html
<h1>Hello from Electron renderer!</h1>
```

Run the app:

```sh
npm install electron --save-dev
npm start
```

### ownframe

Create a Go module and add the library:

```sh
go mod init hello
go get github.com/chinmay-sawant/ownframe
```

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

Run the app:

```sh
go run .
```

The extra Electron lines buy a process model, a main process plus one renderer per window, and the window-lifecycle rules for macOS and Windows. This repo has none of that machinery, so none of it shows up in the program.

## What differs

| Aspect | Electron | ownframe |
|---|---|---|
| What it is | Chromium plus a Node.js runtime with JS APIs | a Go library that lays out HTML and paints it in an Ebiten window |
| UI language | HTML, CSS, JavaScript | HTML, CSS, Go |
| App structure | main process plus one renderer process per window, IPC between them | one process, in-process callbacks |
| Data into the UI | DOM APIs, or IPC over a preload bridge | `page.SetData(any)` |
| Input | any DOM event with JS listeners | one `Handlers` struct: `Click`, `Submit`, `Type`, `Undo`, `Redo`, and the rest |
| Layout | Blink, the Chromium engine | Blinkless HTML and CSS layout with a Go display-list replay path ([screen.md](screen.md)) |
| Painting | Chromium GPU compositor, the whole web platform | Ebiten canvas, display-list replay with a bitmap fallback ([screen.md](screen.md#replay)) |
| Targets | Windows, macOS, Linux | desktop, wasm in a browser, Android and iOS bind, plus a `-web` PNG page ([platforms.md](platforms.md), [web.md](web.md)) |
| Windows | multi-window, frameless, tray, menus, notifications | one window per page |
| Already in it | DevTools, IME, accessibility, video, canvas, WebGL, printing, drag and drop, auto-update, packaging | a page-level semantic tree, themes, IPC, HTML history, fetch, clipboard, crash reports, forms, a file open dialog; v0.0.2 adds DevTools, drag and drop, printing, hot reload, caret editing, and window options ([features.md](features.md)) |
| Not in it | mobile targets, small downloads, a no-JS mode | Chromium fidelity, native screen-reader bridges, video and canvas, WebGL, multi-window, auto-update, an installer |
| Footprint | Chromium and Node bundled, installers run tens to over a hundred MB | one Go binary, a fraction of that |
| Maturity | 10+ years, runs VS Code, Slack, Discord, WhatsApp | v0.0.2, an example for each feature |
| Building and shipping | npm, Forge or electron-builder, code signing, an update server | `go build`, or `scripts/package.sh` for a release archive |

## What Electron does better

- Fidelity. It is Chromium. Video, canvas, WebGL, IME, accessibility, printing, fonts, and every CSS feature work out of the box.
- Tooling. DevTools, hot reload, the npm ecosystem, electron-builder, auto-update, crash reporting, Playwright.
- Maturity. A decade of docs, a large community, and desktop apps used by millions. Most problems already have an answer on Stack Overflow.
- The native shell. Multiple windows, menus, tray icons, notifications, frameless and kiosk modes.
- The security model. The renderer sandbox and context isolation exist, once configured.
- Hiring. JavaScript developers are everywhere.

## What this repo does better

- Size and startup. No Chromium, no Node, no `node_modules`. One Go program, one binary, one process.
- One language. Go handlers, `go test`, `go build`. No JS, no preload script, no `contextBridge`, and no IPC serialization. A click handler is a direct function call.
- Reach. The same screen runs on desktop, as wasm, on Android, and on iOS. Electron is desktop only.
- No Chromium treadmill. Electron ships Chromium security patches on a schedule and callers have to keep up. This repo has a small attack surface and no browser to patch.
- A small API. `New`, `SetData`, `Handle`, `Run`, `RunWithOptions`, `Serve`, `BindMobile` ([window.md](window.md), [platforms.md](platforms.md)).
- A PNG endpoint for free. `-web` serves the same screen as a picture ([web.md](web.md)).

## What is still missing here

- No native screen-reader bridge or spellcheck. The page exposes a semantic tree, but it does not connect the Ebiten window to assistive technology ([accessibility.md](accessibility.md)). Desktop IME composition has a window path, with native input-method checks still open. Tab traversal and the caret landed in v0.0.2 ([interaction.md](interaction.md)).
- No multi-window, no tray, no native menus, no system notifications. Window options cover borderless, floating, fixed-size, and click-through windows ([window.md](window.md)).
- No video, canvas, or WebGL. Audio plays through Ebiten in the examples, not through an HTML element.
- No auto-update, no installer, no code signing. DevTools, drag and drop, printing, hot reload, caret editing, window resizing, and the packaging script landed in v0.0.2 ([devtools.md](devtools.md), [drag-drop.md](drag-drop.md), [printing.md](printing.md), [hot-reload.md](hot-reload.md), [interaction.md](interaction.md), [packaging.md](packaging.md)).
- IPC is in-process only. There is no cross-process bridge and no JavaScript.
- Navigation is `Load`, `Back`, `Forward`, and `data-action` routes. Clicked `http`, `https`, and `mailto` links can call an app callback. The host can dispatch allowlisted deep links, but the app owns OS protocol registration and delivery. Ownframe does not load a URL as a document.
- The display-list path uses the pinned Blinkless module. It does not open an OS web view or expose a browser DOM.

## Sources

- <https://www.electronjs.org/docs/latest/tutorial/tutorial-first-app>
- <https://www.electronjs.org/docs/latest/tutorial/process-model>
