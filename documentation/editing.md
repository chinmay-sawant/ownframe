# Editing

`Handlers` carries one function for each editing action: `Type`, `Backspace`, `DeleteWord`, `Submit`, `Copy`, `Cut`, `Paste`, `SelectAll`, `Undo`, and `Redo`. The same names are calls on `page`, so a button or a key handler can run the action. The window maps Ctrl/Cmd C, X, V, A, Z, Shift+Z, and Y onto the inner screen. [keys.md](keys.md) has the full list. The edited control is a focused text-like input, including `file`, or a textarea. [forms.md](forms.md) covers focus, the control types, and the stored values.

The example buttons call `page`:

```go
page.Handle(ownframe.Handlers{
    Click: func(ctx context.Context, box ownframe.Box) error {
        switch box.ID {
        case "undo":
            return page.Edit().Undo(ctx)
        case "redo":
            return page.Edit().Redo(ctx)
        }
        return nil
    },
})
```

## Typing

`page.Edit().Type`, `Backspace`, `DeleteWord`, and `Paste` share one contract. The handler runs first, then the built-in edit: `Type` and `Paste` insert the text at the caret, `Backspace` removes the selected range or the rune before the caret, and `DeleteWord` removes the selected range or the word before the caret. A non-empty selection is replaced by the insert. A nil handler still edits the focused control. The page draws when the handler ran or the field changed; a nil handler with nothing to edit does nothing. An error from the handler aborts the edit and skips the redraw. An edit that changes the field fires `BeforeEdit` before the write and `Change` after it; [binding.md](binding.md) has both. Typing dirties the field box and the caret run, so only that part of the frame repaints. A field whose text grew or shrank dirties its old box and its new box, and a wrapped line repaints whole.

The caret is a rune offset in the focused value, and the anchor is where the selection started. `page.Edit().Selection` returns the focused field's caret offset and normalized `[start, end)` selection range. A click in the field sets the caret from the clicked glyph, `page.Edit().KeyDown` moves it on `arrowleft`, `arrowright`, `home`, and `end`, and a `ctrl+` or `alt+` prefix on left or right arrows jumps by word. In a `textarea`, `arrowup` and `arrowdown` move across both hard newlines and soft-wrapped visual lines while preserving the preferred horizontal column across short or empty lines until a horizontal move or edit resets it. Plain `home` and `end` in a `textarea` move to the start and end of the current visual line, and `ctrl+home` and `ctrl+end` jump to the start and end of the whole value. A `shift+` prefix keeps the anchor, so the range between the anchor and the caret is the selection. A caret move alone fires neither `BeforeEdit` nor `Change`, and a redraw keeps the caret unless the value shrank.

The caret blinks while a text field holds the focus. `page.Testing().Tick` turns it off after 530 ms of showing it, and on again 530 ms later. A caret move, an edit, a cut, or a select-all shows it and restarts the clock, so the line stays solid while someone types. A range selection paints no caret, so it does not blink. The window calls that tick on the inner screen. `Serve` and `page.Probe().PNG` do not tick, so a still picture keeps the caret. Blinking draws through `Redraw`, so the dirty region is the caret column only.

IME composition reaches fields through the window's `exp/textinput` session. Desktop draws preedit separately from the stored value; Android and iOS use the mobile field edit path until commit ([interaction.md](interaction.md)).

`page.Access().AccessibilityPerform` replaces a supported text control through the same
`BeforeEdit`, binding, and `Change` path. It does not expose password values
through the accessibility snapshot ([accessibility.md](accessibility.md)).

## Submit

`page.Edit().Submit` activates a focused button, or calls `Handlers.Submit` and draws. A nil handler is a no-op. Enter and NumpadEnter call it from the window for single-line inputs and buttons; in a focused `textarea`, plain Enter inserts a newline instead, and `Ctrl+Enter` or `Cmd+Enter` runs `Handlers.Submit`. [keys.md](keys.md) covers the keys. An error from the handler is returned and skips the redraw.

## Copy and Cut

`page.Edit().Copy` and `page.Edit().Cut` return `(text, ok, err)`. A focused field returns its selected range, or its whole value when the range is empty, and does not call the handler. `Copy` leaves the field alone and never draws. `Cut` removes that range, writes the bound field, fires `BeforeEdit` and `Change`, and draws; it fires both hooks even when the field is empty. A focused `type="password"` field refuses both: they return `("", false, nil)`, fire no hook, and leave the value alone, as browsers do. With no focused field the handler runs: `Copy` still does not draw, and `Cut` draws only when the handler returns `ok` with no error. The hook contract is in [binding.md](binding.md).

## SelectAll

`page.Edit().SelectAll` selects the whole range `[0, len)` of the focused control and draws. `FormSelected` reports it, and the selection clears on the next edit, click, or blur. A focused control does not call the handler. With no focused control the handler runs and the page draws; a nil handler is a no-op. A range that a drag, a double-click, or a triple-click selected reports `FormSelected` when it covers the whole value too.

## Undo and Redo

`page.Edit().Undo` and `page.Edit().Redo` call `Handlers.Undo` and `Handlers.Redo`, then draw. A nil handler is a no-op. The library keeps no undo history, so the app owns the stack. The usual shape snapshots the fields in `BeforeEdit` and restores them with `SetFormValue`: [examples/login](../examples/login) does that, and [examples/editing](../examples/editing) and [examples/clipboard](../examples/clipboard) wire their buttons to the methods.
