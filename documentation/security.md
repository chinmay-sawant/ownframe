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

## Boundary by API

- Web mode: [web.md](web.md)
- Fetch and XHR: [fetch.md](fetch.md)
- Crash reports: [crash.md](crash.md)
- IPC: [ipc.md](ipc.md)

Residual risks that ownframe cannot remove on its own include unsandboxed Go
templates, missing host allowlists in `Fetch`, trusted path APIs, and an
unauthenticated `Serve`. The audit lists the full set and the recommended fixes.