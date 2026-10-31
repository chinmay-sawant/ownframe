# Phase 08: accessibility and mobile

Goal: a keyboard and screen-reader path through the viewer, and a phone layout
that reaches the same files.

Status: partly landed. Accessible names, toggle states, dialog roles, the
background focus trap, focus-visible rings, and reduced motion are in. The
Linux AT-SPI live bridge in `internal/atspi` registers the tree with the
accessibility bus, serves the read side, and emits focus and structure events.
Live Orca and Accerciser validation is tracked in issue #48. Mobile and Android
remain out of scope for this pass.

## Design

1. Every icon-only control gets an accessible name. The tree rows announce the
   file name, the language, and the change letter.
2. The tab order follows the visual order: title bar, activity bar, explorer,
   tabs, breadcrumbs, editor, details, status bar.
3. A visible focus ring marks the active control. The ring uses a border, not a
   shadow.
4. The token colours clear 4.5:1 against the editor. The diff backgrounds keep
   the added and deleted lines distinct without colour alone: the sign column
   carries a `+` or `-`.
5. The drawer, toast, and welcome card honour `prefers-reduced-motion`.
6. The phone layout moves the explorer into a sheet and keeps the editor full
   width. The details drawer becomes a full-height sheet.
7. Android host work: insets, the back key, and the text input session.

## Checklist

- [x] Add accessible names to every icon-only control. Evidence:
      `TestAccessibilityLabels`, `accessibility_labels_test.go`.
- [x] Announce the tree row text and the change letter.
- [x] Fix the focus order and add a visible ring. Evidence:
      `TestFeatureModalFocusTrap`, `components/base.css`.
- [ ] Check token and diff contrast. The sign column already carries `+`/`-`.
- [x] Add the reduced-motion path. Evidence: `components/base.css`.
- [x] Add live AT-SPI screen-reader bridging. Evidence: `internal/atspi`
      serves the tree over the accessibility bus; `documentation/atspi.md`
      records the live validation. Orca and Accerciser sign-off is issue #48.
- [ ] Add the phone explorer sheet.
- [ ] Add the Android insets and back key.
- [x] Add a keyboard-only walk test. Evidence: `TestKeyboard`,
      `TestFeatureModalFocusTrap`.
- [x] Add a focus-order test. Evidence: `TestAccessibilityLabels`,
      `TestAccessibilityPaletteDialog`.

## Gate

The page is usable with no pointer, and a phone build reaches the file list and
the diff.

## Risks

- The engine has a partial accessibility tree. The Linux AT-SPI client is
  live for reading and focus events; AT-initiated actions are not wired yet.
- A phone sheet changes the editor height. The windowing callback must rebuild
  for the new height.
- Colour-only diff marks fail a colour-blind check. Keep the sign column.
