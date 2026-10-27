# Keys

`Handlers.KeyDown` and `Handlers.KeyUp` run when a key goes down and up in
the window. `key` is the Ebiten key name lowercased with the `Digit`
prefix dropped: `"space"`, `"arrowup"`, `"arrowdown"`, `"escape"`, `"a"`,
or `"1"`. `Digit1` arrives as `"1"`. Modifier keys arrive the same way:
`"shift"`, `"control"`, `"alt"`, `"meta"`, plus `"shiftleft"` and the
other left and right names.

The window prefixes the four caret keys with the modifiers held:
`arrowleft`, `arrowright`, `home`, and `end` arrive as
`shift+arrowleft`, `ctrl+home`, or `alt+end`. Command uses the `ctrl+`
prefix. Every other key keeps its plain name; [editing.md](editing.md)
reads the prefixes.

```go
page.Handle(ownframe.Handlers{
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
control.

`Ctrl+=`, `Ctrl++`, `Ctrl+-`, and `Ctrl+0` change the page zoom on desktop,
and Command works in place of Ctrl. The window consumes them, so the page's
key handler never sees the presses. One press moves the zoom by 10 percent,
and `Ctrl+0` returns it to 100 percent. While Ctrl is down the wheel zooms
the page too, and does not scroll it. `Config.DisableWheelZoom` turns that
zoom off, and the wheel scrolls. `Config.DisableKeyZoom` turns the keyboard
chords off. The window still consumes them, so the page does not scale.
The zoom stays between 0.25 and 4,
the same range a pinch uses, and it scales the painted page without
changing the layout, so a click still lands on the box under the pointer.

Tab and Shift+Tab move focus when the screen has a focusable control; an
id-bearing button counts. The window calls `FocusNext` or `FocusPrev`, the page
moves the ring, and the handler does not see the key. A page without a
focusable control keeps the key, so a game can still read Tab. Escape closes
the context menu, or clears focus when the menu is closed; a press that closed
the menu is consumed, while its release and both edges of an Escape with no
menu open reach the handler. F11 toggles fullscreen on desktop
and is consumed there; wasm and mobile keep their own fullscreen. Arrow
keys, Home, and End still reach the handler, and the focused field moves
its caret for them. [interaction.md](interaction.md) has the full
picture.

F12 and Ctrl+Shift+I belong to the window while the screen has an inspector
([devtools.md](devtools.md)). The page's key handler never sees either press
or its release. While the overlay is on, `o` toggles the operation outlines
and selects the Ops tab, and does not type. The overlay's keys use the same
two-frame guard as every other key.

Backspace deletes on the frame of the press, then every 4 frames once the
key has been held 30 frames. `Ctrl+Backspace` (`page.Edit().DeleteWord`) deletes on the
same cadence. After a chord fires, its key stops typing while it stays down,
and it types again once the key has been up for 50 ms.

`Serve` has no Ebiten window, so its browser shell forwards keydown and keyup
to `/key`, printable characters to `/type`, and Tab to `host.Focuser`. The
desktop window, the phone build, and the WebAssembly canvas send the same key
names directly.
