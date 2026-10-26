# Phase 05: search and command palette

Goal: reach a file or a line without the pointer.

Status: planned.

## Design

1. `Ctrl+P` opens a modal palette over the editor. The input reuses the filter
   binding. The list shows a fuzzy match score, the path, and the language.
2. Recent files sit above the matches. Enter opens the first row. Escape
   closes the palette and returns focus to the previous control.
3. `Ctrl+F` opens a find bar inside the editor. Matches are marked on the
   visible rows, the count shows `n / total`, and Enter and Shift+Enter move
   between matches.
4. Symbol search reuses the outline list. It matches the symbol name and jumps
   to the line.
5. The filter history keeps the last five queries and clears with Escape.
6. Every palette action has a key path. A reviewer can open, choose, and close
   without a pointer.

## Checklist

- [ ] Add the palette view state and the modal markup.
- [ ] Build a fuzzy matcher over `ws.Files` with a bounded result list.
- [ ] Focus the palette input and trap Escape and Enter.
- [ ] Add the in-file find bar with match offsets.
- [ ] Mark find matches on the visible rows.
- [ ] Add symbol search over the active outline.
- [ ] Add filter history.
- [ ] Restore focus after Escape.
- [ ] Add tests for a keyboard-only open, a keyboard-only close, and a find
      match that scrolls.

## Gate

A file opens from the palette with no pointer, and a find result lands on the
matching row.

## Risks

- The find marks must not break the 20 px row. Use a background on the existing
  span, not a new line.
- A fuzzy list over a large repository needs a cap and a result count.
- The palette is a modal. The page has no native dialog, so the focus trap and
  the Escape path need explicit tests.
