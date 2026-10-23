# Drag and drop

The window reads `ebiten.DroppedFiles()` once per frame, after the pointer
pass. A nil or empty file system means no drop. A non-empty one becomes a
slice of `host.Drop` values, and the window calls `Drop` on a screen that
implements `host.Dropper`. A screen without the method ignores the files.

On desktop, `Drop.Path` is an absolute local path. Copy bytes from `Drop.Read`
inside the callback. Do not log the path. On wasm the path is empty. `Serve`
has no drop.

```go
type Drop struct {
	Name  string
	Size  int64
	IsDir bool
	Path  string
	Read  func() ([]byte, error)
}

type Dropper interface {
	Drop(ctx context.Context, files []Drop) error
}
```

`Path` is absolute on desktop and empty in a browser. `Read` returns the
bytes. A dropped directory arrives as one entry with `IsDir` true, and the
window never walks the tree. `Read` on a directory returns the file system's
error.

## The page handler

`page.Handlers.Drop` takes the same slice, and `ownframe.Drop` is the root name
for the type. `Page.Edit().Drop` calls the handler and draws the page again. An
error returns before the redraw.

```go
page.Handle(ownframe.Handlers{
	Drop: func(ctx context.Context, files []ownframe.Drop) error {
		for _, file := range files {
			if file.IsDir {
				continue
			}
			data, err := file.Read()
			if err != nil {
				return err
			}
			page.SetImage(file.Name, data)
		}
		return nil
	},
})
```

## Frame scope

`ebiten.DroppedFiles` is scoped to the `Update` frame that reads it. The
window builds the values and calls `Drop` inside that frame, so a handler
has one frame to act:

- A handler that keeps `Path` can open the file later, on desktop.
- A handler that reads the bytes and stores them keeps them.
- A handler that returns without reading gets nothing, and a `Read` saved
  for later is not valid after the call.

The API cannot hide this. Read or copy what you need during the call.

## The drag itself

Ebiten reports the files when the drag is released over the window, through
GLFW's drop callback. There is no drag-over event, so the window cannot show
a highlight while a file hovers over it. Nothing arrives until the release
lands on the window, and then the handler runs once. On Linux the drop comes
through X11's Xdnd when the window runs on X11; a drag that starts in a
Wayland-native file manager may not reach an XWayland window if the
compositor does not bridge the two. Under WSLg the drag starts in Windows
Explorer and never becomes an Xdnd event, so no drop reaches the window.

## Web and wasm

`-web` runs the page through `internal/web`, not the Ebiten loop, so
`ebiten.DroppedFiles` is never read and no route carries a drop. The wasm
canvas build runs the same `Update` as the desktop window, so a browser drop
arrives there. The browser is the manual test on WSLg, where a drag from
Windows cannot become an Xdnd drop: `sh scripts/browser.sh drop`, open the
address in a browser on Windows, and drop a file on the canvas.

[examples/drop](../examples/drop) prints one line for every dropped file,
whatever its type: the absolute path on desktop, the entry name in a browser.
The whole window takes a drop, so there is no target to aim at. Its file
control prints the path the desktop picker returns when a drag cannot reach
the window, and the picker asks the Windows dialog under WSL. Its tests drop
names, paths, and a directory, and drive a fake picker through the file
control.

## Trust boundary

`Drop.Path` is a local absolute path on desktop and empty in a browser. Treat
it as sensitive desktop data: do not log it, display it to end users, store it
in crash reports, or expose it through `Serve`, `/debug/state`, IPC, or any
network route. Prefer copying the bytes from `Drop.Read` inside the handler,
because `Read` is only valid for the duration of the `Drop` call; a path saved
for later may no longer point at a reachable file after the frame ends. Dropped
directories arrive as one entry and `Read` returns the file system error.

See [security.md](security.md) for the broader trust boundaries.
