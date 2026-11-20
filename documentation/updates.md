# Choosing an update path

OwnFrame renders HTML templates through Go rather than a live browser DOM. Pick the simplest update path that matches how often and how much of the screen changes:

| What changed | Recommended API | Why |
|---|---|---|
| **Normal UI state** (navigation, form submit, tab switch, modal open/close) | `Page.SetData` + `Page.Redraw` | Executes the template once, diffs the display list, and repaints only damaged tiles. Best default for almost every interaction. |
| **Small targeted update** (status text, badge count, single-row class/attribute on a large page) | `Page.PatchElement` / `Page.PatchElements` | Mutates specific element text or attributes in the cached document and relayouts without re-executing the full template. Falls back cleanly to a full redraw when needed. See [optimizations.md](optimizations.md). |
| **Large scrollable list** (>200 rows: logs, chat threads, file explorers, tables) | `ui/virtual.Window` + `Page.SetScrollWindow` | Renders only the visible slice plus overscan rows, padded by top and bottom spacers so layout stays proportional to the viewport rather than total row count. See [scrolling.md](scrolling.md). |
| **Declarative animation** (hover transitions, spinners, slide/fade keyframes) | CSS `@keyframes` and `transition` | Driven automatically by the window tick with retained 2D transform/opacity layers. No Go animation loop required. See [css-motion.md](css-motion.md). |
| **Imperative per-frame motion** (audio seekbar, physics/game sprite, live telemetry) | `Page.SetTick` + `internal/frame` | Patches retained display ops directly each frame without a redraw. See [frames.md](frames.md). |

## Rules of thumb

1. **Start with `SetData`.** Only drop down to `PatchElement` or `ui/virtual` when profiling (`WithPerf(true)` or DevTools Frame tab) shows template or layout cost dominating frame time.
2. **Never drive one element from both CSS motion and `SetTick`.** Turn off `Page.Window().SetAnimations(false)` or `Page.Window().SetTransitions(false)` if a Go tick callback mutates the same element's retained ops.
3. **Keep `Page` single-threaded.** Run background I/O or git/network work in goroutines, then queue results back onto the page's event/tick loop before calling `SetData`, `PatchElement`, or `Redraw`.
