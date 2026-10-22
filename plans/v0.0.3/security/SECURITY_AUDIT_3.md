# Security audit 3 — Additional general considerations

This note complements `SECURITY_AUDIT.md` and `SECURITY_AUDIT_2.md`. The items below are fifty further considerations those two audits do not cover.

Repo: github.com/chinmay-sawant/ownframe, branch `fix/miscellaneous-fixes-and-lint-fix`, Go 1.26.10.
Baseline already covers: govulncheck/CI gate, Go/x-net bumps, replace/pins/go.work.sum, SSRF and DNS rebinding in fetch, Origin/CSRF, debug and pprof endpoints, TLS absence, security headers/CSP, error and log leakage, races, unsafe/CGO spots, image/drop/clipboard limits, crash dir, demo creds, IPC, template trust, Android SDK, npm overrides.
govulncheck (run 2026-10-22, root module, go1.26.10 toolchain): No vulnerabilities found. The examples module could not be scanned on its own (open DEP-4).

## A. Supply chain and dependencies
1. **Module integrity checks in CI.** Why: library consumers fetch via proxy/sumdb, and builds must fail if go.sum drifts. Action: run `go mod verify` and `GOFLAGS=-mod=readonly`; never set `GONOSUMDB`/`GOINSECURE`/`GOFLAGS=-insecure`; scope `GOPRIVATE` narrowly.
2. **Lock the toolchain.** Why: `GOTOOLCHAIN=auto` can silently download a different compiler. Action: set `GOTOOLCHAIN=local` in CI plus a `toolchain go1.26.x` line, and bump it on purpose.
3. **Retractions and immutable tags.** Why: tags are cached forever by proxy.golang.org, so a bad or vulnerable release can't be deleted. Action: use `retract` directives in go.mod, protect `v*` tags, and never re-tag.
4. **Audit the transitive parser surface.** Why: blinkless pulls goldmark, fpdf, star-tex, minify, brotli, and canvas, which are all parsers of complex input. Action: use `go mod why -m` and `go mod graph` reviews per bump, and a `depguard` rule so they never reach untrusted-input paths unintentionally.
5. **npm install-script hardening (frontend/).** Why: npm lifecycle scripts run arbitrary code on install. Action: `npm ci --ignore-scripts` in CI, `npm audit signatures`, lockfile-only installs.
6. **Gradle dependency verification (Android).** Why: Maven artifacts are trusted on TLS only. Action: `gradle --write-verification-metadata sha256`, then commit `gradle/verification-metadata.xml`.
7. **Dependency review on PRs.** Why: new deps and licenses land unnoticed. Action: `actions/dependency-review-action` (fail on high CVEs and disallowed licenses), plus `go-licenses check`.
8. **AI-assisted code guardrails.** Why: the README says code is AI-assisted, so there is a risk of hallucinated or typosquatted module names ("slopsquatting") and plausible-but-wrong security code. Action: every new module needs manual provenance review (pkg.go.dev, stars, owner), plus required tests for security-relevant changes.

## B. Build, release and CI
9. **Harden GitHub Actions when re-added.** Why: CI is a top supply-chain target. Action: pin actions by full commit SHA, set top-level `permissions: contents: read`, avoid `pull_request_target` with checkout, lint with `zizmor`/`actionlint`.
10. **Secret scanning.** Why: tokens (Apple/Authenticode/signing keys, API keys for examples) can leak into history. Action: `gitleaks` pre-commit and CI job, plus GitHub secret scanning push protection.
11. **Branch and tag protection.** Why: a single compromised session can push to master or retag. Action: protected `master` with required checks and review, signed commits/tags (SSH or gitsign), no force-push.
12. **Signed release artifacts and provenance.** Why: package.sh ships only SHA-256 sums, which detect corruption but not tampering. Action: `cosign sign-blob` (keyless) or `actions/attest-build-provenance` (SLSA L3) via goreleaser.
13. **SBOM per release.** Why: downstream users need to know what's inside the desktop/wasm/apk artifacts. Action: `cyclonedx-gomod app` or `syft` → attach SBOM to the GitHub release.
14. **Reproducible-build verification.** Why: proves that released binaries match the source. Action: build twice in CI with `-trimpath -buildvcs=true`, fixed `SOURCE_DATE_EPOCH`, and compare hashes.
15. **Compile out debug surfaces.** Why: runtime flags (DevTools, Debug, Perf, OWNFRAME_* env) can be flipped. Action: put devtools/pprof/debug handlers behind `//go:build ownframe_debug` so release binaries don't contain them.
16. **Binary hardening flags.** Why: the cgo/purego native code benefits from OS mitigations. Action: `-buildmode=pie`, `CGO_CFLAGS="-O2 -D_FORTIFY_SOURCE=2 -fstack-protector-strong"`, `CGO_LDFLAGS="-Wl,-z,relro,-z,now"`; verify with `checksec`.
17. **OS code signing in CI with protected keys.** Why: unsigned .app/.exe trains users to bypass Gatekeeper/SmartScreen. Action: codesign + notarytool + stapler, and signtool with keys in a cloud HSM/KMS or protected environment secrets, never on dev laptops.
18. **Taint-aware SAST beyond gosec.** Why: gosec is pattern-based, and taint tracking catches source→sink flows (fetch, exec, template). Action: CodeQL `security-extended` plus Semgrep custom rules (ban `template.HTML(`, `InsecureSkipVerify`, `http.DefaultServeMux`, `exec.Command("sh"`).
19. **Broaden golangci-lint.** Why: the current set misses resource/context bugs. Action: enable `bodyclose`, `noctx`, `contextcheck`, `errorlint`, `forcetypeassert`, `gochecknoglobals`, `depguard` (ban `unsafe`, `net/http/pprof`, `math/rand` outside allowlisted packages).

## C. Security process and API design
20. **SECURITY.md and private reporting.** Why: there is no disclosure channel, so researchers would open public issues. Action: add SECURITY.md (supported versions, SLA), enable GitHub Private Vulnerability Reporting, publish GHSAs so they flow into the Go vuln DB and consumers' govulncheck.
21. **Security regression tests.** Why: about 60 fixed findings can silently regress. Action: add one test per fixed ID (e.g. `TestSec_WEB1_MissingOriginRejected`), grouped under `-run Sec`, and make it required in CI.
22. **Secure-by-default API naming.** Why: library users copy examples, so risky options must look risky. Action: name opt-ins `InsecureAllowRemoteBind`, `UnsafeHTML`, etc.; zero values are always safe; deprecate footguns with `// Deprecated:`.
23. **Per-release threat-model refresh.** Why: new features (canvas, video, updater proposal, Android) change the trust boundaries. Action: add a release checklist item that updates documentation/security.md assumptions and data-flow diagram before tagging.

## D. Testing and verification
24. **Native fuzz targets (none exist today).** Why: HTML/CSS/template parsing, image/emoji decode, fetch URL/redirect checks, and click query parsing all take complex input. Action: add `func FuzzXxx` with seed corpora, run `go test -fuzz` for a time-box in CI, and use ClusterFuzzLite for continuous runs.
25. **Cross-arch and checkptr testing.** Why: 32-bit overflows and unsafe misuse only show on some GOARCH. Action: CI matrix with `GOARCH=386`/`arm`, `-gcflags=all=-d=checkptr`, and `go vet -unsafeptr` on linux/darwin/windows.
26. **Goroutine leak detection.** Why: the clipboard, fetch, ipc, and window loops spawn goroutines, and leaks become slow DoS in long-lived desktop apps. Action: `go.uber.org/goleak` `VerifyTestMain` in those packages.
27. **Algorithmic-complexity budgets.** Why: pathological CSS (huge selector lists, deep nesting, many keyframes) can freeze the UI loop. Action: benchmark tests with adversarial inputs plus hard time/alloc assertions (`testing.AllocsPerRun`, deadlines).
28. **Input length caps for form fields.** Why: typing or pasting megabytes into an input grows Page state and undo history without bound. Action: enforce `maxlength` (with a default cap) in FormValue/edit paths, and cap paste size.

## E. Input handling and rendering
29. **DOM depth and node-count limits.** Why: Go stack overflow (deep recursion in layout) is a fatal error that `recover()` can't catch. Action: cap nesting depth and total nodes at parse time; optionally use `debug.SetMaxStack` as a backstop.
30. **URL scheme allowlist for href/src.** Why: `file:`, `javascript:`, and oversized `data:` URIs in templates can read local files or exhaust memory. Action: allowlist `https`, `http`, and size-capped `data:image/*`; reject the rest at resolve time.
31. **Validate before handing URLs to the OS opener.** Why: `xdg-open`/`open`/`ShellExecute` on attacker URLs can launch arbitrary handlers. Action: only open http(s) and mailto, with explicit user action, never auto-open.
32. **Sandbox asset resolution.** Why: relative `<img>`/font/CSS paths resolved against the filesystem can escape the app dir. Action: resolve assets only through `fs.FS` (`embed.FS` or `os.Root`/`os.DirFS`) rooted at the app's asset dir.
33. **Untrusted font policy.** Why: the font parsers (freetype 2017, go-text) are old and complex, a classic memory/CPU attack surface. Action: load only bundled/system fonts by default; cap font file size and glyph count for any user-supplied font.
34. **Unicode spoofing in rendered text.** Why: bidi overrides and homoglyphs in notifications or chat-style examples can disguise content (Trojan Source style). Action: isolate or strip U+202A–U+202E and U+2066–U+2069 in untrusted text; use `golang.org/x/text/secure/precis` for identifiers.
35. **Template FuncMap discipline.** Why: one helper returning `template.HTML/JS/URL` from data defeats html/template escaping. Action: FuncMap funcs return plain strings only; any typed-safe value needs a reviewed constructor and a comment.

## F. HTTP and network
36. **Host header allowlist on Serve (local DNS rebinding).** Why: a malicious website can rebind its domain to 127.0.0.1 and drive the local preview; the Origin check alone misses same-origin rebinding. Action: reject requests whose `Host` isn't `localhost`, `127.0.0.1`, `[::1]`, or the configured name.
37. **Rate limiting on mutating routes.** Why: `/click`, `/type`, and the notify POST can be flooded to peg the render loop. Action: use a `golang.org/x/time/rate` limiter per remote addr, plus a max in-flight semaphore.
38. **Never use the default mux or client.** Why: importing `net/http/pprof` anywhere silently registers on `DefaultServeMux`, and `http.DefaultClient` has no timeout. Action: always construct `http.NewServeMux()` and a custom `http.Client`; enforce with depguard or Semgrep.
39. **Outbound client transport limits.** Why: fetch inherits `HTTP(S)_PROXY` env, so a hostile env can route all traffic through a proxy; slow servers can hang the app. Action: set `Proxy: nil` unless opted in, plus `Timeout`, `TLSHandshakeTimeout`, `ResponseHeaderTimeout`, `MaxResponseHeaderBytes`, and `MinVersion: tls.VersionTLS12`.
40. **Prefer Unix sockets or named pipes for local IPC.** Why: a loopback TCP port (desktop-cat notify) is reachable by every local user and can be port-squatted. Action: use a `0600` Unix socket in the user's runtime dir, or a named pipe with a user-only ACL; authenticate the server too.
41. **WebSocket hardening (lg-remote and similar).** Why: unbounded frames or no deadlines cause memory or goroutine exhaustion. Action: `SetReadLimit`, read/write deadlines, ping/pong timeout, and Origin check on the server side.

## G. Crypto and secrets
42. **No secrets in binaries or wasm.** Why: `-ldflags -X`, `//go:embed`, and the wasm bundle are trivially extractable, and the wasm build is fully public. Action: keep API keys server-side; scan artifacts with `strings`/gitleaks in the release job.
43. **OS keychain for app secrets.** Why: apps built on ownframe (teams/sqlite, download manager) will want tokens, and plaintext files leak. Action: give guidance or a helper using `github.com/zalando/go-keyring` (Keychain, Credential Manager, Secret Service).
44. **Redacting secret types.** Why: `%v` or slog on a struct can dump passwords. Action: wrap secrets in a type implementing `String()`, `GoString()`, `Format`, and `slog.LogValuer` that returns `[REDACTED]`.
45. **Signed auto-update guidance.** Why: the updater proposal leaves updates to apps, and a naive updater means remote code execution. Action: document a reference using TUF (`go-tuf`) or minisign/ed25519 verification, with version rollback protection and HTTPS only.

## H. Runtime and deployment
46. **Shared-library search-order hijacking.** Why: ebiten/purego and syscall code `dlopen`/`LoadLibrary` system libs by name, so a planted DLL or `LD_LIBRARY_PATH` wins. Action: use `windows.NewLazySystemDLL` / `LOAD_LIBRARY_SEARCH_SYSTEM32`, absolute paths on unix, and ignore the CWD.
47. **OS app sandboxing.** Why: limits the blast radius of any renderer bug. Action: macOS hardened runtime with minimal entitlements, a Flatpak manifest with narrow permissions, and a Windows `asInvoker` manifest; document them in packaging.md.
48. **Android manifest hardening.** Why: the Android demo ships an APK. Action: `android:exported` explicit, `allowBackup=false`, a network security config with `cleartextTrafficPermitted=false`, R8/minify on release, `FLAG_SECURE` on screens with sensitive data.
49. **Goroutine panic containment.** Why: a panic in any background goroutine (clipboard, notify, fetch) kills the whole app, bypassing the crash report. Action: launch internal goroutines through a `safeGo` helper that recovers, reports via the crash package, and logs.
50. **Overlay and click-through UI redress.** Why: transparent, always-on-top, click-through windows (desktop-cat) can cover other apps' prompts or capture clicks. Action: never accept input while click-through, show a visible focus/ownership indicator, and don't allow untrusted content to toggle transparency/click-through.
