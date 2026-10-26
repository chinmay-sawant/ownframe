# Phase 05: search and command palette

Goal: reach a file or a line without the pointer.

Status: partly landed. Quick open, five default suggestions, path filtering,
the keyboard path, and in-file find are in `examples/code-view`. Fuzzy ranking,
content search, symbol search, filter history, and the full keyboard sweep are
open.

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

- [x] Add the palette view state and the overlay markup. Evidence:
      `palette.go`, `components/palette.html`, `components/palette.css`.
- [x] Show five default files and filter the open workspace by path.
      Evidence: `TestQuickOpenPalette`.
- [x] Focus the palette input and handle arrows, Enter, Escape, and a backdrop
      click. Evidence: `TestQuickOpenPalette`.
- [ ] Add fuzzy ranking with a total match count.
- [x] Add the in-file find bar with match offsets. Evidence:
      `TestFindInFile`, `find.go`.
- [x] Mark find matches on the visible rows. Evidence: `TestFindInFile`.
- [ ] Add symbol search over the active outline.
- [ ] Add filter history.
- [x] Restore the previous focus after Escape. Evidence: `find.go`,
      `closePalette`.
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
