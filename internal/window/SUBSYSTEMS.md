# internal/window subsystems

Index only. No behavior change. Each group names its entry file and state owner.

Updated 2026-10-22. `replay_frame.go` `fullReplay`: ticking Pages stay on the buffer path; Tick marks mutated ops.

| Group | Entry file | State owner |
|---|---|---|
| devtools | `devtools_dock.go` (overlay open/close, dock layout) | `devState` in `devtools.go`, held as `shell.dev` (`game_state.go`) |
| idle | `idle_policy.go` (`pacePolicy.step`, per-frame decision) | `pacePolicy` in `idle_policy.go`; published `paceState` in `idle_state.go`, held as `shell.pace` |
| touch | `touch_gesture.go` (`touchGesture.frame`) | `touchGesture` in `touch_gesture.go`, held as `gestureState.fingers` (`game_gesture_state.go`) |
| pointer | `pointer.go` (`shell.pointer`, fans out to `pointer_mouse.go`, `pointer_touch.go`, `pointer_drag.go`) | `shell` pointer fields (`mouseDown`, `cursorX/Y`, `hover` in `game_state.go`); drag shape `pointerDrag` |
| replay tile | `replay_partial.go` (`drawReplayPartial`) | `partialState` in `replay_partial_state.go` (`tiles *tileGrid` from `replay_tile.go`), held as `shell.partial`; lifecycle in `replay_tile.go` |
| reload | `reload.go` (`pollReload`), staged as `reload` in `update.go` | `shell.lastPoll` / `shell.lastNote` in `game_state.go` |
| a11y | `a11y_atspi.go` (Linux) / `a11y_atspi_other.go` | `a11yHost` in those files, held as `shell.a11y`; sync from `syncImage` |

Notes: `update.go` stages wire pointer, touch, and reload into each frame. `replay_tile.go` defines the retained-content handle behind `partialState.buf`.
