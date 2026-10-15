# Clipboard

`clipboard.Write` stores the text in memory when `UseMemory(true)` has turned memory-only mode on, and tries the OS clipboard otherwise without touching the memory copy. With memory-only mode off, `clipboard.Read` returns the OS clipboard text, or `""` plus `ErrUnavailable` when that read fails; it never falls back to an older `Write`.

The window still calls `Write` and `Read` for the text chords. `Read` returns `(string, error)`; paste callers skip on error and tests fail on it. No Go module was added. The package does not start `wl-copy`, `xclip`, `pbcopy`, or PowerShell.

| Target | What `Write` and `Read` call |
|--------|------------------------------|
| Linux, `WAYLAND_DISPLAY` set | Nothing. A Wayland data-device client is not in this tree. The memory copy is used. |
| Linux, `DISPLAY` set and Wayland unset | The X11 `CLIPBOARD` selection on `/tmp/.X11-unix/X<n>`. A goroutine answers `SelectionRequest` while this process owns the selection. `UTF8_STRING` is the type. |
| Linux, neither variable set | The memory copy |
| Windows | `user32` `OpenClipboard` / `SetClipboardData` / `GetClipboardData` and `kernel32` global memory. The format is `CF_UNICODETEXT`. |
| macOS, cgo on | `NSPasteboard` `generalPasteboard`, type `NSPasteboardTypeString` |
| macOS, cgo off | The memory copy |
| Android, iOS, wasm | The memory copy |

On X11, `Write` may read `$XAUTHORITY`, or `~/.Xauthority` when that variable is unset, to connect to the X server. It does not create or change that file. If the socket or the handshake fails, the OS write reports through its own path and a later `Read` returns `ErrUnavailable` with no text; no stale text is ever returned.

Tests call `UseMemory(true)` before any `Write`, including `TestMain` in `internal/clipboard` and `examples/clipboard`. They do not touch the desktop clipboard.

The [clipboard example](../examples/clipboard) covers copy, cut, paste, select-all, undo, and redo from buttons; the Ctrl+C/X/V/A/Z/Y chords run the same page methods. The Copy and Cut buttons write the OS clipboard, and the Paste button reads it.

## Trust boundary

The clipboard is a local desktop resource. `Write` and `Read` talk only to
the in memory copy or the OS clipboard of the same process; `Serve` has no
clipboard route, and `/debug/state`, IPC, and the phone bindings never expose
clipboard text. Wayland stays on the memory copy because no Wayland
data device client is in this tree. Tests call `UseMemory(true)` and never
touch the desktop clipboard. `Write` may read `$XAUTHORITY` or
`~/.Xauthority` to reach an X server; it never creates or modifies that file.
