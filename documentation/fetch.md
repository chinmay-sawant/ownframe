# Fetch and XHR

`Fetch` and `XHR` use `net/http` for one request. They are not a browser network stack. The root file `fetch.go` forwards them to `internal/fetch`.

```go
res, err := gpui.Fetch(ctx, "https://example.com/health")
if err != nil {
    return err
}
_, err = gpui.XHR(ctx, "POST", "https://example.com/note",
    map[string]string{"Content-Type": "text/plain"},
    []byte("hi"),
)
```

`Fetch` always uses GET and sends no body. An empty method on `XHR` also means GET.

`FetchResponse` has `Status`, `Header`, and `Body`. `Header` is a flat map. Several values for one key are joined with `", "`. The body is read fully into memory with no size cap. A 4xx or 5xx status is not an error; the caller reads `Status`.

The client rejects an empty URL and any scheme other than `http` or `https`, including `file`, `data`, and `javascript`. That error is `ErrScheme`. A malformed URL returns the `url.Parse` error instead. A redirect whose next URL is not `http` or `https` is also `ErrScheme`. The client follows at most nine redirects; a tenth returns the error `stopped after 10 redirects`.

A nil context returns an error whose text contains `nil context`. Cancellation uses `http.NewRequestWithContext`.

The shared client has no cookie jar and no cache, and it sets no timeout. It does not store a session. Pass a deadline on `ctx` to bound a request.

The fetch example sends the body of a fetched PNG or JPEG to `Page.SetImage` with the source name `fetched`. The template body points `background-image` at that name, so a fetched image becomes the page background. See [screen.md](screen.md) for the paint path.
