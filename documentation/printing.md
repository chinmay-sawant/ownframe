# Printing and PDF export

`Page.PDF`, `Page.WritePDF`, `Page.SavePDF`, and `Page.Print` all refuse
with `ErrNoPDF` (`"ownframe: blinkless does not write PDF"`). The engine
returns a drawing list and stops there, so there is no document writer to
feed. `SavePDF` still validates its path first, so a bad path reports the
path error instead.

## The calls

| Call | What it does |
|------|--------------|
| `Page.PDF(ctx, opts) ([]byte, error)` | Returns `nil, ErrNoPDF`. |
| `Page.WritePDF(ctx, w, opts) error` | Returns `ErrNoPDF`. |
| `Page.SavePDF(ctx, path, opts) error` | Validates `path`, then returns `ErrNoPDF`. |
| `ownframe.SavePDF(ctx, page, path, opts) error` | The same, for a caller that holds the page separately. |
| `Page.Print(ctx, opts) error` | Fails before any OS helper runs, see below. |
| `Page.SupportsPDF() bool` | Always `false`. Branch on it before calling any of the above. |

`PDFOptions` carries three fields. `PageSize` is an engine name such as
`"A4"` or `"Letter"`. `Margin` is one width in millimetres applied to all
four sides. `Profile` is a PDF conformance profile such as `"PDF/A-4"` or
`"PDF/UA-1"`. The fields are currently unused: every call above refuses
before reading them.

```go
if !page.SupportsPDF() {
    // No PDF writer in this build. Use page.PNG and print those bytes.
    png, err := page.PNG(ctx)
    ...
}
_, err := page.PDF(ctx, ownframe.PDFOptions{PageSize: "Letter", Margin: 20})
if errors.Is(err, ownframe.ErrNoPDF) {
    // Render the page another way, e.g. page.PNG, and print those bytes.
}
```

## Printing

`Page.Print` checks for an OS print helper first, then renders through
`WritePDF`, so it never reaches the helper today:

| System | Commands |
|--------|----------|
| Linux | `lp`, then `xdg-open` |
| macOS | `osascript` asking Preview to print, then `open` |
| Windows | `powershell.exe`, then `pwsh.exe`, with `Start-Process -Verb Print` |
| wasm, Android, iOS | none, so `Print` returns `ErrNoPrinter` |

On desktop the call fails with `ErrNoPDF` from `WritePDF`. Without a
helper it fails earlier with `ErrNoPrinter`. Both errors name `SavePDF`
as the fallback, but `SavePDF` itself still returns `ErrNoPDF`, so the
working fallback is `Page.PNG`: encode the picture the window shows and
hand those bytes to the system your app already prints with. No temporary
PDF is left behind in either case, because no PDF is ever written.

`OWNFRAME_PRINT_DEBUG=1` logs each lookup, the command that ran, and each
fallback to stderr.

## Web mode

`Serve` registers `GET /pdf` when the screen implements

```go
PDF(ctx context.Context, opts ownframe.PDFOptions) ([]byte, error)
```

which `*Page` does. The route calls it and, on error, answers `500` with
the error text, so a `*Page` screen answers `500` with the `ErrNoPDF`
message. A screen without that method keeps the old routes, and `/` and
`/frame.png` do not change. `examples/print` serves it on 127.0.0.1:8125.

## Limits

- The engine does not write a PDF. Pagination, page breaks, and headers
  or footers do not exist in `PDFOptions`, and there is no PDF fallback
  inside this module.
- Use `Page.PNG` for a byte picture of the window, and print those bytes
  with an app-side tool.
- Relative resources resolve against nothing, because the source goes to
  the engine as in-memory HTML.
