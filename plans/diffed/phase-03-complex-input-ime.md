# Phase 3: Complex Input & Native Composing IME

> **Parent:** [README.md](README.md)  
> **Area:** Input Subsystems, Text Fields & International Input Methods (IME)  
> **Status:** Open `[ ]`  

---

## Overview

Text input currently handles direct character typing, backspace, selection ranges, and clipboard operations. However, full international text input requires a multi-stage **Input Method Editor (IME)** pipeline:
- **Composing State (Preedit):** Displaying uncommitted phonetic text directly in the active input field with visual feedback (e.g. dotted underline for active composition, inverted colors or solid underline for the focused conversion clause).
- **Candidate Window Positioning:** Reporting accurate screen-space caret coordinates to the operating system so the IME candidate selection menu appears adjacent to the cursor.
- **State Transitions:** Handling arrow key clause navigation, Space/Enter conversion and commit, Escape composition cancellation, and backspace within preedit.
- **Platform Protocols:** Supporting Windows Text Services Framework (TSF), Linux IBus / Fcitx / Wayland text-input, macOS `NSTextInputClient`, and Android `InputConnection`.

---

## Strategy & Principles

1. **Unified Host Interface:** Abstract composition state into `internal/host` so the page renderer remains decoupled from platform-specific IME protocols.
2. **Non-Destructive Preedit:** Composing text must render visually inside the focused form control without mutating the underlying data-bound model value until final commit.
3. **Hardware / Device Validation:** Automated tests simulate composition events; final validation requires interactive verification on native OS input methods (e.g. Japanese Microsoft IME, Google Gboard, macOS Kotoeri).

---

## Checklist

### 3.1 Host & Page Composing Model

- [ ] `IME-01` Define the unified `host.IMEComposition` structure in `internal/host`:
  ```go
  type IMEComposition struct {
      Text         string
      Cursor       int
      ClauseRanges [][2]int
      TargetClause int
  }
  ```
  *Acceptance Criteria:* `internal/host` exports composition primitives; host implementations compile without cyclic dependencies.

- [ ] `IME-02` In `internal/page/form`: Render preedit text inline with distinct visual styles (dotted underline for unselected clauses, solid/reversed underline for active clause). Keep preedit distinct from the committed field value.  
  *Acceptance Criteria:* Unit tests verify that preedit does not fire `Handlers.Change` or overwrite `page.Form().Value()` until committed.

### 3.2 Windows TSF / IMM32 Integration

- [ ] `IME-03` In `internal/window`: Implement Windows composition event handling (`WM_IME_STARTCOMPOSITION`, `WM_IME_COMPOSITION`, `WM_IME_ENDCOMPOSITION`, or TSF `ITfContextOwnerCompositionSink`).  
  *Acceptance Criteria:* Windows window receives composition messages and forwards preedit strings to `host.Screen`.

- [ ] `IME-04` Calculate the bounding rectangle of the current form caret in desktop screen coordinates and notify Windows via `ImmSetCandidateWindow` or TSF view extents.  
  *Acceptance Criteria:* Japanese Microsoft IME candidate popup appears docked immediately beneath the text caret in `examples/login` on Windows.

### 3.3 Linux IBus / Fcitx & Wayland Integration

- [ ] `IME-05` Implement D-Bus client bindings for IBus (`org.freedesktop.IBus`) and Fcitx on Linux X11 desktop sessions. Handle `CommitText`, `UpdatePreeditText`, and `ShowPreeditText`.  
  *Acceptance Criteria:* Composing with `ibus-anthy` or `fcitx5` displays inline preedit on Linux desktop.

- [ ] `IME-06` Implement Wayland `zwp_text_input_v3` protocol handling in the Ebitengine/window loop for pure Wayland sessions. Pass surface-relative cursor rectangles via `set_cursor_rectangle`.  
  *Acceptance Criteria:* Native Wayland compositor positions the virtual keyboard/candidate window accurately relative to OwnFrame fields.

### 3.4 macOS Cocoa Integration

- [ ] `IME-07` In `internal/window/ime_darwin.go`: Implement Cocoa `NSTextInputClient` protocol methods (`hasMarkedText`, `markedRange`, `selectedRange`, `setMarkedText:selectedRange:replacementRange:`, `insertText:replacementRange:`, `firstRectForCharacterRange:actualRange:`).  
  *Acceptance Criteria:* Composing in macOS with Japanese/Pinyin inputs renders inline marked text and commits on Return.

### 3.5 Android Soft Keyboard & InputConnection

- [ ] `IME-08` In `examples/android-host` and `internal/window/ime_android.go`: Extend `EbitenView` with an `InputConnection` implementing `setComposingText`, `finishComposingText`, `commitText`, and `setSelection`. Accurately track IME insets during soft keyboard display.  
  *Acceptance Criteria:* Gboard predictive suggestions and multi-word completions commit cleanly into OwnFrame fields without losing cursor position or producing duplicate characters.
