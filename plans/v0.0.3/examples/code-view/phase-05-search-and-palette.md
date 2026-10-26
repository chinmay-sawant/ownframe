# Phase 05: search and command palette

Goal: reach a file or a line without the pointer.

Status: mostly landed. Quick open, five default suggestions, path filtering,
the keyboard path, in-file find, workspace content search, and the modal focus
trap are in `examples/code-view`. Fuzzy ranking, symbol search, filter history,
and a fuzzy total count are open.

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
- [x] Add workspace content search with `Ctrl+Shift+F`, a case toggle, a
      bounded result list, and jump-to-line. Evidence: `search_ops.go`,
      `workspace_search.go`, `TestFeatureWorkspaceSearch`,
      `TestWorkspaceSearchFindsAcrossFiles`.
- [ ] Add fuzzy ranking with a total match count.
- [x] Add the in-file find bar with match offsets. Evidence:
      `TestFindInFile`, `find.go`.
- [x] Mark find matches on the visible rows. Evidence: `TestFindInFile`.
- [ ] Add symbol search over the active outline.
- [ ] Add filter history.
- [x] Restore the previous focus after Escape. Evidence: `find.go`,
      `closePalette`.
- [x] Trap focus inside the palette, content search, and branch picker with
      `tabindex="-1"` on background controls. Evidence:
      `TestFeatureModalFocusTrap`, `TestAccessibilityPaletteDialog`.
- [x] Add tests for a keyboard-only open, a keyboard-only close, and a find
      match that scrolls. Evidence: `TestFindInFile`, `TestKeyboard`.

## Gate

A file opens from the palette with no pointer, and a find result lands on the
matching row.

## Risks

- The find marks must not break the 20 px row. Use a background on the existing
  span, not a new line.
- A fuzzy list over a large repository needs a cap and a result count.
- The palette is a modal. The page has no native dialog, so the focus trap and
  the Escape path need explicit tests.
