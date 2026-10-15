# Web server, network exposure, and fetch risks

## Serve web mode assumes a trusted local browser

Location: `internal/web`, root `serve.go`

`Serve` exposes a local web preview. Routes include `/`, `/frame.png`, `/click`, `/type`, `/backspace`, `/debug/state`, and conditionally `/pdf`. `ServeWithOptions{Perf: true}` adds `/debug/pprof/*`. There is no authentication, no TLS, no CSRF protection, and no request timeout.

### Vulnerabilities

- Severity: High
  - No `http.Server` timeouts. `ServeWithOptions` calls `net.Listen` and `http.Serve(ln, mux)` directly. A slow client can hold connections open and a large request can be buffered fully.
- Severity: High
  - `/click` is a GET with side effects. It mutates page state and can run click handlers. Cross site GET requests can target it from an `<img>` tag even when the browser is visiting another origin.
- Severity: High
  - The bind address is caller controlled. `Serve` refuses nothing. Any reachable listener gives full page control, including app handlers.
- Severity: Medium
  - `/debug/state` is registered for plain `Serve`. It returns boxes, dimensions, stats, text, and operation counts. It is not gated by `Perf`.
- Severity: Medium
  - `/type` calls `r.ParseForm()` with no `MaxBytesReader`, allowing memory growth from large requests.
- Severity: Medium
  - `click` accepts `NaN` and `Inf` float coordinates without rejecting them.
- Severity: Medium
  - `Serve` does not recover panics. A panic in a handler or PDF render tears down the process.
- Severity: Low
  - Responses lack `X-Content-Type-Options` and framing headers.
  - The `/pdf` type assertion does not use the `ok` form.
- Severity: Low
  - Debug environment variables and local devtools can disclose paths, commands, and metrics.

### Valid scenario

An app calls `Serve("127.0.0.1:8080")` on the developer machine. The browser preview is reached only by the same user or machine.

### Invalid scenario

An app calls `Serve("0.0.0.0:8080")` on a shared machine. A remote peer can open the preview and click controls through `/click`.

### Recommended fixes

1. Use `http.Server` with `ReadHeaderTimeout`, `ReadTimeout`, `WriteTimeout`, and `IdleTimeout`.
2. Add `http.MaxBytesReader` to `/type` and reject non finite coordinates.
3. Change `/click` from GET to POST, or add an origin and CSRF check.
4. Gate `/debug/state` behind the same debugging option used for `/debug/pprof`.
5. Make the `/pdf` type assertion safe.
6. Add nosniff and anti framing headers.
7. Document that non loopback binds require the caller to add authentication.
8. Bind browser helper scripts to localhost explicitly.

## Fetch and XHR

Location: `internal/fetch`

`Fetch` and `XHR` accept any `http` or `https` URL. They do not use a cookie jar and do not upgrade redirects to web schemes from non web schemes. No host or IP allowlist exists. That model is acceptable when callers supply trusted endpoints, but it leaves SSRF and availability concerns when apps forward untrusted URLs.

### Vulnerabilities

- Severity: High
  - No request timeout by default. `client` has no `Timeout`. A caller passing `context.Background()` can hang forever.
- Severity: High
  - Unbounded response body. `io.ReadAll(resp.Body)` allows decompression and memory exhaustion from a small compressed response.
- Severity: Medium
  - No SSRF protection. Loopback, private, link local, and metadata addresses are reachable.
- Severity: Medium
  - DNS rebinding is not mitigated because the host is resolved by the transport after validation.
- Severity: Medium
  - Environment proxy variables can redirect requests and bypass local filtering.
  - HTTPS can redirect to HTTP.
- Severity: Low
  - Arbitrary methods accepted, including `TRACE` and `CONNECT`.
  - Custom headers and `Host` values are forwarded on same domain redirects.
  - TLS, header joining, and method handling lack extra hardening.

### Valid scenario

An app calls `Fetch` with an endpoint it owns, while passing a context deadline.

### Invalid scenario

An app passes a user controlled URL shaped like `http://169.254.169.254/...` to `Fetch`. The library fetches cloud metadata unless the app validates the URL first.

### Recommended fixes

1. Add a default transport with a timeout and still respect `ctx`.
2. Cap response bodies with `io.LimitReader`, then reject oversized responses.
3. Add resolver and IP checks if the library is expected to fetch untrusted URLs.
4. Pin DNS results to prevent rebinding.
5. Document proxy behavior and downgrade behavior.
6. Consider a method allowlist unless generic proxy semantics are intended.
7. Add tests for timeouts, body limits, and metadata endpoint rejection.