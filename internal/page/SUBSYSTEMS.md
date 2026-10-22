# internal/page subsystems

Index only. No behavior change. Each group names its entry file.

| Group | Entry file | Find |
|---|---|---|
| form | `form_html.go` (control rewrite), `form_input.go` (`clickControl`), `form_focus.go` (Tab order) | field paint, click, focus |
| nav | `page_nav.go` (`Load`, `HTML`) | `Back` / `Forward` in `page_nav_history.go`, `Route` in `page_nav_route.go` |
| dirty | `page_dirty.go` (display diff), `page_dirty_take.go` | `TakeDirty` / `TakeDirtyRects`; `Invalidate` in `page_invalidate.go`; `MarkRect` in `page_frame.go` |
| draw | `page_draw.go` (`Redraw`) | commit in `page_draw_commit.go`; pipeline in `page_pipeline.go` |
| patch | `page_patch_element.go` (`PatchElement`) | `PatchDisplay` / `PatchOp` in `page_patch_display.go`; report in `page_patch.go` |
| a11y | `accessibility_tree.go` (`AccessibilityTree`) | actions in `accessibility_action.go` |
| host | `page_host.go` (interface assertions) | tick in `page_tick.go`; windowing in `page_windowing.go`; pin z in `page_pinz.go` |
| tick | `page_tick.go` (`SetTick`, `Tick`) | op-diff dirty in `page_tick_dirty.go`; `UseFrameDirty` alias in `page_frame.go` |

Notes: `Tick` diffs display ops after the callback. A mutating tick sets `FrameDirty` and queues rects; a no-op tick stays clean so idle park can sleep.
