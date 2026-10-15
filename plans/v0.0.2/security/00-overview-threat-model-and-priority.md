# ownframe security overview and threat model

Security audit of the ownframe library and its example apps, October 2026.

## What ownframe does

ownframe opens an HTML template in a desktop, browser, or Android window. `Page` parses HTML with Go `html/template`, applies CSS through the blinkless rendering engine, and draws to an Ebiten canvas. `Serve` also serves a browser preview. IPC, fetch, print, file dialogs, clipboard, and drag and drop are helper APIs that apps can call.

## Trust boundaries

1. Template author. `New`, `SetData`, `Handle`, `Load`, and `Route` take trusted application code. ownframe does not sandbox Go templates. A template can invoke exported methods on `Page` data during execution.
2. Configuration. `Config.File`, `Config.ThemeFile`, `SetCrashDir`, and PDF paths are trusted application paths.
3. End users. Users supply form values, clicks, pasted text, dropped files, fetch URLs, exported data, and link choices. Anything that reaches a `Handler`, `Filter`, or app callback must be validated in the app.
4. Local host. `Serve`, the devtools dock, profilers, debug environment flags, and the crash directory only make sense when the caller controls the machine.
5. External resources. `Fetch` and `XHR` can load remote URLs. Templates and themes cannot load remote CSS, fonts, images, or scripts.

## Assets at risk

- Application data exposed through templates and `Page.Data`.
- Local files reachable by `Config.File`, `ThemeFile`, `SavePDF`, crash reports, and file dialogs.
- Network reachability and cloud metadata if apps forward untrusted URLs to `Fetch`.
- Page content and layout when `Serve` is exposed on a non loopback address.
- OS clipboard contents when devtools or browser preview input is exposed.

## Attacker capabilities assumed

- A malicious template author, a compromised dependency, or an app that feeds untrusted data into `Load` or `Route`.
- A network peer who can reach the `Serve` port.
- A local user on the same machine or session who can read world readable files, open devtools, or read a shared clipboard.
- A dropped file or file pick result that an app opens without validation.
- Untrusted URLs passed to `Fetch` or `XHR`.

## Out of scope for ownframe

- Go template sandboxing. Go `html/template` has none.
- Application authentication, session management, or CSRF protection.
- Host allowlisting in `Fetch`.
- Parsing behavior of the external blinkless engine, unless an ownframe specific boundary fails.
- Malware or hostile local environments that replace binaries on `PATH`.

## Priority summary

| Priority | Area | Files |
|---|---|---|
| P1 | `Serve` has no timeouts, no origin check, and mutating GET click endpoint | `internal/web` |
| P1 | `/debug/state` is always available, even in plain web mode | `internal/web/debug_http.go` |
| P1 | Runtime template loading through `Load` and `Route` exposes `Page.Data` | `internal/page/page_nav.go` |
| P2 | Untrusted URLs reach `Fetch` and `XHR` | `internal/fetch` |
| P2 | Path traversal is possible when apps forward untrusted data into file APIs | `internal/page`, `internal/print` |
| P3 | AppleScript, PowerShell, and path argv option injection | `internal/print`, examples |
| P3 | Clipboard, drop metadata, and focus are local convenience APIs | `internal/clipboard`, `internal/window` |
| P4 | Devtools, crash permissions, and debug environment flags | `internal/crash`, `internal/window` |

P1 items can lead to data exposure or state mutation by network peers. P2 items depend on application misuse. P3 and P4 are developer environment issues or require a compromised local machine.