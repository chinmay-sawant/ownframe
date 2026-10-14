# Compare with Electron

Electron and OwnFrame both accept HTML and CSS for the interface.
Electron runs a browser-based interface with JavaScript. OwnFrame connects its own Go web view to Go app logic.
These examples display the same plain text, `Hello world`, with no added styles.

## The same screen

These are UI excerpts. Both omit imports, application initialization, window setup, and error handling.
Complete setup is linked separately for each framework.

### Electron

The body content in `index.html`:

```html
<div>Hello world</div>
```

The complete app creates a `BrowserWindow` and loads the HTML file with `loadFile`.
Electron's [first app tutorial](https://www.electronjs.org/docs/latest/tutorial/tutorial-first-app) includes installation and window setup.

### OwnFrame

The HTML field in a page configuration:

```go
ownframe.Config{
	HTML: `<div>Hello world</div>`,
}
```

A complete app also supplies a window size, creates the page with `New`, and opens it with `Run`.
See the [complete OwnFrame program](../README.md#get-started) and [window setup](window.md).

## What runs the screen

Electron uses Chromium for the interface and Node.js in its main process.
A `BrowserWindow` loads the page into a renderer process, where HTML, CSS, and JavaScript run.
The [Electron process model](https://www.electronjs.org/docs/latest/tutorial/process-model) explains those roles.

OwnFrame uses our own HTML and CSS view written in Go.
We built [Blinkless](https://github.com/chinmay-sawant/blinkless) from scratch as the shared engine across OwnFrame screens.
Blinkless parses HTML, applies CSS, and calculates layout.
OwnFrame handles Go data, actions, and input. Ebitengine opens the window and draws the result.
The [screen guide](screen.md) describes that flow.

## Where the app logic lives

| You write | Electron | OwnFrame |
|---|---|---|
| Screen | HTML | HTML templates |
| Styles | CSS | CSS |
| App logic | JavaScript, with Electron APIs | Go data and handlers |

OwnFrame's view supports a documented HTML and CSS subset, with no live DOM or JavaScript runtime.
Existing web apps that depend on browser APIs need changes to run in OwnFrame.
Read the [features and limits](features.md) for the supported behavior.
Desktop, WebAssembly, Android, and iOS setup is in the [platform guide](platforms.md).
