# Lint findings

This page is historical. It is a snapshot from when the `make lint` gate was
added, against commit `master` with the working tree as it stood at that
moment. The counts below are not current.

The live gate is golangci-lint v2.15.0 (`GOLANGCI_LINT_VERSION` in the
Makefile) with a curated list in `.golangci.yml`, then `make size-check`. Do
not treat this file as an `enable-all` or v1 policy.

The snapshot used golangci-lint v1.64.8. At that time the policy started from
`enable-all` and disabled only the linters that restated house style, so the
numbers below were the defect count then, not a formatting backlog.

## Totals

| Tree | Findings |
|------|----------|
| Library (root module) | 614 |
| Examples module | 1681 |
| **Total** | **2295** |

## Dead code (`unused`)

30 unexported symbols with no caller. The library ones are safe to delete
unless a plan is about to use them.

Library (8):

- `internal/element/reconcile_nodes.go:44`: `cloneNode`
- `internal/page/page_form_focus_hit.go:4`: `(*Page).isPressUsed`
- `internal/replay/color.go:20,31,35`: `withAlpha`, `channel`, `clamp`
- `internal/state/path_write.go:16`: `setValueAtPath`
- `internal/window/devtools_line_width.go:9`: `devLine.width`
- `internal/window/replay_plan.go:10`: `repaintFull`

Examples (22):

- `examples/css-preview/cost_play_test.go:27`: `costStartAll`
- `examples/css-preview/cost_report_test.go:49`: `costSpan`
- `examples/css-preview/paint.go:14,20`: `inside`, `write`
- `examples/{dino,eyes,lg-remote,telegram}/mobile/page_profile_other.go:7`: `setAndroidPageProfile` (4 copies)
- `examples/flappy-bird/flappy/flappy.go:29`: `point`
- `examples/flappy-bird/flappy/helpers_test.go:64`: `release`
- `examples/live-log-viewer/ui/fake_test.go:48`: `(*fakeFeed).count`
- `examples/live-log-viewer/ui/format.go:64`: `elapsedText`
- `examples/pet-care/mobile/start_other.go:5`: `start`
- `examples/spotify-player/player/seek_animate.go:11`: `(*App).animateSeek`
- `examples/spreadsheet/ui/worker.go:47`: field `text`
- `examples/system-monitor/collector/live.go:15,17,19`: `unsupported` and its methods
- `examples/system-monitor/storage/errors.go:39`: `errNotImplemented`
- `examples/teams/files/filter.go:6`: `clone`
- `examples/telegram/telegram/box_test.go:51`: `boxText`
- `examples/tetris/game/helpers_test.go:21`: `boardFrom`

## Magic numbers (`mnd`)

1156 findings (257 library, 899 examples). The library ones cluster in the
packages that do bit-level work, where the literal is the protocol or the
pixel math: `internal/window` (89), `internal/clipboard` (40, the X11 wire
format), `internal/replay` (24), `internal/bitmap` (20), `internal/page`
(14), `internal/masking` (11).

The examples ones cluster in the demo renderers: `examples/eyes/eyes` (158),
`examples/spotify-player/player` (67), `examples/system-monitor/collector`
(52), `examples/lg-remote/remote` (52), `examples/spreadsheet/ui` (36),
`examples/dino/dino` (34), `examples/telegram/telegram` (33).

## Other high-impact findings

| Linter | Library | Examples | What it catches |
|--------|---------|----------|-----------------|
| `staticcheck` | 79 | 7 | 66 uses of deprecated Ebiten and `image` APIs (`image.Dispose`, `ebiten.TouchIDs`, `ebiten.ScheduleFrame`, `ebiten.SetFPSMode`), 16 nil `context.Context` in tests, 2 possible nil pointer dereferences (`SA5011`) |
| `gosec` | 23 | 145 | 94 `G115` integer overflow conversions, 39 `G306` file permissions above 0600, 11 `G404` weak random, 8 `G202` SQL string building, 2 `G204` subprocess with tainted input, 1 `G107` URL from variable |
| `errcheck` | 7 | 51 | unchecked errors, including `internal/ipc/send.go:25` and `internal/web/serve_build.go:58` in library code |
| `errorlint` | 13 | 42 | `err ==` comparison and `%w` misuse |
| `forcetypeassert` | 48 | 12 | unchecked type assertions that can panic |
| `exhaustive` | 18 | 40 | switches over `layout.OpKind` and other enums that miss cases |
| `goconst` | 35 | 101 | repeated string literals that should be constants |
| `revive` | 45 | 134 | exported-symbol and naming rules, plus `redefines-builtin-id` on `max`, `min`, `close` |
| `nilerr` / `nilnil` | 5 | 6 | a non-nil error swallowed, or a nil error returned with a nil value |
| `ineffassign` / `wastedassign` | 3 | 2 | assignments never read |
| `unparam` | 4 | 15 | parameters that are always the same value |
| `musttag` | 3 | 10 | struct fields serialized without a JSON tag |
| `noctx` | 0 | 9 | HTTP requests made without a context |
| `contextcheck` / `containedctx` | 5 | 47 | context not passed down, or stored in a struct |
| `goimports` | 8 | 19 | import grouping and formatting |
| `thelper` / `dogsled` | 27 | 33 | test helpers that do not call `t.Helper()`, and blank assignments that hide errors |

## Notes

- `mnd` is disabled for `_test.go` files: fixtures spell out the expected
  pixels and byte layouts on purpose.
- The file-size gate (`make size-check`) is green. Two files are over the
  2,000-character limit and both are recorded in
  `scripts/file-size-allowlist.txt`: `internal/csstransition/bind_to.go`
  (2283) and `internal/page/page_fixed_hit_test.go` (3181).
- `golangci-lint` skips the `examples/` tree by default. The policy sets
  `exclude-dirs-use-default: false` so the examples module is linted too.
