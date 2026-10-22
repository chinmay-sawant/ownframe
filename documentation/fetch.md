# Fetch and XHR

`Fetch` and `XHR` use `net/http` for one request. They are not a browser network stack. The root file `fetch.go` forwards them to `internal/fetch`.

```go
res, err := ownframe.Fetch(ctx, "https://example.com/health")
if err != nil {
    return err
}
_, err = ownframe.XHR(ctx, "POST", "https://example.com/note",
    map[string]string{"Content-Type": "text/plain"},
    []byte("hi"),
)
```

`Fetch` always uses GET and sends no body. An empty method on `XHR` also means GET.

`FetchResponse` has `Status`, `Header`, and `Body`. `Header` is a flat map. Several values for one key are joined with `", "`. The body is read fully into memory with a 10 MiB cap (`MaxBodyBytes`). A larger body returns `ErrBodyTooLarge`. A 4xx or 5xx status is not an error; the caller reads `Status`.

The client rejects an empty URL and any scheme other than `http` or `https`, including `file`, `data`, and `javascript`. That error is `ErrScheme`. A malformed URL returns the `url.Parse` error instead. A redirect whose next URL is not `http` or `https` is also `ErrScheme`, and a redirect that steps from `https` down to `http` returns `ErrDowngrade`. The client follows at most nine redirects; a tenth returns an error whose text contains `stopped after 10 redirects`. Header-byte accounting is a byte-length approximation.

A nil context returns an error whose text contains `nil context`. Cancellation uses `http.NewRequestWithContext`.

The shared client has a 30 second default timeout, no cookie jar, and no cache. A context deadline shorter than 30 seconds still wins. It does not store a session. Pass a deadline on `ctx` to bound a request; a nil context returns an error whose text contains `nil context`.

The fetch example sends the body of a fetched PNG or JPEG to `Page.SetImage` with the source name `fetched`. The template body points `background-image` at that name, so a fetched image becomes the page background. See [screen.md](screen.md) for the paint path.

## Trust boundary

`Fetch` and `XHR` load whatever `http` or `https` URL the app gives them,
then reject hosts that resolve to loopback, private, link-local, multicast,
or unspecified addresses, including the `169.254.169.254` metadata range. A
hostname is resolved and every answer is checked, and a failed lookup fails
the request. Redirect targets are checked the same way. The client stores no
cookies and no cache, and it has a 30 second default timeout and a 10 MiB
response body cap. A small compressed response can still expand up to that
cap.

`AllowPrivate` returns a context that skips the address block for tests and
trusted local tooling. Scheme and userinfo checks still apply. Method and
header checks still apply: only `GET`, `POST`, `PUT`, `PATCH`, `DELETE`,
`HEAD`, and `OPTIONS` are allowed, and `Cookie`, `Authorization`, `Host`,
`Content-Length`, `Transfer-Encoding`, and `Connection` headers are blocked.

Recommended protections:

- Pass URLs that the app owns or has validated. Do not forward user controlled
  URLs to `Fetch` or `XHR`.
- Pass a context with a deadline. A nil context returns an error and
  `context.Background()` can hang forever.
- When any input is untrusted, add host allowlisting in the app. The client
  resolves each dial host once and dials a validated IP, rechecking every
  redirect hop, so a hostile resolver cannot rebind between check and
  connection for direct dials. Proxied requests still depend on the proxy's
  resolver.
- The client only sends allowlisted methods with blocked sensitive headers
  stripped by rejection. `Fetch` always uses `GET` with no body.

See the security audit in
[plans/v0.0.2/security](../plans/v0.0.2/security/00-overview-threat-model-and-priority.md).
