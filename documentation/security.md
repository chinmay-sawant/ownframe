# Security

ownframe treats templates, themes, and source files passed to `New`, `Load`,
`Route`, `Config`, and `SetData` as trusted author content. End user input,
remote URLs, and filesystem paths passed through app code need validation in the
app. This page collects the trust boundaries and links to the full audit.

## Audit

The complete audit lives under
[`plans/v0.0.2/security/`](../plans/v0.0.2/security/). Start with the overview,
then read the focused notes:

- [00-overview-threat-model-and-priority.md](../plans/v0.0.2/security/00-overview-threat-model-and-priority.md)
- [01-template-content-xss-and-injection.md](../plans/v0.0.2/security/01-template-content-xss-and-injection.md)
- [02-web-server-network-and-fetch.md](../plans/v0.0.2/security/02-web-server-network-and-fetch.md)
- [03-file-paths-print-commands-and-local-data.md](../plans/v0.0.2/security/03-file-paths-print-commands-and-local-data.md)
- [04-runtime-devtools-crash-and-residual-risks.md](../plans/v0.0.2/security/04-runtime-devtools-crash-and-residual-risks.md)
- [05-security-phase-checklist.md](../plans/v0.0.2/security/05-security-phase-checklist.md)

## Assumptions

- Templates, themes, and files named in `Config` and passed to `Page` are
  trusted application content.
- Path inputs such as `Config.File`, `ThemeFile`, `SetCrashDir`, and `SavePDF`
  are trusted configuration.
- `Serve`, devtools, and debug flags only make sense on a machine the caller
  controls.
- `Fetch` and `XHR` load remote URLs chosen and validated by the app.

## Drop path handling

`Drop.Path` is desktop only and is an absolute local path. It is sensitive
and must not be surfaced in UI, logs, crash reports, web previews, or IPC.
Copy bytes from `Drop.Read` during the `Drop` call; the read is only valid
for that frame. See [drag-drop.md](drag-drop.md).

## Mobile captured images

Examples that accept camera or gallery images (such as `examples/telegram`)
decode them inside `addPhoto` before allocation. Callers should bound image
dimensions before decoding and reject oversized or malformed payloads rather
than allocating from an untrusted source. Photos are queued on a bounded
channel and applied on the game loop.

## Debug and developer surfaces

- `Serve` exposes no `/debug/state` endpoint. `ServeWithOptions` registers
  `GET /debug/state` only with `ServeOptions{Debug: true}` or
  `ServeOptions{Perf: true}`. Treat it as privileged. Reduce or gate it before
  exposing a preview. `ServeOptions{Perf: true}` is the only opt in that adds
  `GET /debug/pprof/*`.
- DevTools summarizes rendered data on screen. Keep it off in shipped binaries
  unless debugging is intended.
- Crash reports can contain stack traces, panic text, page titles, and source
  paths. Directory mode is `0700` and file mode `0600`, but callers should
  still choose a private directory on shared machines.
- `SetCrashDir` and other path APIs are trusted; the crash directory is not
  sandboxed.
- Environment debug flags (`OWNFRAME_PRINT_DEBUG`, `OWNFRAME_FILEPICK_DEBUG`)
  write to stderr and should remain off in shipped binaries.

## Boundary by API

- Web mode: [web.md](web.md)
- Fetch and XHR: [fetch.md](fetch.md)
- Crash reports: [crash.md](crash.md)
- IPC: [ipc.md](ipc.md)
- Drag and drop: [drag-drop.md](drag-drop.md)

Residual risks that ownframe cannot remove on its own include unsandboxed Go
templates, missing host allowlists in `Fetch`, trusted path APIs, and an
unauthenticated `Serve`. The current implementation does add transport
timeouts on `Serve`, `nosniff` and framing headers, MaxBytes parsing for
`/type`, finite coordinate checks, a 30 second fetch timeout, a 10 MiB fetch
body cap, crash files at mode `0600`, and a source size limit of
`MaxSourceSize` (8 MiB). The audit tracks the remaining items.


## Helper executables

Native helper tools such as `lp`, `xdg-open`, `zenity`, `powershell.exe`, and `pwsh.exe` are discovered through `PATH` and their normal OS locations, including WSL `/mnt` paths. Treat the executable search path and installed helpers as trusted local environment components.

## Devtools clipboard

The devtools dock has no setting for OS clipboard writes: clicking an Ops row copies its displayed text to the clipboard. Keep devtools closed or disabled when rendered data or coordinates are sensitive.


## Captured photo retention

The Telegram example registers each captured photo under a `photo-N` key and keeps it in `Page` image memory until the app calls `SetImage(id, nil)`. There is no automatic eviction. An app that handles sensitive images should remove the entry when the message or chat no longer needs it.
