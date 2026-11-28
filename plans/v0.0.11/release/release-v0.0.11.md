# OwnFrame v0.0.11

Release on top of `v0.0.10`, delivering hover/a11y/text-cache performance fixes for dense interactive views, non-form `PatchElements` fast paths without trailing full redraws, and `ownframegen -bind` full Go `html/template` compilation against existing Go structs.

- Range: `v0.0.10` through `v0.0.11` (2026-11-28)
- Engine: pins `github.com/chinmay-sawant/blinkless` to `v0.1.12`

## Summary

| Area | What changed |
| :--- | :--- |
| **Hover, A11y, & Viewport Invalidation (`internal/page`, `internal/page/a11y`, `internal/window`)** | Keyed `publishA11y` and viewport full-repaint / `EvictDisplay` invalidation on `layerGeneration()` instead of `app.Generation()` so pure hover/cursor updates never rebuild AT-SPI trees or evict cached GPU surfaces. Added zero-slice-allocation `collapseSpaces`, `nodeBoxesForRoles` (skipping `#document`, `#text`, hidden, and non-role spans), zero-allocation `Page.BoxByID(id string) (Box, bool)`, and gated `markFormDirty` in `paintHover` on `handled == true`. |
| **Text Raster Cache (`internal/replay/text_cache.go`, `renderer.go`)** | Raised `textRunMaxEntries` (`1024 -> 4096`) and `textRunMaxPixels` (`8M -> 16M`), stored `*textRasterEntry` with inline `tick uint64` for single-lookup cache hits, and added batched eviction (`evictBatch`, 64 victims per scan). |
| **Retained `PatchElements` Non-Form Fast Path (`internal/page/patch.go`)** | Added `editsTouchFormControls` and `clearFormDirty(p)` inside `PatchElements` so structural `InnerHTML` patches without `<input>`/`<textarea>`/`<select>` controls do not trigger a trailing full template+layout redraw on `RedrawIfChanged`. |
| **Full Go Template Compiler (`cmd/ownframegen/gotmpl.go`, `main.go`, `gen_test.go`)** | Added `ownframegen -bind <StructType>` (`-bind-dir`, `-func`, comma-separated `-in`) using stdlib `text/template/parse` and `go/parser` (including cross-package local module struct resolution). Compiles full Go `html/template` syntax (`and`/`or`/`not`/`eq`/`ne`/`lt`/`le`/`gt`/`ge`/`len`, `else if`, `$` root refs, `$var := ...`, `range $i, $v := ...`, `with`, `define`/`template`, nested dotted paths, and `template.HTML`) into zero-reflection `Render<Func>Buf` / `Config.Render` functions writing directly into `ownframe`'s pooled `*bytes.Buffer`. |

## Measured Impact (Linux amd64, i7-13700HX)

- **Hover + A11y (`BenchmarkDiffHoverAndA11y`):** `11.6 µs/op`, `0 B/op`, `0 allocs/op`
- **`ownframegen -bind` (19 templates, 2,874 actions):** `2.14 ms / 24,042 allocs` (`html/template`) → **`32.6 µs / 0 allocs`** (**65.7× faster**); end-to-end full redraw `-14.4%` time and `-39.4%` allocs (`61,074 → 37,039`)

## Breaking changes

None.

## Validation

- `go test -count=1 ./internal/replay/... ./internal/page/... ./internal/window/... ./cmd/ownframegen/...`
- PR CI checks green on `master`
