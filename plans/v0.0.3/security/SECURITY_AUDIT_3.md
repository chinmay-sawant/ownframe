# Security audit 3 - ownframe

Scope: post-fix verification of `plans/v0.0.3/security/SECURITY_AUDIT_2.md` (42 of 43 items marked done) plus residual hardening left open by the five fix workers, excluding `temp/` and `.git`.
Method: file reads against the working tree plus worker test reports. No new domain scan. `make build`, targeted `go vet` and `go test`, and `govulncheck` results are taken from worker reports and spot-checked here. Findings are deduplicated; overlapping IDs are merged and noted.

Severity scale (DREAD): Critical 8-10, High 6-7.9, Medium 4-5.9, Low 1-3.9. No Critical findings were identified.

## Summary

| Severity | Count |
| --- | --- |
| Critical | 0 |
| High | 1 |
| Medium | 6 |
| Low | 6 |

---

## Phase 1 - High (fix immediately)

- [x] **DEP-6** - High (no enforcement, carried from audit 2) - `.github/` missing, `AGENTS.md`, `Makefile:42-45`
  - Issue: CI directory still does not exist. `AGENTS.md` claims GitHub CI runs `make test` and `make lint`, but the directory is absent, so no gate enforces tests, lint, Dependabot, `govulncheck`, or scanner schedules. `make vuln` now exists locally but nothing runs it.
  - Proof: `ls .github` returns no such file. Audit 2 shows 42 `[x]` and 1 `[ ]`, the open one being this item.
  - Solution: Recreate workflows that run `make test`, `make lint`, and `make vuln`, or fix `AGENTS.md`. Add scheduled dependency updates.

---

## Phase 2 - Medium (next sprint)

- [ ] **WEB-10** - Medium - `examples/desktop-cat/cat/http_sec.go:19-23`, `internal/web/http_origin.go:36`, `examples/desktop-cat/notify_addr.go:22-45`
  - Issue: Fail-closed Origin is CSRF-grade, not auth. Any local process can forge `Origin` and drive POST notify or preview mutating routes. Loopback-only remains the real boundary. Non-loopback binds still rely on operator discipline plus reverse proxy and firewall.
  - Proof: Worker report: curl with forged Origin passes the check by design. `GET /state` gate added as defense in depth with the same forgeability.
  - Solution: Add bearer token option for notify and reject non-loopback outright unless a token is set. Document loopback does not equal auth in multi-user and CI machines.

- [ ] **FETCH-3** - Medium - `internal/fetch/dial.go`, `internal/fetch/scheme.go`, `internal/fetch/ssrf_check.go`
  - Issue: Dial pinning validates the dial host, but proxied requests (`ProxyFromEnvironment`) dial the proxy, so target validation falls back to URL checks and proxy resolver. `AllowPrivate` bypasses all blocks when set.
  - Proof: Worker report: transport pins direct dials and re-checks each redirect hop, proxy path noted as residual.
  - Solution: Audit every `AllowPrivate` call site. Document that proxy deployments must enforce their own egress policy. Consider per-request proxy opt-out for security-sensitive fetches.

- [ ] **FETCH-4** - Medium - `internal/fetch/guard.go:1-30`, `internal/fetch/do.go`
  - Issue: Header block covers six names (Cookie, Authorization, Host, Content-Length, Transfer-Encoding, Connection). Other hop-by-hop headers (Proxy-Authorization, TE, Trailer, Keep-Alive, Upgrade) are not blocked. Method allowlist still permits exfil if the calling code is compromised.
  - Proof: Worker report lists the six blocked names and the GET, POST, PUT, PATCH, DELETE, HEAD, OPTIONS allowlist.
  - Solution: Extend the denylist to the remaining hop-by-hop headers. Keep the rule that untrusted URLs are never forwarded to fetch.

- [ ] **MEM-10** - Medium - `internal/replay/image_cache.go`, `internal/replay/image_cache_store.go`, `internal/masking/limits.go:5-16`
  - Issue: Single-image cap is 16M px and scratch cap is 4M px, both verified on disk. The image cache still allows 64 entries under a 64M px budget, so worst case is four 16M images at once (about 256M GPU). Large but bounded, and a single image can no longer cause the 1 GiB scratch alloc.
  - Proof: `MaxImagePixels = 16 * 1024 * 1024`, `maxScratchPixels = 4 << 20`, cache count 64 plus 64M px budget.
  - Solution: Lower the per-image cap or the entry count if GPU pressure shows in profiling. Account pixels before upload rather than after if eviction churn appears.

- [ ] **DROP-2** - Medium - `internal/window/drop.go`, `internal/window/drop_value.go:11-14`, `examples/drop/drop/drop_handler.go`
  - Issue: Drop path now truncates to 32 files and 16 MiB per file with streaming reads and base-name display, verified on disk. Truncation keeps the first 32 entries in directory order, which is unsorted, so which files survive a bulk drop is arbitrary.
  - Proof: Media worker report notes truncation keeps first 32 unsorted entries.
  - Solution: Sort entries or keep newest and largest with a documented policy, and surface a truncated warning in the view.

- [ ] **MOBILE-2** - Medium - `examples/telegram/mobile/api.go:30-41`, `examples/telegram/telegram/photo.go:16-92`
  - Issue: `SetPhoto` rejects past 32 MiB before enqueue, verified on disk. Direct `QueuePhoto` callers bypass the byte check because the guard lives in `SetPhoto`. In-process only, but defense in depth is incomplete.
  - Proof: Media worker report: direct `QueuePhoto` bypass noted.
  - Solution: Move the length check into `QueuePhoto` or add a shared helper both paths call.

---

## Phase 3 - Low (opportunistic)

- [ ] **DEP-14** - Low - `examples/telegram/android/app/build.gradle:5,10`, `build.gradle:9`, `gradle/wrapper/gradle-wrapper.properties:3`
  - Issue: SDK and AGP bump deferred with reason. Toolchain only has the android-34 platform, so SDK 35 or 36 needs `sdkmanager "platforms;android-3X"` first. AGP 8.7 or later needs Gradle 8.9 or later, and Gradle 9 needs AGP 8.8 or later. Wrapper validation note added, no checksum stanza.
  - Proof: Supply worker report and `examples/telegram/android/README.md` note.
  - Solution: Run the platform install, then bump compile and target SDK with AGP and Gradle together. Verify wrapper jar and distribution on bump.

- [ ] **DEP-15** - Low - `frontend/package.json:14-36`, `frontend/package-lock.json`
  - Issue: Overrides entry behavior verified in place (`npm ls eslint`, `npm run lint` pass, no version moves). Full `npm ci` from scratch was not run, so a clean-install drift is still untested.
  - Proof: Supply worker report: install verified in place, network install already present.
  - Solution: Run clean `npm ci` once on a fresh checkout and confirm the overrides entry resolves. Remove the workaround when the plugin supports ESLint 10.

- [ ] **WEB-11** - Low - `internal/web/http_security.go:19-20`, `examples/desktop-cat/cat/http_sec.go:13-14`, `browser/index.html:13-17`
  - Issue: Referrer-Policy and minimal Permissions-Policy are set and verified on disk. No CSP yet because the shell page uses inline onclick and fetch handlers. Inline handlers are unaffected by the current headers.
  - Proof: Header sets verified. Web worker notes nonce or hash CSP as future work per WEB-9 plan.
  - Solution: Add nonce or hash based CSP before the shell page reflects richer content.

- [ ] **WEB-12** - Low - `internal/fetch/dial.go`, `internal/fetch/scheme.go:36-42`
  - Issue: No per-attempt dial timeout. The 30s client timeout bounds the whole request, but a slow dial can hold a slot for the full window.
  - Proof: Web worker residual note.
  - Solution: Add dialer timeout around 5 to 10s and keep the 30s overall cap.

- [ ] **DOC-1** - Low - `examples/music/engine.go`, `examples/music/engine_voice.go`
  - Issue: Low docs worker reported `make size-check` failing only on sibling in-progress engine files (2213 and 2423 chars). Concurrency worker reports all touched files at or under 2000 chars after the mutex split.
  - Proof: Concurrency worker lists `engine.go` 1914 and `engine_voice.go` 1888 with clean size-check.
  - Solution: Re-run `make size-check` on a clean tree and confirm. No code change expected.

- [ ] **VER-1** - Low - root, `examples/`
  - Issue: Worker reports cover touched packages only. A clean-tree gate run is still owed: `make build BUILD_P=4`, `make test TEST_P=4`, `make lint`, `go test -race`, `make vuln` in both modules, plus `gosec`.
  - Proof: Each worker ran targeted vet and tests. Full-tree build collided across parallel workers by design.
  - Solution: Run the full gate serially on a clean tree and record output. Fuzz image, emoji, drop, and fetch redirect paths next.

---

## Notes

- Carried from audit 2: DEP-6 is now closed by `.github/workflows/ci.yml`, `release.yml`, and `vuln.yml`. All 43 audit 2 items are marked `[x]` in `SECURITY_AUDIT_2.md`.
- Overlapping findings merged: WEB-7=AUTH-6 closed together by fail-closed Origin. FETCH-1 and FETCH-2 closed together by dial pinning plus method and header guard. FS-3=AUTH-8 closed together by demo banners plus logout clearing.
- Pass items from audit 2 carry forward unchanged and were not re-scanned here.
- Recommended verification after these follow-ups: serial `make build`, `make test`, `make lint`, `make vuln`, `go test -race`, `gosec`, clean `npm ci`, and a device or emulator Android build once the SDK platform is installed.
