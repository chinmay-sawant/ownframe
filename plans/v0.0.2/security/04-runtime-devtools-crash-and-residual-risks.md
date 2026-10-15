# Runtime, devtools, crash reporting, and residual risks

## Devtools and debugging surfaces

Location: `internal/window`, `internal/web`, `internal/crash`

The devtools dock opens with F12 or Ctrl+Shift+I even when `Config.DevTools` is false. It exposes element properties, operation inventory, and frame metrics for whatever the page renders. `Serve` exposes `/debug/state` without authentication. Crash report files and temporary assets may carry local details. Debug environment variables such as `OWNFRAME_FILEPICK_DEBUG`, `OWNFRAME_PRINT_DEBUG`, `OWNFRAME_PACE_DEBUG`, and `OWNFRAME_DISPLAY_PATCH_DEBUG` print paths, commands, and fallback reasons.

### Vulnerabilities

- Severity: Medium
  - `/debug/state` is always available in plain web mode. It returns box text, coordinates, stats, and operation counts.
- Severity: Medium
  - Crash report directories use mode `0755` and files use mode `0644`.
- Severity: Low
  - Devtools can be opened by anyone with local keyboard access, regardless of startup configuration.
- Severity: Low
  - Mobile and desktop runtimes recover panics into local crash files. `Serve` does not recover panics at all.

### Valid scenario

A developer opens devtools on a trusted local machine to inspect dimensions and operations.

### Invalid scenario

A remote network user reaches the web address and reads `/debug/state` without having to enable any debugging option.

### Recommended fixes

1. Gate `/debug/state` behind a debugging option or return only counts instead of box text.
2. Make crash files and directories user private.
3. Keep devtools disabled outside development workflows.
4. Avoid enabling debug environment variables in shipped binaries.
5. Recover or isolate panics in `Serve`, or document why that boundary is intentional.

## Mobile and example app surfaces

Location: `examples/telegram/mobile`, `examples/telegram/android`, `examples/desktop-cat`, `examples/snipping-tool`

The Android wrapper calls exported mobile functions such as `Back`, `AttachRequest`, `SetPhoto`, and `SetInsets`. The desktop cat HTTP server defaults to `127.0.0.1:6969`, limits POST bodies to 4096 bytes, and rejects unknown JSON fields. Example apps can still bind broadly or expose file paths.

### Valid scenario

The desktop cat notification server accepts local requests with an expression and message.

### Invalid scenario

An app copies example code that binds `0.0.0.0` or logs path result from Android intents and file dialogs.

### Recommended fixes

1. Review every example that binds addresses before shipping it unchanged.
2. Validate bytes and intent results in each mobile handler, including size and extension.
3. Keep desktop cat server settings documented, including localhost binding.

## Confirmed non issues

- CSS cannot exfiltrate data through network URLs. Registered image names are required before a CSS `url(...)` paints. There is no filesystem or network fetch in the ownframe CSS path.
- Templates cannot load remote stylesheets, fonts, or scripts because ownframe does not run browser HTML.
- `Jump` and monitor rendering do not include devtools in `Page.PNG`, display lists, or box lists.
- `/debug/pprof/*` appears only when `ServeOptions{Perf: true}`.
- `RuntimeSnapshot` stays nil in normal web mode, so runtime metrics are not returned there.
- Hot reload does not execute arbitrary code. It reads and reparses files already named in application configuration.
- `file`, `data`, `javascript`, `vbscript`, `blob`, and `about` deep link schemes are blocked at configuration time.

## residual risks that need application side controls

- Go templates are not sandboxed. Apps must not load user or network HTML into `Page.Load` or `Page.Route`.
- `Fetch` and `XHR` have no host allowlist. Apps must validate URLs and add timeouts.
- Path containing APIs are trusted. Apps must validate data before passing it into `Config.File`, `ThemeFile`, `SetCrashDir`, `SavePDF`, or file dialog results.
- `Serve` has no authentication. Apps must bind locally or add their own protections.
- Clipboard content can contain secrets. Apps should avoid exposing clipboard in preview mode.
- Drop paths are sensitive. Apps should avoid logging them.
- Crash reports may include source details. Apps should choose secure crash directories on shared systems.

## Primary follow up work

1. Gate `/debug/state` and `/click`.
2. Add timeouts and body limits.
3. Harden `Fetch` availability and document its trust boundary.
4. Crash report permission changes.
5. Template loading documentation and no template execution option.
6. Fuzz post template HTML rewriters and `injectTheme`.