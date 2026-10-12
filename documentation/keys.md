# Keys

`Handlers.KeyDown` and `Handlers.KeyUp` run when a key goes down and up in
the window. `key` is the Ebiten key name lowercased with the `Digit`
prefix dropped: `"space"`, `"arrowup"`, `"arrowdown"`, `"escape"`, `"a"`,
or `"1"`. `Digit1` arrives as `"1"`. Modifier keys arrive the same way:
`"shift"`, `"control"`, `"alt"`, `"meta"`, plus `"shiftleft"` and the
other left and right names.

```go
page.Handle(gpui.Handlers{
    KeyDown: func(ctx context.Context, key string) error {
        if key == "space" {
            jump()
        }

        return nil
    },
    KeyUp: func(ctx context.Context, key string) error {
        if key == "space" {
            stopHolding()
        }

        return nil
    },
})
```

Neither handler draws. A game changes its state and paints from
`Page.SetTick`; a page that must change its HTML calls `Page.Redraw` from
the handler. An error from either handler stops the window, and `Run`
returns it. The window drops auto-repeat pulses, so a held key arrives as
one press and one release, not a stream of pairs. The guard is two frames
long: a release counts only after the key has been down two frames, and
the next press only after it has been up two frames. A tap shorter than
two frames loses its release.

The window still handles `Ctrl+C`, `Ctrl+X`, `Ctrl+V`, `Ctrl+A`, `Ctrl+Z`,
`Ctrl+Shift+Z`, `Ctrl+Y`, `Ctrl+Insert`, `Shift+Insert`, `Shift+Delete`,
`Ctrl+Backspace`, Enter, and NumpadEnter on its own. Command works in place
of Ctrl. `Ctrl+Alt` without Meta is AltGr; it fires no chord and types
normally. Enter and NumpadEnter call `Handlers.Submit`. A key handler sees
those keys too, and typing in a focused form control still edits the
control. Tab, Escape, and the arrow keys reach the handler like any other
key; the window does not move focus or scroll from them.

Backspace deletes on the frame of the press, then every 4 frames once the
key has been held 30 frames. `Ctrl+Backspace` (`DeleteWord`) deletes on the
same cadence. After a chord fires, its key stops typing while it stays down,
and it types again once the key has been up for 50 ms.

`Serve` has no window, so its page receives no key events. The desktop
window, the phone build, and the WebAssembly canvas all send them.
