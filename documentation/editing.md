# Editing

`Handlers` carries one function for each editing action: `Type`, `Backspace`, `DeleteWord`, `Submit`, `Copy`, `Cut`, `Paste`, `SelectAll`, `Undo`, and `Redo`. `Page` has a method with the same name, so a button or a key handler can run the action. The window maps Ctrl/Cmd C, X, V, A, Z, Shift+Z, and Y to these methods on its own; [keys.md](keys.md) has the full list. The edited control is a focused text-like input, including `file`, or a textarea; [forms.md](forms.md) covers focus, the control types, and the stored values.

The example buttons call the page methods:

```go
page.Handle(gpui.Handlers{
    Click: func(ctx context.Context, box gpui.Box) error {
        switch box.ID {
        case "undo":
            return page.Undo(ctx)
        case "redo":
            return page.Redo(ctx)
        }
        return nil
    },
})
```

## Typing

`Type`, `Backspace`, `DeleteWord`, and `Paste` share one contract. The handler runs first, then the built-in edit: `Type` and `Paste` insert the text, `Backspace` drops the last rune, and `DeleteWord` drops the last word and the spaces after it. A nil handler still edits the focused control. The page draws when the handler ran or the field changed; a nil handler with nothing to edit does nothing. An error from the handler aborts the edit and skips the redraw. An edit that changes the field fires `BeforeEdit` before the write and `Change` after it; [binding.md](binding.md) has both.

## Submit

`Submit` calls `Handlers.Submit` and draws. A nil handler is a no-op. Enter and NumpadEnter call it from the window; [keys.md](keys.md) covers the keys. An error from the handler is returned and skips the redraw.

## Copy and Cut

`Copy` and `Cut` return `(text, ok, err)`. A focused field returns its value and does not call the handler. `Copy` leaves the field alone and never draws. `Cut` clears the field, writes the bound field, fires `BeforeEdit` and `Change`, and draws; it fires both hooks even when the field is empty. With no focused field the handler runs: `Copy` still does not draw, and `Cut` draws only when the handler returns `ok` with no error. The hook contract is in [binding.md](binding.md).

## SelectAll

`SelectAll` selects the whole value of the focused control and draws. `FormSelected` reports it, and the selection clears on the next edit, click, or blur. A focused control does not call the handler. With no focused control the handler runs and the page draws; a nil handler is a no-op.

## Undo and Redo

`Undo` and `Redo` call `Handlers.Undo` and `Handlers.Redo`, then draw. A nil handler is a no-op. The library keeps no undo history, so the app owns the stack. The usual shape snapshots the fields in `BeforeEdit` and restores them with `SetFormValue`: [examples/login](../examples/login) does that, and [examples/editing](../examples/editing) and [examples/clipboard](../examples/clipboard) wire their buttons to the methods.
