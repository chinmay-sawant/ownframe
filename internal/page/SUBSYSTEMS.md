# internal/page subsystems

Index only. No behavior change. Each group names its entry file.

| Group | Entry file | Find |
|---|---|---|
| form | `internal/page/form` (control, caret, focus, value helpers); `form_html.go` (control rewrite), `form_input.go` (`clickControl`), `form_focus.go` (Tab order) | field paint, click, focus |
| nav | `page_nav.go` (`Load`, `HTML`) | `Load`, `Back`, `Forward`, `Route`, and history placement together in `page_nav.go` |
| dirty | `page_dirty.go` (display diff), `page_dirty_box.go` | `TakeDirty` / `TakeDirtyRects`; `Invalidate` in `page_dirty_box.go`; `MarkRect` in `page_tick.go` |
| draw | `internal/page/draw` (`Frame`, list vs bitmap); `page_draw.go` (`Redraw`) | commit in `page_draw.go`; pipeline in `page_pipeline.go` |
| patch | `page_patch_element.go` (`PatchElement`) | `PatchDisplay` / `PatchOp` in `page_patch_display.go`; report in `page_patch.go` |
| a11y | `internal/page/a11y` (roles, names, geometry); `accessibility_tree.go` (`AccessibilityTree`) | actions in `accessibility_action.go` |
| scroll | `internal/page/scroll` (`State`, `Windowing`); `page_scroll.go` | windowing in `page_scroll.go`; pin z in `page_scroll.go` |
| host | `page_host.go` (interface assertions) | tick in `page_tick.go`; windowing in `page_scroll.go`; pin z in `page_scroll.go` |
| tick | `page_tick.go` (`SetTick`, `Tick`) | op-diff dirty in `page_tick.go`; `UseFrameDirty` alias in `page_tick.go` |

The page owns its changed-property mutex; independent pages do not share a metadata lock. Paint damage invalidates the encoded PNG cache through `markRect` and `markFull`.

Notes: `Tick` diffs display ops after the callback. A mutating tick sets `FrameDirty` and queues rects; a no-op tick stays clean so idle park can sleep.
