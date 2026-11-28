# Compile-time HTML code generation (`ownframegen`)

`cmd/ownframegen` compiles `.html` templates into plain Go source files before your binary builds. The generated code writes HTML directly into a `strings.Builder` without `html/template` reflection, exposes a fluent builder for your view struct, extracts static element IDs and `data-action` names as typed Go constants, and emits a fluent `PatchBuilder` for batched `Page.PatchElements` calls.

A complete runnable app and benchmark suite live in [examples/codegen](../examples/codegen).

## Why compile HTML templates to Go

`html/template` walks your structs through `reflect` on every `Page.Redraw`, allocating per-field values and formatting numbers on the heap. It also leaves element IDs and `data-action` strings unchecked until runtime: a typo in `id="status-badge"` fails only when a user clicks or a patch runs.

Compiling the template with `cmd/ownframegen` fixes both problems at build time:

- Static HTML chunks compile to direct `b.WriteString(...)` calls with a pre-sized `b.Grow(...)` estimate.
- Typed numeric fields (`int`, `int64`, `uint64`, `float64`) format through `strconv.AppendInt` / `strconv.AppendFloat` into a stack buffer with zero heap allocations.
- Untrusted string fields escape through `html.EscapeString`, while `{{raw .Field}}` inserts trusted fragments (such as rendered `ui.Element` trees) directly.
- Every static `id="..."` becomes a typed constant (`IDStatusBadge`) and a set of fluent patch methods (`StatusBadgeText`, `StatusBadgeAddClass`, `StatusBadgeToggleClass`, `StatusBadgeAppendHTML`, `StatusBadgeInnerHTML`, `StatusBadgeRemoveChild`).
- Every static `data-action="..."` becomes a typed constant (`ActionToggleStatus`).

## Measured benchmarks

Run the benchmark suite yourself on any machine with no display server required:

```sh
go test -run='^$' -bench=. -benchmem ./examples/codegen/...
```

Measured on Linux amd64 (13th Gen Intel Core i7-13700HX, Go 1.26.11):

### HTML generation (`html/template` vs generated `RenderTo`)

| Workload | `html/template` | `ownframegen` (`RenderTo`) | Speedup | Allocs (`html/template` -> `ownframegen`) |
|---|---:|---:|---:|---:|
| 50 rows | 191.3 µs (26.3 KB) | **6.8 µs** (18.4 KB) | **28.2x** | 1,322 -> **1** |
| 200 rows | 686.2 µs (103.3 KB) | **26.9 µs** (65.5 KB) | **25.5x** | 5,222 -> **1** |
| 1,000 rows | 3,573.4 µs (520.6 KB) | **126.8 µs** (270.3 KB) | **28.2x** | 26,767 -> **1** |

### Full OwnFrame redraw vs generated `PatchBuilder` (200 rows)

| Path | Time/op | Bytes/op | Allocs/op |
|---|---:|---:|---:|
| `Page.Redraw` (`html/template`) | 4.24 ms | 1.45 MB | 9,351 |
| `Page.Redraw` (`Config.Render` generated) | 4.62 ms | 1.50 MB | **3,785** |
| `PatchBuilder` (1 element, `RetainState: true`) | **58.3 µs** | **62.0 KB** | **72** |
| `PatchBuilder` (10 elements batched) | **317.2 µs** | **292.6 KB** | **924** |

Two details stand out in those numbers. First, `Config.Render` cuts full-redraw allocations by 60% (9,351 down to 3,785), while Blinkless HTML parsing and layout still account for most of a full redraw's wall time. Second, pairing the generated `PatchBuilder` with `Config.RetainState: true` skips both HTML re-serialization and full-page relayout, running **73x faster** than a full redraw for 1 element and **13.4x faster** for 10 elements in one batch.

## Running `cmd/ownframegen`

Run the CLI directly:

```sh
go run github.com/chinmay-sawant/ownframe/cmd/ownframegen \
  -in view.html \
  -out view_gen.go \
  -pkg main \
  -type DashboardView
```

Or wire it into `go generate` inside your Go package:

```text
//go:generate go run github.com/chinmay-sawant/ownframe/cmd/ownframegen -in view.html -out view_gen.go -pkg main -type DashboardView
```

Inside this repository's `examples/` workspace, use a relative path so `go generate` runs the local checkout:

```text
//go:generate go run ../../cmd/ownframegen -in view.html -out view_gen.go -pkg main -type DashboardView
```

### Flags

| Flag | Default | Description |
|---|---|---|
| `-in` | required | Input `.html` file, glob pattern, or directory of `.html` files. |
| `-out` | `<stem>_gen.go` | Output `.go` file path, or `-` to write to stdout. |
| `-pkg` | `main` | Go package name for the generated file. |
| `-type` | `<FileStem>View` | Generated view struct name (single-file input only). |

## Template syntax

Templates use standard `{{ ... }}` delimiters with optional Go type annotations on fields:

```html
{{/* Comments are stripped at compile time */}}
<div id="status-badge" class="badge{{if not .Online}} offline{{end}}">
  {{.Title}}
</div>
<span id="open-count">{{.OpenCount:int}}</span>
<span id="uptime-pct">{{.UptimePct:float64}}</span>
<div class="slot">{{raw .BannerHTML}}</div>
<button data-action="toggle-status">Toggle</button>
<div id="task-list">
  {{- range .Tasks -}}
  <div id="task-{{.ID:int}}" class="row{{if .Done}} done{{end}}">
    {{.Name}}
  </div>
  {{- end -}}
</div>
```

Supported constructs:

- **Escaped string field.** `{{.Title}}` generates a `string` field and emits `html.EscapeString(v.Title)`.
- **Typed scalar field.** `{{.Count:int}}`, `{{.Total:int64}}`, `{{.Bytes:uint64}}`, `{{.Price:float64}}`, and `{{.Flag:bool}}` generate typed struct fields and format them with `strconv` without reflection. Any other Go type falls back to `html.EscapeString(fmt.Sprint(...))`.
- **Trusted raw HTML slot.** `{{raw .BannerHTML}}` writes the string directly without escaping. Use this for pre-built `ui.Element` HTML rendered through `ui.Render`.
- **Conditionals.** `{{if .Flag}}...{{else}}...{{end}}` and `{{if not .Flag}}...{{end}}` infer `bool` unless the field already has another type (strings check `!= ""`, numbers check `!= 0`, slices check `len(...) > 0`).
- **Loops.** `{{range .Tasks}}...{{end}}` generates a slice field (`Tasks []DashboardViewTask`) and a typed item struct with its own fluent builder methods.
- **Whitespace trim and comments.** `{{-` and `-}}` trim adjacent whitespace at compile time. `{{/* ... */}}` comments emit no Go code.
- **Static IDs and actions.** Every static `id="..."` and `data-action="..."` emits a Go constant (`IDStatusBadge`, `ActionToggleStatus`). Dynamic IDs containing `{{` (such as `id="task-{{.ID:int}}"`) are skipped for constant generation and can be targeted with `PatchBuilder.ByID(...)`.

## Using the generated builder, `Config.Render`, and `PatchBuilder`

Construct the view with fluent builder methods and pass `view.Render` to `ownframe.Config`:

```text
view := NewDashboardView().
    WithTitle("OwnFrame Codegen Dashboard").
    WithOnline(true).
    WithOpenCount(3).
    WithBannerHTML(bannerHTML).
    AddTask(DashboardViewTask{ID: 1, Name: "Compile view.html", Owner: "build"})

page, err := ownframe.New(ownframe.Config{
    Title:       "OwnFrame Codegen Dashboard",
    Render:      view.Render,
    Width:       960,
    Height:      640,
    RetainState: true,
})
```

Handle clicks with generated action constants and batch retained DOM updates with `view.Patch()`:

```text
switch box.Action {
case ActionToggleStatus:
    return view.Patch().
        StatusBadgeText("Offline").
        StatusBadgeToggleClass("offline").
        LastEventText("Status switched to Offline").
        Apply(ctx, page)
}
```

Multiple fluent calls targeting the same element ID (such as `StatusBadgeText` followed by `StatusBadgeToggleClass`) coalesce into a single `ownframe.ElementPatch` before `Apply` calls `Page.PatchElements`. For dynamic range rows, chain `.ByID("task-1", ownframe.ElementEdit{Text: &txt})` onto the same patch batch.
