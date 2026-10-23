# Clipboard

On GNOME/Wayland, paste prefers `wl-copy`/`wl-paste`. If those are missing, it uses the X11 `CLIPBOARD` selection when `DISPLAY` is set.

`clipboard.Write` stores the text in memory when `UseMemory(true)` has turned memory-only mode on, and tries the OS clipboard otherwise without touching the memory copy. With memory-only mode off, `clipboard.Read` returns the OS clipboard text, or `""` plus `ErrUnavailable` when that read fails; it never falls back to an older `Write`.

The window still calls `Write` and `Read` for the text chords. `Read` returns `(string, error)`; paste callers skip on error and tests fail on it. No Go module was added. On Wayland, `Write` and `Read` run `wl-copy` and `wl-paste` when those binaries are on PATH. If they are missing and `DISPLAY` is set, Linux uses the X11 `CLIPBOARD` path instead. The package does not start `xclip`, `pbcopy`, or PowerShell.

| Target | Files | What `Write` and `Read` call |
|--------|-------|------------------------------|
| Linux, `WAYLAND_DISPLAY` set | `clipboard_linux.go`, `clipboard_wayland.go` | `wl-copy --type text/plain --` with the text on stdin, and `wl-paste --no-newline`. `Clear` runs `wl-copy --clear`. If those tools are missing and `DISPLAY` is set, the X11 `CLIPBOARD` path. If neither helpers nor `DISPLAY`, a write is a no-op and a read returns `ErrUnavailable`. |
| Linux, `DISPLAY` set and Wayland unset | `clipboard_linux.go`, `x11_*.go` | The X11 `CLIPBOARD` selection on `/tmp/.X11-unix/X<n>`. A goroutine answers `SelectionRequest` while this process owns the selection. `UTF8_STRING` is the type. |
| Linux, neither variable set | `clipboard_linux.go` | A write is a no-op and a read returns `ErrUnavailable`. Not the in-process memory copy. |
| Windows | `clipboard_windows.go`, `clipboard_windows_read.go` | `user32` `OpenClipboard` / `SetClipboardData` / `GetClipboardData` and `kernel32` global memory. The format is `CF_UNICODETEXT`. |
| macOS, cgo on | `clipboard_darwin.go` (`darwin && !ios && cgo`) | `NSPasteboard` `generalPasteboard`, type `NSPasteboardTypeString` |
| macOS, cgo off | `clipboard_darwin_nocgo.go` | The memory copy |
| wasm | `clipboard_other.go` | The memory copy |
| Android | `clipboard_other.go` (`android` matches this stub) | The memory copy |

Paths are under `internal/clipboard/`. iOS would also hit `clipboard_other.go`; iOS is not a 1.0 target.

On X11, `Write` may read `$XAUTHORITY`, or `~/.Xauthority` when that variable is unset, to connect to the X server. It does not create or change that file. If the socket or the handshake fails, the OS write reports through its own path and a later `Read` returns `ErrUnavailable` with no text; no stale text is ever returned.

Tests call `UseMemory(true)` before any `Write`, including `TestMain` in `internal/clipboard` and `examples/clipboard`. They do not touch the desktop clipboard.

The [clipboard example](../examples/clipboard) covers copy, cut, paste, select-all, undo, and redo from buttons. Those buttons call `page`. The Ctrl/Cmd chords run the same work on the inner screen. The Copy and Cut buttons write the OS clipboard, and the Paste button reads it.

## Trust boundary

The clipboard is a local desktop resource. `Write` and `Read` talk only to
the in memory copy or the OS clipboard of the same process; `Serve` has no
clipboard route, and `/debug/state`, IPC, and the phone bindings never expose
clipboard text. Wayland uses `wl-copy` and `wl-paste` when they are on
PATH. If they are missing, Linux uses X11 `CLIPBOARD` when `DISPLAY`
is set, and otherwise leaves the OS clipboard unused. Clipboard text is
never logged. Tests call `UseMemory(true)` and never touch the desktop
clipboard. `Write` may read `$XAUTHORITY` or `~/.Xauthority` to reach
an X server; it never creates or modifies that file.
