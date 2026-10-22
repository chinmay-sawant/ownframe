# internal/replay subsystems

Index only. No behavior change. Each group names its entry file.

| Group | Entry file | Find |
|---|---|---|
| tile blit | window `replay_partial.go` (`drawReplayPartial`) | tiles paint with `DrawVisibleUnfixedZ` in `pinned_z.go`; grid lives in `internal/window/replay_tile.go` |
| group / mask | `group_paint.go` (`paintDisplay`) | mask texture in `group_mask.go`; layer in `group_layer.go`; shader in `group_shader.go` |
| text cache | `text_cache.go` (faces) | run raster cache in `text_raster_cache.go`; bake in `text_bake.go` |
| reject / fallback | `internal/render/replayable.go` (`Replayable`) | gate tests in `internal/render/replay_reject_test.go`; bitmap fallback is page-side (`page_draw.go`) |
| DrawRect | `draw_rect.go` | test and helper path only |

The window does not call `DrawRect`. Partial frames go `drawReplayPartial` to tiles, then `DrawVisibleUnfixedZ`. `DrawRect` is for tests and GPU parity.

Pinned z-layers (`DrawPinnedZ`, `DrawVisibleUnfixedZ`) stay off the tile buffer. 1.0 chrome uses `SetScrollWindow` / `SetViewportPinZ` and leaves the tile path.
